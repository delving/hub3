package oaipmh

import (
	"time"

	"github.com/kiivihal/goharvest/oai"
)

type Request struct {
	BaseURL        string
	Set            string
	MetadataPrefix string
	Verb           string
}

type HarvestInfo struct {
	LastCheck    time.Time
	LastModified time.Time
	Error        string
}

type HarvestTask struct {
	OrgID      string
	Name       string
	Request    Request
	CheckEvery int
	CallbackFn func(r *oai.Response)
}

func (ht *HarvestTask) getHarvestInfo() (HarvestInfo, error) {

	return HarvestInfo{}, nil
}
