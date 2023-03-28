package nde

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
				total:       7012,
				currentPage: 0,
			},
			&HydraView{
				ID:          "/catalog?page=1",
				Type:        hydraType,
				First:       firstPage,
				Next:        map[string]string{"@id": "/catalog?page=2"},
				Last:        map[string]string{"@id": "/catalog?page=7"},
				baseID:      baseID,
				TotalItems:  7012,
				currentPage: 1,
			},
		},
		{
			"with page; more pages returned",
			fields{
				Type:        hydraType,
				baseID:      baseID,
				total:       7012,
				currentPage: 4,
			},
			&HydraView{
				ID:          "/catalog?page=4",
				Type:        hydraType,
				First:       map[string]string{"@id": "/catalog?page=1"},
				Next:        map[string]string{"@id": "/catalog?page=5"},
				Last:        map[string]string{"@id": "/catalog?page=7"},
				baseID:      baseID,
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
