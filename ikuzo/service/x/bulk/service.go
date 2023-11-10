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

	stdlog "log"

	"github.com/go-redis/redis/v8"
	_ "github.com/marcboeker/go-duckdb"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/index"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	index       *index.Service
	indexTypes  []string
	postHooks   map[string][]domain.PostHookService
	log         zerolog.Logger
	orgs        domain.OrgConfigRetriever
	ctx         context.Context
	blobCfg     BlobConfig
	logRequests bool
	rc          *redis.Client
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		indexTypes: []string{"v2"},
		postHooks:  map[string][]domain.PostHookService{},
		ctx:        context.Background(),
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if err := s.setupRedis(); err != nil {
		stdlog.Printf("unable to setup redis: %q", err)
		return nil, err
	}

	return s, nil
}

func (s *Service) setupRedis() error {
	// Create a new Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",   // Redis server address
		Password: "sOmE_sEcUrE_pAsS", // Redis password (if required)
		DB:       1,                  // Redis database index
	})

	// Ping the Redis server to check if it's running
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	fmt.Println("Connected to Redis:", pong)

	s.rc = client

	return nil
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
	s.log = b.Logger.With().Str("svc", "bulk").Logger()
	s.orgs = b.Orgs
}
