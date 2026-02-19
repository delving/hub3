package v2adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/domain/semantic"
	elastic "github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"
)

// V2IntrospectionAdapter implements semantic.IntrospectionStore using ES aggregations.
type V2IntrospectionAdapter struct {
	client *elastic.Client
	index  string
	log    zerolog.Logger
}

// NewV2IntrospectionAdapter creates a new introspection adapter.
func NewV2IntrospectionAdapter(client *elastic.Client, index string, logger zerolog.Logger) *V2IntrospectionAdapter {
	return &V2IntrospectionAdapter{
		client: client,
		index:  index,
		log:    logger.With().Str("component", "v2_introspect").Logger(),
	}
}

// buildClassAggregation creates a nested aggregation to discover RDF class types
// from the resources.types field.
func buildClassAggregation() *elastic.NestedAggregation {
	return elastic.NewNestedAggregation().Path("resources").
		SubAggregation("class_types",
			elastic.NewTermsAggregation().Field("resources.types").Size(100),
		)
}

// buildPropertyAggregation creates a nested aggregation to discover properties
// for a given class. If classFilter is non-empty, properties are scoped to that class.
func buildPropertyAggregation(classFilter string) *elastic.NestedAggregation {
	agg := elastic.NewNestedAggregation().Path("resources")

	entriesAgg := elastic.NewNestedAggregation().Path("resources.entries").
		SubAggregation("predicates",
			elastic.NewTermsAggregation().Field("resources.entries.predicate").Size(200),
		)

	if classFilter != "" {
		filterAgg := elastic.NewFilterAggregation().
			Filter(elastic.NewTermQuery("resources.types", classFilter)).
			SubAggregation("entries", entriesAgg)
		agg.SubAggregation("class_filter", filterAgg)
	} else {
		agg.SubAggregation("entries", entriesAgg)
	}

	return agg
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
// Both slash-delimited and hash-delimited namespace URIs are handled because
// the commonNamespaces map includes the trailing separator character.
func classInfoFromBucket(uri string, count int64) semantic.ClassInfo {
	label := uri

	// Try to match against known namespaces
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

// IntrospectClasses discovers all RDF class types in the index using a nested terms aggregation.
func (a *V2IntrospectionAdapter) IntrospectClasses(ctx context.Context, opts *semantic.QueryOptions) ([]semantic.ClassInfo, error) {
	search := a.client.Search(a.index).Size(0)

	if opts != nil && opts.Query != nil && opts.Query.Value != "" {
		search = search.Query(elastic.NewQueryStringQuery(opts.Query.Value))
	}

	search = search.Aggregation("classes", buildClassAggregation())

	result, err := search.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("introspect classes: %w", err)
	}

	nested, found := result.Aggregations.Nested("classes")
	if !found {
		return nil, nil
	}

	terms, found := nested.Terms("class_types")
	if !found {
		return nil, nil
	}

	classes := make([]semantic.ClassInfo, 0, len(terms.Buckets))
	for _, bucket := range terms.Buckets {
		uri, ok := bucket.Key.(string)
		if !ok {
			continue
		}

		classes = append(classes, classInfoFromBucket(uri, bucket.DocCount))
	}

	return classes, nil
}

// IntrospectProperties discovers properties for a given class using nested aggregations on entries.
func (a *V2IntrospectionAdapter) IntrospectProperties(ctx context.Context, classURI string, opts *semantic.QueryOptions) ([]semantic.PropertyInfo, error) {
	search := a.client.Search(a.index).Size(0)
	search = search.Aggregation("properties", buildPropertyAggregation(classURI))

	result, err := search.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("introspect properties: %w", err)
	}

	_ = result

	return nil, nil // TODO: implement result parsing when ES is available for integration testing
}

// IntrospectField returns value distribution for a specific field.
func (a *V2IntrospectionAdapter) IntrospectField(ctx context.Context, field string, opts *semantic.QueryOptions) (*semantic.PropertyInfo, error) {
	return nil, nil // TODO: implement
}

// IntrospectPaths returns predicate paths between classes.
func (a *V2IntrospectionAdapter) IntrospectPaths(ctx context.Context, opts *semantic.QueryOptions) ([]semantic.PathInfo, error) {
	return nil, nil // TODO: implement
}
