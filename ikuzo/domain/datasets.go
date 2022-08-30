package domain

import (
	"context"
	"time"
)

type ListOptions struct {
	Start int
	Limit int
}

type Dataset struct {
	OrgID    string    `json:"orgID"`
	ID       string    `json:"datasetID"`
	Modified time.Time `json:"modified"`
}

type DatasetLister interface {
	List(ctx context.Context, options *ListOptions) []*Dataset
}

type Record struct {
	OrgID     string
	DatasetID string
	Modified  time.Time
	Hash      string
}

type RecordLister interface {
	List(ctx context.Context, options *ListOptions) []*Record
}
