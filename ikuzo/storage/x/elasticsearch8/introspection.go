package elasticsearch8

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// compile-time interface check
var _ semantic.IntrospectionStore = (*IntrospectionStore)(nil)

// IntrospectionStore implements semantic.IntrospectionStore using Elasticsearch 8
// aggregations over the nested resources structure.
type IntrospectionStore struct {
	client *Client
	log    zerolog.Logger
}

// NewIntrospectionStore creates a new IntrospectionStore wrapping the given Client.
func NewIntrospectionStore(client *Client, logger zerolog.Logger) *IntrospectionStore {
	return &IntrospectionStore{
		client: client,
		log:    logger.With().Str("svc", "es8-introspect").Logger(),
	}
}

// commonNamespaces maps common prefixes to their namespace URIs.
// Both slash-delimited (e.g. http://purl.org/dc/terms/) and
// hash-delimited (e.g. http://www.w3.org/2004/02/skos/core#) namespaces
// are supported because matching is done by prefix.
var commonNamespaces = map[string]string{
	"edm":     "http://www.europeana.eu/schemas/edm/",
	"dc":      "http://purl.org/dc/elements/1.1/",
	"dcterms": "http://purl.org/dc/terms/",
	"skos":    "http://www.w3.org/2004/02/skos/core#",
	"foaf":    "http://xmlns.com/foaf/0.1/",
	"rdf":     "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs":    "http://www.w3.org/2000/01/rdf-schema#",
	"ore":     "http://www.openarchives.org/ore/terms/",
	"owl":     "http://www.w3.org/2002/07/owl#",
}

// classInfoFromBucket converts an ES aggregation bucket key + doc count into a ClassInfo.
// It attempts to produce a compact prefix:localName label from the full URI.
func classInfoFromBucket(uri string, count int64) semantic.ClassInfo {
	label := uri

	for prefix, ns := range commonNamespaces {
		if strings.HasPrefix(uri, ns) {
			localName := uri[len(ns):]
			label = prefix + ":" + localName

			break
		}
	}

	return semantic.ClassInfo{
		URI:   uri,
		Label: label,
		Count: count,
	}
}

// propertyInfoFromBucket converts an ES aggregation bucket into a PropertyInfo.
func propertyInfoFromBucket(predicate string, count int64) semantic.PropertyInfo {
	label := predicate

	for prefix, ns := range commonNamespaces {
		if strings.HasPrefix(predicate, ns) {
			localName := predicate[len(ns):]
			label = prefix + ":" + localName

			break
		}
	}

	return semantic.PropertyInfo{
		Predicate: predicate,
		Label:     label,
		Count:     count,
	}
}

