package bulk

import "github.com/delving/hub3/hub3/fragments"

// PostHookItem holds the input data that a PostHookService can manipulate
// before submitting it to the endpoint
type PostHookItem struct {
	Graph   *fragments.SortedGraph
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
	// Metrics()
	// OrgID returns OrgID that the posthook applies to
	OrgID() string
}
