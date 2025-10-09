package v2adapter

import (
	"testing"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestResultTranslator_TranslateSearchResult(t *testing.T) {
	rt := NewResultTranslator()

	tests := []struct {
		name         string
		scrollResult *fragments.ScrollResultV4
		opts         *semantic.QueryOptions
		wantTotal    int64
		wantItems    int
		wantFacets   int
		wantErr      bool
		description  string
	}{
		{
			name: "basic result with semantic items",
			scrollResult: &fragments.ScrollResultV4{
				Query: &fragments.Query{
					Numfound: 10,
					Terms:    "test query",
				},
				Items: []*fragments.FragmentGraph{
					{
						Meta: &fragments.Header{
							HubID: "item1",
						},
						Semantic: &fragments.SemanticView{
							Context: "/contexts/edm/latest/context.jsonld",
							ID:      "http://example.org/item1",
							Type:    []string{"ProvidedCHO"},
							Fields: map[string]fragments.SemanticValue{
								"title": &fragments.SemanticLiteral{Value: "Test Item 1"},
							},
						},
					},
					{
						Meta: &fragments.Header{
							HubID: "item2",
						},
						Semantic: &fragments.SemanticView{
							Context: "/contexts/edm/latest/context.jsonld",
							ID:      "http://example.org/item2",
							Type:    []string{"ProvidedCHO"},
							Fields: map[string]fragments.SemanticValue{
								"title": &fragments.SemanticLiteral{Value: "Test Item 2"},
							},
						},
					},
				},
			},
			opts:        &semantic.QueryOptions{},
			wantTotal:   10,
			wantItems:   2,
			wantFacets:  0,
			wantErr:     false,
			description: "Should extract semantic items correctly",
		},
		{
			name: "result with facets",
			scrollResult: &fragments.ScrollResultV4{
				Query: &fragments.Query{
					Numfound: 5,
				},
				Items: []*fragments.FragmentGraph{
					{
						Meta: &fragments.Header{
							HubID: "item1",
						},
						Semantic: &fragments.SemanticView{
							ID:   "http://example.org/item1",
							Type: []string{"ProvidedCHO"},
						},
					},
				},
				Facets: []*fragments.QueryFacet{
					{
						Field:       "creator",
						Name:        "Creator",
						Total:       100,
						MissingDocs: 5,
						OtherDocs:   10,
						Links: []*fragments.FacetLink{
							{
								Value:         "John Doe",
								DisplayString: "John Doe (25)",
								Count:         25,
								IsSelected:    false,
							},
							{
								Value:         "Jane Smith",
								DisplayString: "Jane Smith (15)",
								Count:         15,
								IsSelected:    true,
							},
						},
					},
				},
			},
			opts: &semantic.QueryOptions{
				Facets: []semantic.FacetRequest{
					{Field: "creator", Limit: 10},
				},
			},
			wantTotal:   5,
			wantItems:   1,
			wantFacets:  1,
			wantErr:     false,
			description: "Should translate facets correctly",
		},
		{
			name: "empty result",
			scrollResult: &fragments.ScrollResultV4{
				Query: &fragments.Query{
					Numfound: 0,
				},
				Items: []*fragments.FragmentGraph{},
			},
			opts:        &semantic.QueryOptions{},
			wantTotal:   0,
			wantItems:   0,
			wantFacets:  0,
			wantErr:     false,
			description: "Should handle empty results",
		},
		{
			name: "fallback when semantic field missing",
			scrollResult: &fragments.ScrollResultV4{
				Query: &fragments.Query{
					Numfound: 1,
				},
				Items: []*fragments.FragmentGraph{
					{
						Meta: &fragments.Header{
							HubID: "item1",
						},
						// No Semantic field
					},
				},
			},
			opts:        &semantic.QueryOptions{},
			wantTotal:   1,
			wantItems:   1,
			wantFacets:  0,
			wantErr:     false,
			description: "Should fallback when Semantic field is missing",
		},
		{
			name:         "nil scroll result",
			scrollResult: nil,
			opts:         &semantic.QueryOptions{},
			wantErr:      true,
			description:  "Should error on nil scroll result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rt.TranslateSearchResult(tt.scrollResult, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("TranslateSearchResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if result.Total != tt.wantTotal {
				t.Errorf("%s: Total = %d, want %d", tt.description, result.Total, tt.wantTotal)
			}

			if len(result.Results) != tt.wantItems {
				t.Errorf("%s: len(Results) = %d, want %d", tt.description, len(result.Results), tt.wantItems)
			}

			if len(result.Facets) != tt.wantFacets {
				t.Errorf("%s: len(Facets) = %d, want %d", tt.description, len(result.Facets), tt.wantFacets)
			}

			// Check that ResultIDs are extracted
			if tt.wantItems > 0 && len(result.ResultIDs) != tt.wantItems {
				t.Errorf("%s: len(ResultIDs) = %d, want %d", tt.description, len(result.ResultIDs), tt.wantItems)
			}
		})
	}
}

