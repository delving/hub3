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

package bulk

import (
	"context"
	"net/http"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type Option func(*Service) error

type Service struct {
	index      *index.Service
	indexTypes []string
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		indexTypes: []string{"v2"},
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func SetIndexService(is *index.Service) Option {
	return func(s *Service) error {
		s.index = is
		return nil
	}
}

func SetIndexTypes(indexTypes ...string) Option {
	return func(s *Service) error {
		s.indexTypes = indexTypes
		return nil
	}
}

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func (s *Service) Handle(w http.ResponseWriter, r *http.Request) {
	p := s.NewParser()
	if err := p.Parse(r.Context(), r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.Status(r, http.StatusCreated)
	log.Info().Msgf("stats: %#v", p.stats)
	render.JSON(w, r, p.stats)
}

func (s *Service) NewParser() *Parser {
	return &Parser{
		stats:         &Stats{},
		indexTypes:    s.indexTypes,
		bi:            s.index,
		sparqlUpdates: []fragments.SparqlUpdate{},
	}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}
