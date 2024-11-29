package pid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/html"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/olivere/elastic/v7"
)

func (s *Service) handleArk() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/ark:") {
			s.renderError(w, r, fmt.Errorf("unknown ark: %s", r.URL.Path), http.StatusNotFound)
			return
		}

		ark := validateARK(r.URL.Path)
		if !strings.HasSuffix(r.URL.Path, ark) {
			http.Redirect(w, r, "/"+ark, http.StatusFound)
			return
		}

		store, err := s.GetStore(domain.GetOrganizationID(r).String())
		if err != nil {
			s.renderError(w, r, err, http.StatusBadRequest)
			return
		}

		resp, err := store.Get(ark, "")
		if err != nil {
			switch {
			case errors.Is(err, ErrTombstone):
				s.renderError(w, r, err, http.StatusGone)
				return
			case errors.Is(err, ErrPIDNotFound):
				if resp.Empty() {
					s.renderError(w, r, err, http.StatusNotFound)
					return
				}
			default:
				s.renderError(w, r, err, http.StatusBadRequest)
				return
			}
		}

		if strings.EqualFold(r.URL.Query().Get("format"), "json") {
			render.JSON(w, r, resp)
			return
		}

		if strings.EqualFold(r.URL.Query().Get("format"), "rdf") {
			resp.CurrentPID.IsShownAt = ""
		}

		if resp.Empty() {
			slog.Error("unable to find pid", "error", err, "pid", ark)
			for naan, endpoint := range s.fallbackNaan {
				slog.Info("checking fallback", "naan", naan, "endpoint", endpoint)
				if strings.Contains(ark, "/"+naan+"/") {
					resp = StoreResponse{
						CurrentPID: &PID{
							ID:         "",
							ExternalID: ark,
							Type:       Ark,
							IsShownAt:  "",
						},
						OtherIDs: []*PID{},
						Fallback: endpoint,
					}
				}
			}
			if resp.Empty() {
				s.renderError(w, r, ErrPIDNotFound, http.StatusNotFound)
				return
			}
		}

		if resp.Tombstone() {
			s.renderError(w, r, ErrTombstone, http.StatusGone)
			return
		}

		if resp.CurrentPID.IsShownAt != "" {
			http.Redirect(w, r, resp.CurrentPID.IsShownAt, http.StatusFound)
			return
		}

		var g *rdf.Graph
		switch {
		case resp.Fallback != "":
			subj := "https://n2t.net/" + resp.CurrentPID.ExternalID
			b, err := s.SparqlQuery(r.Context(), resp.Fallback, subj)
			if err != nil {
				s.renderError(w, r, err, http.StatusBadRequest)
				return
			}

			g, err = ntriples.Parse(bytes.NewReader(b), nil)
			if err != nil {
				slog.Error("unable to parse graph; %w", "error", err)
				s.renderError(w, r, err, http.StatusBadRequest)
				return
			}

			s, _ := rdf.NewIRI(subj)
			g.Subject = rdf.Subject(s)
		default:
			g, err = s.GetGraph(r.Context(), resp.CurrentPID.Meta.HubID)
			if err != nil {
				s.renderError(w, r, err, http.StatusBadRequest)
				return
			}
		}

		err = html.RenderRDF(w, r, g)
		if err != nil {
			s.renderError(w, r, err, http.StatusBadRequest)
			return
		}
	}
}

func (s *Service) GetResource(ctx context.Context, uri string) (*rdf.Resource, error) {
	g, err := s.GetGraph(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get graph: %w", err)
	}
	subj, _ := rdf.NewIRI(uri)
	rsc, ok := g.Get(rdf.Subject(subj))
	if !ok {
		return nil, fmt.Errorf("unable to get resource: %s", uri)
	}
	return rsc, nil
}

// SparqlQuery executes a SPARQL query against the provided endpoint and URI
func (s *Service) SparqlQuery(ctx context.Context, endpoint, uri string) ([]byte, error) {
	// Base URL
	baseURL := "https://data.brabantcloud.nl/k3/query/"

	// URL encode the provided URI
	encodedURI := url.QueryEscape(fmt.Sprintf("describe <%s>", uri))

	// Construct the full URL
	fullURL := fmt.Sprintf("%s?query=%s", baseURL, encodedURI)

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response status code is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

func (s *Service) GetGraph(ctx context.Context, uri string) (*rdf.Graph, error) {
	fg, err := s.getSearchRecord(ctx, uri)
	if err != nil {
		for _, endpoint := range s.fallbackTripleStores {
			b, queryErr := s.SparqlQuery(ctx, endpoint, uri)
			if queryErr != nil {
				slog.Error("unable to get search record; %w", "error", queryErr, "uri", uri)
				continue
			}
			if len(b) == 0 {
				continue
			}

			g, parseErr := ntriples.Parse(bytes.NewReader(b), nil)
			if parseErr != nil {
				slog.Error("unable to parse graph; %w", "error", parseErr)
				continue
			}

			return g, nil
		}
		return nil, fmt.Errorf("unable to get search record: %w", err)
	}

	return fg.Graph()
}

func buildQuery(id string) (*elastic.BoolQuery, error) {
	// Create the main bool query
	mainQuery := elastic.NewBoolQuery()

	// Create the inner bool query with should clauses
	innerBoolQuery := elastic.NewBoolQuery()

	// Create the term query for meta.entryURI with higher boost
	metaEntryURIQuery := elastic.NewTermQuery("meta.entryURI", id).Boost(2.0)

	// Create the nested query for resources.id with lower boost
	resourcesQuery := elastic.NewNestedQuery(
		"resources",
		elastic.NewTermQuery("resources.id", id),
	).Boost(1.0)

	// Add both queries as should clauses to the inner bool query
	innerBoolQuery = innerBoolQuery.Should(metaEntryURIQuery, resourcesQuery)

	// Set minimum_should_match to 1 to ensure at least one of these conditions must match
	innerBoolQuery = innerBoolQuery.MinimumShouldMatch("1")

	// Add the inner bool query as a must clause to the main bool query
	mainQuery = mainQuery.Must(innerBoolQuery)

	return mainQuery, nil
}

func (s *Service) getSearchRecord(ctx context.Context, id string) (*fragments.FragmentGraph, error) {
	unexpectedResponseMsg := "expected response != nil; got: %v"

	org, ok := domain.GetOrganizationFromCtx(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to get organization from context")
	}

	q := elastic.NewBoolQuery()
	switch {
	case strings.HasPrefix(id, "urn:") && strings.HasSuffix(id, "/graph"):
		q = q.Must(elastic.NewTermQuery("meta.namedGraphURI", id))
	case strings.HasPrefix(id, "http"), strings.HasPrefix(id, "urn:"):
		boolQuery, err := buildQuery(id)
		if err != nil {
			return nil, fmt.Errorf("unable to build query: %w", err)
		}
		q = q.Must(boolQuery)
	default:
		q = q.Must(elastic.NewTermQuery("meta.hubID", id))
	}

	res, err := index.ESClient().Search().
		Index(org.Config.GetIndexName()).
		Query(q).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf(unexpectedResponseMsg, res)
	}

	if res.Hits.TotalHits.Value == 0 {
		return nil, fmt.Errorf("%s was not found", domain.LogUserInput(id))
	}

	record := res.Hits.Hits[0].Source

	return decodeFragmentGraph(record)
}

func decodeFragmentGraph(hit json.RawMessage) (*fragments.FragmentGraph, error) {
	r := new(fragments.FragmentGraph)
	if err := json.Unmarshal(hit, r); err != nil {
		return nil, err
	}
	return r, nil
}
