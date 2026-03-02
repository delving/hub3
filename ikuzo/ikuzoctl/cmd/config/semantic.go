package config

import (
	"fmt"

	contextsfs "github.com/delving/hub3/docs/contexts"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/domain/semantic"
	semanticService "github.com/delving/hub3/ikuzo/service/x/semantic"
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

	// Build service options
	var serviceOpts []semanticService.Option

	// Create both backends — one as primary, the other as alternate.
	// This enables runtime switching via ?backend= query parameter.

	// V2 adapter backend
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

	// Native ES8 backend
	es8Client := elasticsearch8.NewClientFromExisting(esClient.ES(), logger)
	es8Store := elasticsearch8.NewStore(es8Client, logger)
	es8Introspect := elasticsearch8.NewIntrospectionStore(es8Client, logger)

	// Wire primary and alternate based on config
	var primaryStore semantic.SearchStore
	var primaryName string
	var primaryIntrospect semantic.IntrospectionStore

	if s.UseES8Backend {
		logger.Info().
			Bool("primary", true).
			Str("backend", "es8").
			Msg("initializing semantic API with native ES8 primary backend")

		primaryStore = es8Store
		primaryName = "es8"
		primaryIntrospect = es8Introspect

		// Register v2 adapter as alternate
		serviceOpts = append(serviceOpts,
			semanticService.WithAlternateStore("v2", v2Store),
		)
	} else {
		logger.Info().
			Bool("primary", true).
			Str("backend", "v2").
			Msg("initializing semantic API with v2 adapter primary backend")

		primaryStore = v2Store
		primaryName = "v2"
		primaryIntrospect = v2Introspect

		// Register ES8 as alternate
		serviceOpts = append(serviceOpts,
			semanticService.WithAlternateStore("es8", es8Store),
		)
	}

	_ = es8Introspect // both introspection stores are created for future use

	// Create registry with default resource types
	registry := semantic.DefaultRegistry()

	// Set base URL
	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = "/api/semantic/v1"
	}

	// Add common service options
	serviceOpts = append(serviceOpts,
		semanticService.WithStore(primaryStore),
		semanticService.WithStoreName(primaryName),
		semanticService.WithRegistry(registry),
		semanticService.WithBaseURL(baseURL),
		semanticService.WithIntrospectionStore(primaryIntrospect),
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
