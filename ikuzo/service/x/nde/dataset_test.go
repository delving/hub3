package nde

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/delving/hub3/hub3/models"
)

func TestHydraView_setPager(t *testing.T) {
	const baseID = "/catalog"
	const hydraType = "hydra:PartialCollectionView"
	firstPage := map[string]string{"@id": "/catalog?page=1"}

	type fields struct {
		Type        string
		baseID      string
		total       int
		currentPage int
	}
	tests := []struct {
		name   string
		fields fields
		want   *HydraView
	}{
		{
			"no page",
			fields{
				Type:        hydraType,
				baseID:      baseID,
				total:       100,
				currentPage: 0,
			},
			&HydraView{
				ID:          "/catalog?page=1",
				Type:        hydraType,
				baseID:      baseID,
				First:       firstPage,
				TotalItems:  100,
				currentPage: 1,
			},
		},
		{
			"no page; more pages returned",
			fields{
				Type:        "hydra:PartialCollectionView",
				baseID:      baseID,
				total:       3506,
				currentPage: 0,
			},
			&HydraView{
				ID:          "/catalog?page=1",
				Type:        hydraType,
				First:       firstPage,
				Next:        map[string]string{"@id": "/catalog?page=2"},
				Last:        map[string]string{"@id": "/catalog?page=8"},
				baseID:      baseID,
				TotalItems:  3506,
				currentPage: 1,
			},
		},
		{
			"with page; more pages returned",
			fields{
				Type:        hydraType,
				baseID:      baseID,
				total:       3506,
				currentPage: 4,
			},
			&HydraView{
				ID:          "/catalog?page=4",
				Type:        hydraType,
				First:       map[string]string{"@id": "/catalog?page=1"},
				Next:        map[string]string{"@id": "/catalog?page=5"},
				Last:        map[string]string{"@id": "/catalog?page=8"},
				baseID:      baseID,
				TotalItems:  3506,
				currentPage: 4,
			},
		},
		{
			"page before last has next",
			fields{
				Type:        hydraType,
				baseID:      baseID,
				total:       3506,
				currentPage: 7,
			},
			&HydraView{
				ID:          "/catalog?page=7",
				Type:        hydraType,
				First:       firstPage,
				Next:        map[string]string{"@id": "/catalog?page=8"},
				Last:        map[string]string{"@id": "/catalog?page=8"},
				baseID:      baseID,
				TotalItems:  3506,
				currentPage: 7,
			},
		},
		{
			"last page has no next",
			fields{
				Type:        hydraType,
				baseID:      baseID,
				total:       3506,
				currentPage: 8,
			},
			&HydraView{
				ID:          "/catalog?page=8",
				Type:        hydraType,
				First:       firstPage,
				Last:        map[string]string{"@id": "/catalog?page=8"},
				baseID:      baseID,
				TotalItems:  3506,
				currentPage: 8,
			},
		},
		{
			"exact multiple of page size",
			fields{
				Type:        hydraType,
				baseID:      baseID,
				total:       3500,
				currentPage: 1,
			},
			&HydraView{
				ID:          "/catalog?page=1",
				Type:        hydraType,
				First:       firstPage,
				Next:        map[string]string{"@id": "/catalog?page=2"},
				Last:        map[string]string{"@id": "/catalog?page=7"},
				baseID:      baseID,
				TotalItems:  3500,
				currentPage: 1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			hv := &HydraView{
				Type:        tt.fields.Type,
				baseID:      tt.fields.baseID,
				TotalItems:  tt.fields.total,
				currentPage: tt.fields.currentPage,
			}
			hv.setPager()

			if diff := cmp.Diff(tt.want, hv, cmp.AllowUnexported(HydraView{})); diff != "" {
				t.Errorf("HydraView.setPager() %s; mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestService_createDataSet(t *testing.T) {
	cfg := &RegisterConfig{RDFBaseURL: "https://example.org", URLPrefix: "datasets"}
	s := &Service{defaultCfg: cfg, recordTypeLookup: map[string]*RegisterConfig{"narthex": cfg}}

	d, err := s.createDataSet(&models.DataSet{Spec: "spec-1", Label: "Label 1"})
	if err != nil {
		t.Fatal(err)
	}

	if d.Identifier != "spec-1" || d.Name != "spec-1" || d.Description != "Label 1" {
		t.Errorf("identifier, name, description = %q, %q, %q", d.Identifier, d.Name, d.Description)
	}
}

func TestService_createDataSet_nameFromNarthex(t *testing.T) {
	cfg := &RegisterConfig{RDFBaseURL: "https://example.org", URLPrefix: "datasets"}
	s := &Service{defaultCfg: cfg, recordTypeLookup: map[string]*RegisterConfig{"narthex": cfg}}

	// syncNarthex stores the narthex dataset description as the dataset description
	d, err := s.createDataSet(&models.DataSet{Spec: "spec-1", Description: "Synced title"})
	if err != nil {
		t.Fatal(err)
	}

	if d.Name != "Synced title" || d.Identifier != "spec-1" {
		t.Errorf("name, identifier = %q, %q", d.Name, d.Identifier)
	}
}
