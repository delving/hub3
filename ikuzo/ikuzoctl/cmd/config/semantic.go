package config

import (
	"fmt"

	contextsfs "github.com/delving/hub3/docs/contexts"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/domain/semantic"
	semanticService "github.com/delving/hub3/ikuzo/service/x/semantic"
	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch"
	"github.com/delving/hub3/ikuzo/storage/x/v2adapter"
)

type Semantic struct {
	Enabled      bool   `json:"enabled"`
	BaseURL      string `json:"baseURL"`
	UseV2Adapter bool   `json:"useV2Adapter"`
}

func (s *Semantic) AddOptions(cfg *Config) error {
	if !s.Enabled {
		return nil
	}

	// Get Elasticsearch client
	esClient, err := cfg.ElasticSearch.NewCustomClient(&cfg.logger)
	if err != nil {
		return fmt.Errorf("semantic API requires elasticsearch client: %w", err)
	}

	// Get zerolog.Logger from CustomLogger
	logger := cfg.logger.With().Str("svc", "semantic").Logger()

	// Create semantic store - choose implementation based on configuration
	var store semantic.SearchStore

	if s.UseV2Adapter {
		// Use v2 adapter for gradual migration
		// OrgID is extracted from request context by the adapter (multi-tenant)
		logger.Info().
			Bool("use_v2_adapter", true).
			Msg("initializing semantic API with v2 adapter (multi-tenant)")

		store = v2adapter.NewV2SearchAdapter(
			esClient.SearchClient(),
			cfg.ElasticSearch.IndexName,
			logger,
		)
	} else {
		// Use direct Elasticsearch implementation
		logger.Info().
			Bool("use_v2_adapter", false).
			Msg("initializing semantic API with direct Elasticsearch store")

		store = elasticsearch.NewSemanticStore(
			esClient.SearchClient(),
			cfg.ElasticSearch.IndexName,
			logger,
		)
	}

	// Create registry with default resource types
	registry := semantic.DefaultRegistry()

	// Set base URL
	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = "/api/semantic/v1"
	}

	// Create semantic service
	svc, err := semanticService.NewService(
		semanticService.WithStore(store),
		semanticService.WithRegistry(registry),
		semanticService.WithBaseURL(baseURL),
	)
	if err != nil {
		return fmt.Errorf("unable to create semantic service: %w", err)
	}

	cfg.options = append(cfg.options, ikuzo.RegisterService(svc))

	if files := contextsfs.List(); len(files) > 0 {
		cfg.options = append(cfg.options, ikuzo.SetRouters(
			semanticService.ContextRouter(contextsfs.Files(), "", files),
		))
	}

	return nil
}
