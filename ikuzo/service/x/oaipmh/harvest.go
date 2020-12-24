package oaipmh

import (
	"time"

	"github.com/kiivihal/goharvest/oai"
)

const (
	ListRecords     = "ListRecords"
	ListIdentifiers = "ListIdentifiers"
	DateFormat      = "2006-01-02T15:04:05Z"
	UnixStart       = "1970-01-01T12:00:00Z"
)

type HarvestInfo struct {
	LastCheck    time.Time
	LastModified time.Time
	Error        string
}

type HarvestCallback func(r *oai.Response) (recordTime time.Time)

type HarvestTask struct {
	OrgID       string
	Name        string
	CheckEvery  time.Duration
	HarvestInfo *HarvestInfo
	Request     oai.Request
	CallbackFn  HarvestCallback
}

// GetLastCheck returns last time the task has run.
func (ht *HarvestTask) GetLastCheck() time.Time {
	if ht.HarvestInfo == nil {
		ht.HarvestInfo = &HarvestInfo{
			LastModified: time.Now(),
		}
	}
	return ht.HarvestInfo.LastCheck
}

// SetLastCheck sets time the task has run.
func (ht *HarvestTask) SetLastCheck(t time.Time) {
	ht.HarvestInfo.LastCheck = t
}

// SetUnixStartFrom sets the From param to unix start datetime.
func (ht *HarvestTask) SetUnixStartFrom() {
	ht.Request.From = UnixStart
}

// SetRelativeFrom sets the From param based on the last check minus the duration check.
func (ht *HarvestTask) SetRelativeFrom() {
	lt := ht.GetLastCheck()
	if lt.IsZero() {
		lt = time.Now()
	}
	ht.Request.From = lt.Add(ht.CheckEvery * -1).Format(DateFormat)
}
