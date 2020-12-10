package oaipmh

import (
	"time"

	"github.com/kiivihal/goharvest/oai"
)

type HarvestInfo struct {
	LastCheck    time.Time
	LastModified time.Time
	Error        string
}

type HarvestTask struct {
	OrgID       string
	Name        string
	CheckEvery  time.Duration
	HarvestInfo *HarvestInfo
	Request     oai.Request
	CallbackFn  func(r *oai.Response)
}

// GetLastCheck returns last time the task has run.
func (ht *HarvestTask) GetLastCheck() time.Time {
	if ht.HarvestInfo == nil {
		ht.HarvestInfo = &HarvestInfo{
			LastModified: time.Now(),
			LastCheck:    time.Now(),
		}
	}
	return ht.HarvestInfo.LastCheck
}

// SetLastCheck sets time the task has run.
func (ht *HarvestTask) SetLastCheck(t time.Time) {
	ht.HarvestInfo.LastCheck = t
}

// SetRelativeFrom sets the From param based on the last check minus the duration check.
func (ht *HarvestTask) SetRelativeFrom() {
	ht.Request.From = ht.GetLastCheck().Add(ht.CheckEvery * -1).Format(time.RFC3339)
}
