package config

import (
	"fmt"

	contextsfs "github.com/delving/hub3/docs/contexts"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/domain/semantic"
	semanticService "github.com/delving/hub3/ikuzo/service/x/semantic"
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
	UseV2Adapter bool `json:"useV2Adapter"`
	// UseES8Backend is reserved for future native-backend work. The current
	// Semantic V1 contract is delivered through the v2 adapter.
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
	if !s.UseV2Adapter {
		s.UseV2Adapter = true
	}
	if s.UseES8Backend {
		cfg.logger.Warn().
			Msg("semantic.useES8Backend is reserved for future work; using v2 adapter backend")
	}

	// Get Elasticsearch client
	esClient, err := cfg.ElasticSearch.NewCustomClient(&cfg.logger)
	if err != nil {
		return fmt.Errorf("semantic API requires elasticsearch client: %w", err)
	}

	// Get zerolog.Logger from CustomLogger
	logger := cfg.logger.With().Str("svc", "semantic").Logger()

	v2Store := v2adapter.NewV2SearchAdapter(
		esClient.SearchClient(),
		cfg.ElasticSearch.IndexName,
		logger,
	)
	v2Introspect := v2adapter.NewV2IntrospectionAdapter(
		esClient.SearchClient(),
		cfg.ElasticSearch.IndexName,
		logger,
	)

	logger.Info().
		Str("backend", "v2").
		Msg("initializing semantic API with v2 adapter backend")

	// Create registry with default resource types
	registry := semantic.DefaultRegistry()

	// Set base URL
	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = "/api/semantic/v1"
	}

	serviceOpts := []semanticService.Option{
		semanticService.WithStore(v2Store),
		semanticService.WithStoreName("v2"),
		semanticService.WithRegistry(registry),
		semanticService.WithBaseURL(baseURL),
		semanticService.WithIntrospectionStore(v2Introspect),
	}

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
