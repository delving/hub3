package oaipmh

import (
	"context"
	"time"
)

type RequestConfig struct {
	ID             string // ID of the harvest step parsed from resumption token
	FirstRequest   *Request
	CurrentRequest RawToken
	StoreCursor    string
	OrgID          string
	DatasetID      string
	TotalSize      int
	Finished       bool
}

func (q *RequestConfig) IsResumedRequest() bool {
	return q.CurrentRequest.HarvestID != ""
}

func (q *RequestConfig) NextResumptionToken(res *Resumable) *ResumptionToken {
	cursor := q.CurrentRequest.Cursor
	token := RawToken{
		HarvestID:    q.ID,
		Cursor:       cursor + res.Len(),
		StorePayload: res.StorePayload,
	}

	rt := &ResumptionToken{
		CompleteListSize: q.TotalSize,
		Cursor:           cursor,
	}

	if rt.CompleteListSize > token.Cursor {
		rt.Token = token.String()
		rt.ExperationDate = time.Now().Add(1 * time.Minute).Format(TimeFormat)
	}

	return rt
}

type Resumable struct {
	Sets         []Set
	Headers      []Header
	Records      []Record
	Errors       []Error
	StorePayload string
	Total        int // Total only needs to be returned the first time
}

func (res *Resumable) Len() int {
	size := len(res.Records)
	if size > 0 {
		return size
	}

	size = len(res.Headers)
	if size > 0 {
		return size
	}

	return len(res.Sets)
}

type Store interface {
	ListSets(ctx context.Context, q *RequestConfig) (Resumable, error)
	ListIdentifiers(ctx context.Context, q *RequestConfig) (Resumable, error)
	ListRecords(ctx context.Context, q *RequestConfig) (Resumable, error)
	GetRecord(ctx context.Context, q *RequestConfig) (Record, []Error, error)
	ListMetadataFormats(ctx context.Context, q *RequestConfig) ([]MetadataFormat, error)
}
