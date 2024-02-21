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
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog"
	"github.com/teris-io/shortid"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/index"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	index             *index.Service
	indexTypes        []string
	postHooks         map[string][]domain.PostHookService
	log               zerolog.Logger
	orgs              domain.OrgConfigRetriever
	scheduler         *gocron.Scheduler
	harvestConfigPath string
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

	sid, err := shortid.New(1, shortid.DefaultABC, 2305)
	if err != nil {
		return nil, err
	}
	shortid.SetDefault(sid)

	if err := s.scheduleTasks(); err != nil {
		return s, fmt.Errorf("unable to start scheduled tasks for bulk service; %w", err)
	}

	return s, nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

func (s *Service) Shutdown(ctx context.Context) error {
	return s.index.Shutdown(ctx)
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "sitemap").Logger()
	s.orgs = b.Orgs
}