func TestResultTranslator_TranslateItems(t *testing.T) {
	rt := NewResultTranslator()

	tests := []struct {
		name        string
		records     []*fragments.FragmentGraph
		wantItems   int
		wantIDs     []string
		wantContext string
		wantErr     bool
		description string
	}{
		{
			name: "semantic items with EDM context",
			records: []*fragments.FragmentGraph{
				{
					Meta: &fragments.Header{
						HubID: "cho123",
					},
					Semantic: &fragments.SemanticView{
						Context: "/contexts/edm/latest/context.jsonld",
						ID:      "http://example.org/cho/123",
						Type:    []string{"ProvidedCHO"},
						Fields: map[string]fragments.SemanticValue{
							"title": &fragments.SemanticLiteral{Value: "Cultural Heritage Object"},
						},
					},
				},
			},
			wantItems:   1,
			wantIDs:     []string{"cho123"},
			wantContext: "/contexts/edm/latest/context.jsonld",
			wantErr:     false,
			description: "Should extract EDM context URL",
		},
		{
			name: "semantic items with inline context",
			records: []*fragments.FragmentGraph{
				{
					Meta: &fragments.Header{
						HubID: "doc456",
					},
					Semantic: &fragments.SemanticView{
						Context: map[string]string{
							"dc":      "http://purl.org/dc/elements/1.1/",
							"dcterms": "http://purl.org/dc/terms/",
						},
						ID:   "http://example.org/doc/456",
						Type: []string{"Document"},
					},
				},
			},
			wantItems: 1,
			wantIDs:   []string{"doc456"},
			wantErr:   false,
			description: "Should handle inline context map",
		},
		{
			name: "fallback items without semantic field",
			records: []*fragments.FragmentGraph{
				{
					Meta: &fragments.Header{
						HubID: "item789",
					},
					// No Semantic field
				},
			},
			wantItems:   1,
			wantIDs:     []string{"item789"},
			wantErr:     false,
			description: "Should fallback gracefully",
		},
		{
			name:        "nil records",
			records:     nil,
			wantItems:   0,
			wantIDs:     []string{},
			wantErr:     false,
			description: "Should handle nil records",
		},
		{
			name: "records with nil items",
			records: []*fragments.FragmentGraph{
				nil,
				{
					Meta: &fragments.Header{
						HubID: "item1",
					},
					Semantic: &fragments.SemanticView{
						ID: "http://example.org/item1",
					},
				},
				nil,
			},
			wantItems:   1,
			wantIDs:     []string{"item1"},
			wantErr:     false,
			description: "Should skip nil records",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, ids, err := rt.translateItems(tt.records)

			if (err != nil) != tt.wantErr {
				t.Errorf("translateItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(items) != tt.wantItems {
				t.Errorf("%s: len(items) = %d, want %d", tt.description, len(items), tt.wantItems)
			}

			if len(ids) != len(tt.wantIDs) {
				t.Errorf("%s: len(ids) = %d, want %d", tt.description, len(ids), len(tt.wantIDs))
			}

			for i, wantID := range tt.wantIDs {
				if i < len(ids) && ids[i] != wantID {
					t.Errorf("%s: ids[%d] = %s, want %s", tt.description, i, ids[i], wantID)
				}
			}

			// Check context if specified
			if tt.wantContext != "" && len(items) > 0 {
				if ctx, ok := items[0]["@context"]; ok {
					if ctxStr, ok := ctx.(string); ok && ctxStr != tt.wantContext {
						t.Errorf("%s: @context = %s, want %s", tt.description, ctxStr, tt.wantContext)
					}
				}
			}
		})
	}
}

