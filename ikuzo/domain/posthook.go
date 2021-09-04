package domain

import (
	"net/http"

	"github.com/delving/hub3/ikuzo/service/x/resource"
)

// PostHookItem holds the input data that a PostHookService can manipulate
// before submitting it to the endpoint
type PostHookItem struct {
	Graph   *resource.SortedGraph
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
	// Metrics()
	// OrgID returns OrgID that the posthook applies to
	OrgID() string
	Name() string
}
