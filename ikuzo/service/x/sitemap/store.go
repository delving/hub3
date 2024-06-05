package sitemap

import (
	"context"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
)

type Store interface {
	Datasets(ctx context.Context, cfg domain.SitemapConfig) (locations []Location, err error)
	Locations(ctx context.Context, spec string, cfg domain.SitemapConfig, cb func(loc Location) error) error
}

type Location struct {
	ID          string     `json:"id,omitempty"` // relative path to unique identifier
	LastMod     *time.Time `json:"lastMod,omitempty"`
	RecordCount int64      `json:"-"`
}
