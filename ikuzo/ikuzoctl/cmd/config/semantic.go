package config

import (
	"fmt"

	contextsfs "github.com/delving/hub3/docs/contexts"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/domain/semantic"
	semanticService "github.com/delving/hub3/ikuzo/service/x/semantic"
	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch"
	elasticsearch8 "github.com/delving/hub3/ikuzo/storage/x/elasticsearch8"
	"github.com/delving/hub3/ikuzo/storage/x/v2adapter"
)

type Semantic struct {
	// Enabled controls whether the semantic v1 API is registered.
	// Defaults to true — set to false to explicitly disable.
	Enabled *bool  `json:"enabled"`
	BaseURL string `json:"baseURL"`
	// UseV2Adapter uses the v2 adapter backend which converts
	// FragmentGraph documents to semantic JSON-LD at query time
	// with cycle detection. This is the default backend.
	UseV2Adapter  bool `json:"useV2Adapter"`
	UseES8Backend bool `json:"useES8Backend"`
}

// isEnabled returns true unless explicitly set to false.
func (s *Semantic) isEnabled() bool {
	if s.Enabled == nil {
		return true // default: on
	}

	return *s.Enabled
}

func (s *Semantic) AddOptions(cfg *Config) error {
	if !s.isEnabled() {
		return nil
	}

	// Default to v2 adapter when no backend is explicitly chosen.
	if !s.UseV2Adapter && !s.UseES8Backend {
		s.UseV2Adapter = true
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

	// Build service options (populated below based on backend choice)
	var serviceOpts []semanticService.Option

	if s.UseES8Backend {
		// Use native go-elasticsearch/v8 backend
		logger.Info().
			Bool("use_es8_backend", true).
			Msg("initializing semantic API with native ES8 backend")

		es8Client := elasticsearch8.NewClientFromExisting(esClient.ES(), logger)
		store = elasticsearch8.NewStore(es8Client, logger)

		// Also create introspection store
		introspect := elasticsearch8.NewIntrospectionStore(es8Client, logger)
		serviceOpts = append(serviceOpts, semanticService.WithIntrospectionStore(introspect))
	} else if s.UseV2Adapter {
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

		// Add introspection adapter when using v2 adapter
		introspect := v2adapter.NewV2IntrospectionAdapter(
			esClient.SearchClient(),
			cfg.ElasticSearch.IndexName,
			logger,
		)
		serviceOpts = append(serviceOpts, semanticService.WithIntrospectionStore(introspect))
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

	// Add common service options
	serviceOpts = append(serviceOpts,
		semanticService.WithStore(store),
		semanticService.WithRegistry(registry),
		semanticService.WithBaseURL(baseURL),
	)

	// Create semantic service
	svc, err := semanticService.NewService(serviceOpts...)
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
