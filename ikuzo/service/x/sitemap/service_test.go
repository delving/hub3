package sitemap

import (
	"context"
	"reflect"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
)

// fakeStore is a minimal Store implementation used to exercise sitemapRoot
// without a live Elasticsearch backend.
type fakeStore struct{}

// Datasets returns a single fixed dataset so sitemapRoot has something to
// build a sub-sitemap location for.
func (f *fakeStore) Datasets(ctx context.Context, cfg domain.SitemapConfig) ([]Location, error) {
	return []Location{{ID: "spec-a", RecordCount: 1}}, nil
}

// Locations is unused by TestSitemapRootIndexBase but required to satisfy Store.
func (f *fakeStore) Locations(ctx context.Context, spec string, cfg domain.SitemapConfig, cb func(loc Location) error) error {
	return nil
}

func Test_getMaxPages(t *testing.T) {
	type args struct {
		count int64
	}
	tests := []struct {
		name      string
		args      args
		wantPages []int
	}{
		{
			name:      "single page",
			args:      args{count: 100},
			wantPages: []int{1},
		},
		{
			name:      "two page",
			args:      args{count: 65000},
			wantPages: []int{1, 2},
		},
		{
			name:      "multiple pages",
			args:      args{count: 565000},
			wantPages: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPages := getMaxPages(tt.args.count); !reflect.DeepEqual(gotPages, tt.wantPages) {
				t.Errorf("getMaxPages() = %v, want %v", gotPages, tt.wantPages)
			}
		})
	}
}

// TestSitemapRootIndexBase asserts sub-sitemap locations use IndexBaseURL
// when set, so a customer-site BaseURL never leaks into sitemap-index Locs.
func TestSitemapRootIndexBase(t *testing.T) {
	svc, err := NewService(SetStore(&fakeStore{}))
	if err != nil {
		t.Fatal(err)
	}

	cfg := domain.SitemapConfig{
		ID:           "vvu",
		BaseURL:      "https://verhaalvanutrecht.nl/collecties/",
		IndexBaseURL: "https://prod.utralt.hubs.delving.org",
	}

	smi, err := svc.sitemapRoot(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	loc := smi.URLs[0].Loc
	want := "https://prod.utralt.hubs.delving.org/api/sitemap/vvu/spec-a/1"
	if loc != want {
		t.Errorf("sitemapRoot Loc = %q, want %q", loc, want)
	}
}
