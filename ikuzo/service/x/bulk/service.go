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
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type Option func(*Service) error

type Service struct {
	index      *index.Service
	indexTypes []string
	postHooks  map[string][]domain.PostHookService
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		indexTypes: []string{"v2"},
		postHooks:  map[string][]domain.PostHookService{},
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

func SetPostHookService(hooks ...domain.PostHookService) Option {
	return func(s *Service) error {
		for _, hook := range hooks {
			s.postHooks[hook.OrgID()] = append(s.postHooks[hook.OrgID()], hook)
		}

		return nil
	}
}

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func (s *Service) Handle(w http.ResponseWriter, r *http.Request) {
	p := s.NewParser()

	if err := p.Parse(r.Context(), r.Body); err != nil {
		log.Error().Err(err).Msg("issue with bulk request")
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if len(s.postHooks) != 0 && len(p.postHooks) != 0 {
		applyHooks, ok := s.postHooks[p.stats.OrgID]
		if ok {
			go func() {
				for _, hook := range applyHooks {
					validHooks := []*domain.PostHookItem{}

					for _, ph := range p.postHooks {
						if hook.Valid(ph.DatasetID) {
							validHooks = append(validHooks, ph)
						}
					}

					if err := hook.Publish(validHooks...); err != nil {
						log.Error().Err(err).Msg("unable to submit posthooks")
					}

					log.Debug().Int("nr_hooks", len(validHooks)).Msg("submitted posthooks")
				}
			}()
		}
	}

	render.Status(r, http.StatusCreated)
	log.Info().Msgf("stats: %+v", p.stats)
	render.JSON(w, r, p.stats)
}

func (s *Service) NewParser() *Parser {
	p := &Parser{
		stats:         &Stats{},
		indexTypes:    s.indexTypes,
		bi:            s.index,
		sparqlUpdates: []fragments.SparqlUpdate{},
	}

	if len(s.postHooks) != 0 {
		p.postHooks = []*domain.PostHookItem{}
	}

	return p
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// added to implement ikuzo service interface
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}
