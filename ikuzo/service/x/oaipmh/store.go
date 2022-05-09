package oaipmh

import "context"

type HarvestStep struct {
	ID        string
	Request   *Request
	TotalSize int
	Finished  bool
}

type QueryConfig struct {
	ID             string // ID of the harvest step
	Identifier     string
	From           string
	Until          string
	MetadataPrefix string
	OrgID          string
	DatasetID      string
	Cursor         string // string based cursor instead of offset
	NextCursor     string
	Offset         int // maybe remove
	Limit          int
	TotalSize      int
	Finished       bool
}

// TODO(kiivihal): implement harvest step
func (q *QueryConfig) NextResumptionToken() *ResumptionToken {
	// TODO(kiivihal): implement logic
	return &ResumptionToken{}
}

type Store interface {
	// TODO(kiivihal): maybe add get Config
	ListSets(ctx context.Context, q *QueryConfig) ([]Set, []Error, error)
	ListIdentifiers(ctx context.Context, q *QueryConfig) ([]Header, []Error, error)
	ListRecords(ctx context.Context, q *QueryConfig) ([]Record, []Error, error)
	GetRecord(ctx context.Context, q *QueryConfig) (Record, []Error, error)
}
