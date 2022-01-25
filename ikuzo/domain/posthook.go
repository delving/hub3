package domain

import (
	"net/http"

	"github.com/kiivihal/rdf2go"
)

// PostHookItem holds the input data that a PostHookService can manipulate
// before submitting it to the endpoint
type PostHookItem struct {
	Graph   *rdf2go.Graph
	Deleted bool
	Subject string

	// TODO(kiivihal): replace with single domain.HubID later
	OrgID     string
	DatasetID string
	HubID     string
	Revision  int
}

type PostHookService interface {
	// Add adds PostHookItems to the processing queue
	// Add(item ...PostHookItem) error
	// Publish pushes all the submitted jobs to PostHook endpoint
	Publish(item ...*PostHookItem) error
	Valid(datasetID string) bool
	DropDataset(id string, revision int) (*http.Response, error)
	// run this custom function before DropDataset
	Run(datasetID string) error
	// Metrics()
	// OrgID returns OrgID that the posthook applies to
	OrgID() string
	Name() string
}
