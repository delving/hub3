package search

// Service is the central search service that should be initialised once and
// shared between requests. It is safe for concurrent use by multiple goroutines.
type Service struct {
	responseSize    int
	maxResponseSize int
	facetSize       int
}

// OptionFunc is a function that configures a Service.
// It is used in NewService.
type OptionFunc func(*Service) error

// NewService creates a new Service to query the Hub3 search-index.
//
// NewService, by default, is meant to be long-lived and shared across
// your application.
//
// The caller can configure the new service by passing configuration options
// to the func.
//
// Example:
//
//   service, err := search.NewService(
//     search.ResponseSize(20),
//	 )
//
// An error is also returned when some configuration option is invalid.
func NewService(options ...OptionFunc) (*Service, error) {
	s := &Service{
		responseSize:    16,
		maxResponseSize: 500,
		facetSize:       50,
	}

	// Run the options on it
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// ResponseSize sets the default number of results returned in the Response.
func ResponseSize(size int) OptionFunc {
	return func(s *Service) error {
		if size >= s.maxResponseSize {
			size = s.maxResponseSize
		}
		s.responseSize = size
		return nil
	}
}
