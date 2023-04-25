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
	"database/sql"
	"fmt"
	"net/http"

	stdlog "log"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/index"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	index      *index.Service
	indexTypes []string
	postHooks  map[string][]domain.PostHookService
	log        zerolog.Logger
	orgs       domain.OrgConfigRetriever
	dbPath     string
	db         *sql.DB
	ctx        context.Context
	blobCfg    BlobConfig
	mc         *minio.Client
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		indexTypes: []string{"v2"},
		postHooks:  map[string][]domain.PostHookService{},
		dbPath:     "hub3-bulksvc.db",
		ctx:        context.Background(),
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if err := s.setupDB(); err != nil {
		stdlog.Printf("unable to setup db: %q", err)
		return nil, err
	}

	if blobSetupErr := s.setupBlobStorage(); blobSetupErr != nil {
		return nil, blobSetupErr
	}

	return s, nil
}

func (s *Service) setupDB() error {
	db, err := sql.Open("duckdb", s.dbPath+"?access_mode=READ_WRITE")
	if err != nil {
		s.log.Error().Err(err).Msgf("unable to open duckdb at %s", s.dbPath)
		return err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		return pingErr
	}
	s.db = db

	if setupErr := s.setupTables(); setupErr != nil {
		return setupErr
	}

	s.log.Info().Str("path", s.dbPath).Msg("started duckdb for bulk service")
	log.Printf("started duckdb for bulk service; %q", s.dbPath)

	return nil
}

func (s *Service) setupTables() error {
	query := `create table if not exists dataset (
    orgID text,
    datasetID text,
    published boolean,
);
create unique index if not exists org_dataset_idx ON dataset (orgID, datasetID);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *Service) setupBlobStorage() error {
	if s.blobCfg.Endpoint == "" {
		return fmt.Errorf("blob storage config must have endpoint")
	}
	// Initialize minio client object.
	minioClient, err := minio.New(s.blobCfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.blobCfg.AccessKeyID, s.blobCfg.SecretAccessKey, ""),
		Secure: s.blobCfg.UseSSL,
	})
	if err != nil {
		return err
	}

	s.mc = minioClient

	err = minioClient.MakeBucket(s.ctx, s.blobCfg.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(s.ctx, s.blobCfg.BucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", s.blobCfg.BucketName)
		} else {
			return err
		}
	} else {
		s.log.Info().Msgf("Successfully created %s\n", s.blobCfg.BucketName)
	}

	bucketCfg, err := s.mc.GetBucketVersioning(s.ctx, s.blobCfg.BucketName)
	if err != nil {
		return err
	}

	if !bucketCfg.Enabled() {
		if err := s.mc.EnableVersioning(s.ctx, s.blobCfg.BucketName); err != nil {
			return fmt.Errorf("version must be enabled for bucket %s; %w", s.blobCfg.BucketName, err)
		}
	}

	return nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.db.Close()
	return s.index.Shutdown(ctx)
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "bulk").Logger()
	s.orgs = b.Orgs
}
