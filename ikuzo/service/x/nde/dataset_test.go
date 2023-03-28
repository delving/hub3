package nde

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHydraView_setPager(t *testing.T) {
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
				Type:        "hydra:PartialCollectionView",
				baseID:      "/catalog",
				total:       100,
				currentPage: 0,
			},
			&HydraView{
				ID:          "/catalog?page=1",
				Type:        "hydra:PartialCollectionView",
				First:       map[string]string{"@id": "/catalog?page=1"},
				baseID:      "/catalog",
				TotalItems:  100,
				currentPage: 1,
			},
		},
		{
			"no page; more pages returned",
			fields{
				Type:        "hydra:PartialCollectionView",
				baseID:      "/catalog",
				total:       7012,
				currentPage: 0,
			},
			&HydraView{
				ID:          "/catalog?page=1",
				Type:        "hydra:PartialCollectionView",
				First:       map[string]string{"@id": "/catalog?page=1"},
				Next:        map[string]string{"@id": "/catalog?page=2"},
				Last:        map[string]string{"@id": "/catalog?page=7"},
				baseID:      "/catalog",
				TotalItems:  7012,
				currentPage: 1,
			},
		},
		{
			"with page; more pages returned",
			fields{
				Type:        "hydra:PartialCollectionView",
				baseID:      "/catalog",
				total:       7012,
				currentPage: 4,
			},
			&HydraView{
				ID:          "/catalog?page=4",
				Type:        "hydra:PartialCollectionView",
				First:       map[string]string{"@id": "/catalog?page=1"},
				Next:        map[string]string{"@id": "/catalog?page=5"},
				Last:        map[string]string{"@id": "/catalog?page=7"},
				baseID:      "/catalog",
				TotalItems:  7012,
				currentPage: 4,
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
