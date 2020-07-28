// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package search

// Service is the central search service that should be initialized once and
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