func TestResultTranslator_TranslateFacets(t *testing.T) {
	rt := NewResultTranslator()

	tests := []struct {
		name        string
		v2Facets    []*fragments.QueryFacet
		facetReqs   []semantic.FacetRequest
		wantCount   int
		wantType    string
		description string
	}{
		{
			name: "enum facet",
			v2Facets: []*fragments.QueryFacet{
				{
					Field:       "type",
					Name:        "Type",
					Total:       50,
					MissingDocs: 2,
					OtherDocs:   5,
					Links: []*fragments.FacetLink{
						{Value: "image", DisplayString: "Image", Count: 20, IsSelected: false},
						{Value: "video", DisplayString: "Video", Count: 15, IsSelected: true},
					},
				},
			},
			facetReqs:   []semantic.FacetRequest{{Field: "type"}},
			wantCount:   1,
			wantType:    "enum",
			description: "Should translate enum facet",
		},
		{
			name: "date histogram facet",
			v2Facets: []*fragments.QueryFacet{
				{
					Field:        "year",
					Name:         "Year",
					Type:         "histogram",
					DateInterval: "year",
					Total:        100,
					Links: []*fragments.FacetLink{
						{Value: "2020", Count: 30},
						{Value: "2021", Count: 40},
					},
				},
			},
			facetReqs:   []semantic.FacetRequest{{Field: "year"}},
			wantCount:   1,
			wantType:    "date",
			description: "Should identify date histogram facet",
		},
		{
			name: "range facet",
			v2Facets: []*fragments.QueryFacet{
				{
					Field: "price",
					Name:  "Price Range",
					Min:   "0",
					Max:   "1000",
					Total: 200,
					Links: []*fragments.FacetLink{
						{Value: "0-100", Count: 50},
						{Value: "100-500", Count: 80},
					},
				},
			},
			facetReqs:   []semantic.FacetRequest{{Field: "price"}},
			wantCount:   1,
			wantType:    "range",
			description: "Should identify range facet",
		},
		{
			name: "boolean facet",
			v2Facets: []*fragments.QueryFacet{
				{
					Field: "hasImage",
					Name:  "Has Image",
					Total: 100,
					Links: []*fragments.FacetLink{
						{Value: "true", Count: 70},
						{Value: "false", Count: 30},
					},
				},
			},
			facetReqs:   []semantic.FacetRequest{{Field: "hasImage"}},
			wantCount:   1,
			wantType:    "boolean",
			description: "Should identify boolean facet",
		},
		{
			name:        "nil facets",
			v2Facets:    nil,
			facetReqs:   []semantic.FacetRequest{},
			wantCount:   0,
			description: "Should handle nil facets",
		},
		{
			name: "facets with nil items",
			v2Facets: []*fragments.QueryFacet{
				nil,
				{
					Field: "creator",
					Links: []*fragments.FacetLink{
						{Value: "test", Count: 1},
					},
				},
				nil,
			},
			facetReqs:   []semantic.FacetRequest{{Field: "creator"}},
			wantCount:   1,
			description: "Should skip nil facet items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := rt.translateFacets(tt.v2Facets, tt.facetReqs)

			if len(results) != tt.wantCount {
				t.Errorf("%s: len(results) = %d, want %d", tt.description, len(results), tt.wantCount)
				return
			}

			if tt.wantCount > 0 && tt.wantType != "" {
				if results[0].FacetType != tt.wantType {
					t.Errorf("%s: FacetType = %s, want %s", tt.description, results[0].FacetType, tt.wantType)
				}
			}
		})
	}
}

func TestResultTranslator_DetermineFacetType(t *testing.T) {
	rt := NewResultTranslator()

	tests := []struct {
		name     string
		v2Facet  *fragments.QueryFacet
		wantType string
	}{
		{
			name: "date histogram",
			v2Facet: &fragments.QueryFacet{
				Type:         "histogram",
				DateInterval: "year",
			},
			wantType: "date",
		},
		{
			name: "date histogram alt",
			v2Facet: &fragments.QueryFacet{
				Type: "date_histogram",
			},
			wantType: "date",
		},
		{
			name: "range by type",
			v2Facet: &fragments.QueryFacet{
				Type: "range",
			},
			wantType: "range",
		},
		{
			name: "range by min/max",
			v2Facet: &fragments.QueryFacet{
				Min: "0",
				Max: "100",
			},
			wantType: "range",
		},
		{
			name: "boolean facet",
			v2Facet: &fragments.QueryFacet{
				Links: []*fragments.FacetLink{
					{Value: "true"},
					{Value: "false"},
				},
			},
			wantType: "boolean",
		},
		{
			name: "enum facet",
			v2Facet: &fragments.QueryFacet{
				Links: []*fragments.FacetLink{
					{Value: "value1"},
					{Value: "value2"},
					{Value: "value3"},
				},
			},
			wantType: "enum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType := rt.determineFacetType(tt.v2Facet)
			if gotType != tt.wantType {
				t.Errorf("determineFacetType() = %s, want %s", gotType, tt.wantType)
			}
		})
	}
}

func TestResultTranslator_ExtractResultIDs(t *testing.T) {
	rt := NewResultTranslator()

	records := []*fragments.FragmentGraph{
		{Meta: &fragments.Header{HubID: "id1"}},
		nil,
		{Meta: &fragments.Header{HubID: "id2"}},
		{Meta: nil},
		{Meta: &fragments.Header{HubID: "id3"}},
	}

	ids := rt.extractResultIDs(records)

	wantIDs := []string{"id1", "id2", "id3"}
	if len(ids) != len(wantIDs) {
		t.Errorf("len(ids) = %d, want %d", len(ids), len(wantIDs))
	}

	for i, wantID := range wantIDs {
		if i < len(ids) && ids[i] != wantID {
			t.Errorf("ids[%d] = %s, want %s", i, ids[i], wantID)
		}
	}
}
