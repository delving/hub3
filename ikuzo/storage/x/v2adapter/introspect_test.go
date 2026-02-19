package v2adapter

import (
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestBuildClassAggregation(t *testing.T) {
	q := buildClassAggregation()

	src, err := q.Source()
	if err != nil {
		t.Fatalf("failed to build aggregation source: %v", err)
	}

	agg, ok := src.(map[string]any)
	if !ok {
		t.Fatal("expected map source from aggregation")
	}

	if _, ok := agg["nested"]; !ok {
		t.Error("expected nested aggregation for resources")
	}
}

func TestBuildPropertyAggregation(t *testing.T) {
	q := buildPropertyAggregation("edm:ProvidedCHO")

	src, err := q.Source()
	if err != nil {
		t.Fatalf("failed to build aggregation source: %v", err)
	}

	if src == nil {
		t.Error("expected non-nil aggregation source")
	}
}

func TestClassInfoFromBucket(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		count     int64
		wantLabel string
	}{
		{
			name:      "EDM ProvidedCHO",
			uri:       "http://www.europeana.eu/schemas/edm/ProvidedCHO",
			count:     45000,
			wantLabel: "edm:ProvidedCHO",
		},
		{
			name:      "DC Terms Agent",
			uri:       "http://purl.org/dc/terms/Agent",
			count:     1000,
			wantLabel: "dcterms:Agent",
		},
		{
			name:      "SKOS Concept with hash",
			uri:       "http://www.w3.org/2004/02/skos/core#Concept",
			count:     500,
			wantLabel: "skos:Concept",
		},
		{
			name:      "unknown namespace",
			uri:       "http://example.org/UnknownClass",
			count:     10,
			wantLabel: "http://example.org/UnknownClass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := classInfoFromBucket(tt.uri, tt.count)

			if ci.Label != tt.wantLabel {
				t.Errorf("Label = %v, want %v", ci.Label, tt.wantLabel)
			}

			if ci.Count != tt.count {
				t.Errorf("Count = %d, want %d", ci.Count, tt.count)
			}

			if ci.URI != tt.uri {
				t.Errorf("URI = %v, want %v", ci.URI, tt.uri)
			}
		})
	}
}

func TestV2IntrospectionAdapterImplementsInterface(t *testing.T) {
	// Verify compile-time interface compliance
	var _ semantic.IntrospectionStore = (*V2IntrospectionAdapter)(nil)
}
