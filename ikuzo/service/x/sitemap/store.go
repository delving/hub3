package sitemap

import (
	"context"
	"time"
)

type Store interface {
	Datasets(ctx context.Context, cfg Config) (locations []Location, err error)
	LocationCount(ctx context.Context, cfg Config) (int, error)
	Locations(ctx context.Context, cfg Config, start, end int) []Location
}

type Location struct {
	ID          string // relative path to unique identifier
	LastMod     *time.Time
	RecordCount int64
}