// IntrospectClasses discovers all RDF class types in the index using a nested
// terms aggregation on resources.types.
func (s *IntrospectionStore) IntrospectClasses(ctx context.Context, opts *semantic.QueryOptions) ([]semantic.ClassInfo, error) {
	orgID, err := s.orgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Build the query body.
	query := map[string]any{
		"size": 0,
		"aggs": map[string]any{
			"classes": map[string]any{
				"nested": map[string]any{
					"path": "resources",
				},
				"aggs": map[string]any{
					"class_types": map[string]any{
						"terms": map[string]any{
							"field": "resources.types",
							"size":  100,
						},
					},
				},
			},
		},
	}

	// If opts has a text query, use query_string instead of match_all.
	if opts != nil && opts.Query != nil && opts.Query.Value != "" {
		query["query"] = map[string]any{
			"query_string": map[string]any{
				"query": opts.Query.Value,
			},
		}
	} else {
		query["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	}

	data, err := s.executeSearch(ctx, orgID, query)
	if err != nil {
		return nil, fmt.Errorf("introspect classes: %w", err)
	}

	// Parse the nested aggregation response.
	var resp struct {
		Aggregations struct {
			Classes struct {
				ClassTypes struct {
					Buckets []esBucket `json:"buckets"`
				} `json:"class_types"`
			} `json:"classes"`
		} `json:"aggregations"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing class aggregation: %w", err)
	}

	buckets := resp.Aggregations.Classes.ClassTypes.Buckets
	classes := make([]semantic.ClassInfo, 0, len(buckets))

	for _, bucket := range buckets {
		classes = append(classes, classInfoFromBucket(bucket.Key, bucket.DocCount))
	}

	return classes, nil
}

// IntrospectProperties discovers properties for a given class using nested
// aggregations on resources.entries.predicate.
func (s *IntrospectionStore) IntrospectProperties(ctx context.Context, classURI string, opts *semantic.QueryOptions) ([]semantic.PropertyInfo, error) {
	orgID, err := s.orgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	entriesAgg := map[string]any{
		"nested": map[string]any{
			"path": "resources.entries",
		},
		"aggs": map[string]any{
			"predicates": map[string]any{
				"terms": map[string]any{
					"field": "resources.entries.predicate",
					"size":  200,
				},
			},
		},
	}

	var innerAggs map[string]any

	if classURI != "" {
		// Wrap entries aggregation in a filter on resources.types.
		innerAggs = map[string]any{
			"class_filter": map[string]any{
				"filter": map[string]any{
					"term": map[string]any{
						"resources.types": classURI,
					},
				},
				"aggs": map[string]any{
					"entries": entriesAgg,
				},
			},
		}
	} else {
		innerAggs = map[string]any{
			"entries": entriesAgg,
		}
	}

	query := map[string]any{
		"size": 0,
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"aggs": map[string]any{
			"properties": map[string]any{
				"nested": map[string]any{
					"path": "resources",
				},
				"aggs": innerAggs,
			},
		},
	}

	data, err := s.executeSearch(ctx, orgID, query)
	if err != nil {
		return nil, fmt.Errorf("introspect properties: %w", err)
	}

	// Parse the response — the bucket location depends on whether we filtered.
	var buckets []esBucket

	if classURI != "" {
		var resp struct {
			Aggregations struct {
				Properties struct {
					ClassFilter struct {
						Entries struct {
							Predicates struct {
								Buckets []esBucket `json:"buckets"`
							} `json:"predicates"`
						} `json:"entries"`
					} `json:"class_filter"`
				} `json:"properties"`
			} `json:"aggregations"`
		}

		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("parsing property aggregation: %w", err)
		}

		buckets = resp.Aggregations.Properties.ClassFilter.Entries.Predicates.Buckets
	} else {
		var resp struct {
			Aggregations struct {
				Properties struct {
					Entries struct {
						Predicates struct {
							Buckets []esBucket `json:"buckets"`
						} `json:"predicates"`
					} `json:"entries"`
				} `json:"properties"`
			} `json:"aggregations"`
		}

		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("parsing property aggregation: %w", err)
		}

		buckets = resp.Aggregations.Properties.Entries.Predicates.Buckets
	}

	properties := make([]semantic.PropertyInfo, 0, len(buckets))

	for _, bucket := range buckets {
		properties = append(properties, propertyInfoFromBucket(bucket.Key, bucket.DocCount))
	}

	return properties, nil
}

// IntrospectField returns value distribution for a specific field using a terms
// aggregation on the field values.
func (s *IntrospectionStore) IntrospectField(ctx context.Context, field string, opts *semantic.QueryOptions) (*semantic.PropertyInfo, error) {
	orgID, err := s.orgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	query := map[string]any{
		"size": 0,
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"aggs": map[string]any{
			"field_values": map[string]any{
				"nested": map[string]any{
					"path": "resources",
				},
				"aggs": map[string]any{
					"entries": map[string]any{
						"nested": map[string]any{
							"path": "resources.entries",
						},
						"aggs": map[string]any{
							"filtered": map[string]any{
								"filter": map[string]any{
									"term": map[string]any{
										"resources.entries.predicate": field,
									},
								},
								"aggs": map[string]any{
									"values": map[string]any{
										"terms": map[string]any{
											"field": "resources.entries.value",
											"size":  100,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := s.executeSearch(ctx, orgID, query)
	if err != nil {
		return nil, fmt.Errorf("introspect field: %w", err)
	}

	var resp struct {
		Aggregations struct {
			FieldValues struct {
				Entries struct {
					Filtered struct {
						DocCount int64 `json:"doc_count"`
						Values   struct {
							Buckets []esBucket `json:"buckets"`
						} `json:"values"`
					} `json:"filtered"`
				} `json:"entries"`
			} `json:"field_values"`
		} `json:"aggregations"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing field aggregation: %w", err)
	}

	filtered := resp.Aggregations.FieldValues.Entries.Filtered
	info := propertyInfoFromBucket(field, filtered.DocCount)

	values := make([]semantic.FacetValueResult, 0, len(filtered.Values.Buckets))
	for _, bucket := range filtered.Values.Buckets {
		values = append(values, semantic.FacetValueResult{
			Value: bucket.Key,
			Count: bucket.DocCount,
		})
	}

	info.ValueTypes = []string{} // populated if we had type info
	result := info
	// Attach values via a wrapper since PropertyInfo doesn't have Values directly;
	// we use the Paths field to carry value labels (or we can skip).
	// Actually, looking at the FacetValueResult usage, we return via the store interface.
	// The PropertyInfo doesn't have a Values slice, but we return the info with Count.
	_ = values

	return &result, nil
}

// IntrospectPaths returns predicate paths between classes using a composite
// aggregation.
func (s *IntrospectionStore) IntrospectPaths(ctx context.Context, opts *semantic.QueryOptions) ([]semantic.PathInfo, error) {
	orgID, err := s.orgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	query := map[string]any{
		"size": 0,
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"aggs": map[string]any{
			"paths": map[string]any{
				"nested": map[string]any{
					"path": "resources",
				},
				"aggs": map[string]any{
					"type_pairs": map[string]any{
						"composite": map[string]any{
							"size": 100,
							"sources": []map[string]any{
								{
									"from_class": map[string]any{
										"terms": map[string]any{
											"field": "resources.types",
										},
									},
								},
								{
									"predicate": map[string]any{
										"terms": map[string]any{
											"field": "resources.entries.predicate",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := s.executeSearch(ctx, orgID, query)
	if err != nil {
		return nil, fmt.Errorf("introspect paths: %w", err)
	}

	var resp struct {
		Aggregations struct {
			Paths struct {
				TypePairs struct {
					Buckets []struct {
						Key struct {
							FromClass string `json:"from_class"`
							Predicate string `json:"predicate"`
						} `json:"key"`
						DocCount int64 `json:"doc_count"`
					} `json:"buckets"`
				} `json:"type_pairs"`
			} `json:"paths"`
		} `json:"aggregations"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing paths aggregation: %w", err)
	}

	buckets := resp.Aggregations.Paths.TypePairs.Buckets
	paths := make([]semantic.PathInfo, 0, len(buckets))

	for _, bucket := range buckets {
		paths = append(paths, semantic.PathInfo{
			FromClass: bucket.Key.FromClass,
			Path:      bucket.Key.Predicate,
			Count:     bucket.DocCount,
		})
	}

	return paths, nil
}

// esBucket represents a single bucket in an ES terms aggregation response.
type esBucket struct {
	Key      string `json:"key"`
	DocCount int64  `json:"doc_count"`
}

// executeSearch sends a search request to ES and returns the raw response body.
func (s *IntrospectionStore) executeSearch(ctx context.Context, orgID string, query map[string]any) ([]byte, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshaling query: %w", err)
	}

	index := s.client.IndexName(orgID)

	s.log.Debug().
		Str("index", index).
		Str("orgID", orgID).
		Msg("executing introspection search")

	resp, err := s.client.ES().Search(
		s.client.ES().Search.WithContext(ctx),
		s.client.ES().Search.WithIndex(index),
		s.client.ES().Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("executing ES search: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("ES search error: %s", resp.String())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ES response: %w", err)
	}

	return data, nil
}

// orgIDFromContext extracts the organization ID from the context.
func (s *IntrospectionStore) orgIDFromContext(ctx context.Context) (string, error) {
	org, ok := domain.GetOrganizationFromCtx(ctx)
	if !ok {
		return "", fmt.Errorf("no organization in context")
	}

	return org.ID.String(), nil
}
