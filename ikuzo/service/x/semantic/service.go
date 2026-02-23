package semantic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/semantic"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

var _ domain.Service = (*Service)(nil)

// Service provides the semantic search API.
type Service struct {
	store      semantic.SearchStore
	introspect semantic.IntrospectionStore
	registry   *semantic.ConfigRegistry
	log        zerolog.Logger
	baseURL    string
	includes   map[string]semantic.IncludeProvider
}

// NewService creates a new semantic API service.
func NewService(options ...Option) (*Service, error) {
	s := &Service{
		registry: semantic.DefaultRegistry(),
		baseURL:  "/api/semantic/v1",
		includes: make(map[string]semantic.IncludeProvider),
	}

	// Apply options
	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate required dependencies
	if s.store == nil {
		return nil, fmt.Errorf("semantic.SearchStore implementation cannot be nil")
	}

	return s, nil
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

// Routes registers the service routes with the given router.
func (s *Service) Routes(pattern string, r chi.Router) {
	if pattern == "" {
		pattern = s.baseURL
	}

	r.Route(pattern, func(router chi.Router) {
		// Search endpoint - supports both GET and POST
		router.Get("/search", s.handleSearch)
		router.Post("/search", s.handleSearchPost)

		// Detail view endpoint
		router.Get("/resource/{id}", s.handleResourceDetail)

		// Entry point and documentation endpoints
		router.Get("/", s.handleEntryPoint)
		router.Get("/docs", s.handleAPIDocumentation)

		// Resource type specific endpoints
		router.Route("/type/{resourceType}", func(typeRouter chi.Router) {
			typeRouter.Get("/search", s.handleTypeSearch)
			typeRouter.Post("/search", s.handleTypeSearchPost)
			typeRouter.Get("/docs", s.handleTypeDocumentation)
		})

		// Introspection routes
		router.Route("/introspect", func(ir chi.Router) {
			ir.Get("/", s.handleIntrospectClasses)
			ir.Get("/classes", s.handleIntrospectClasses)
			ir.Get("/classes/{class}", s.handleIntrospectProperties)
			ir.Get("/classes/{class}/properties", s.handleIntrospectProperties)
			ir.Get("/fields/{field}", s.handleIntrospectField)
			ir.Get("/paths", s.handleIntrospectPaths)
		})

		// Query context CRUD routes
		router.Route("/contexts/query", func(cr chi.Router) {
			cr.Post("/", s.handleCreateQueryContext)
			cr.Get("/{contextID}", s.handleGetQueryContext)
			cr.Delete("/{contextID}", s.handleDeleteQueryContext)
		})
	})
}

// SetServiceBuilder sets the service builder for dependency injection.
func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "semantic").Logger()
}

// Shutdown gracefully shuts down the service.
func (s *Service) Shutdown(ctx context.Context) error {
	s.log.Info().Msg("shutting down semantic service")
	return nil
}

// GetRegistry returns the configuration registry.
func (s *Service) GetRegistry() *semantic.ConfigRegistry {
	return s.registry
}

// GetStore returns the search store.
func (s *Service) GetStore() semantic.SearchStore {
	return s.store
}

// handleSearch handles GET requests to the search endpoint.
func (s *Service) handleSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters into QueryOptions
	opts, err := parseQueryParams(r)
	if err != nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Invalid Query Parameters",
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	// Get resource config (default to first registered config)
	config := s.registry.GetDefault()
	if config == nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Configuration Error",
			"No resource configuration available",
			http.StatusInternalServerError,
		))
		return
	}

	// Validate query options
	if err := config.ValidateQueryOptions(opts); err != nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Validation Error",
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	// Execute search
	result, err := s.store.Search(ctx, opts, config)
	if err != nil {
		s.log.Error().Err(err).Msg("search failed")
		s.respondError(w, r, semantic.NewHydraError(
			"Search Error",
			"An error occurred while executing the search",
			http.StatusInternalServerError,
		))
		return
	}

	// Build response
	collection := s.buildCollection(r, result, opts, config)

	// Respond with JSON-LD
	s.respondJSON(w, r, collection, http.StatusOK)
}

// handleSearchPost handles POST requests to the search endpoint.
func (s *Service) handleSearchPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse JSON-LD search query
	var searchQuery semantic.SearchQuery
	if err := parseJSONBody(r, &searchQuery); err != nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Invalid Request Body",
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	// Convert to QueryOptions
	opts := searchQuery.ToQueryOptions()

	// Parse filters and facets from raw messages
	if len(searchQuery.Filters) > 0 {
		filters, err := semantic.ParseFilters(searchQuery.Filters)
		if err != nil {
			s.respondError(w, r, semantic.NewHydraError(
				"Invalid Filters",
				err.Error(),
				http.StatusBadRequest,
			))
			return
		}
		opts.Filters = filters
	}

	if len(searchQuery.Facets) > 0 {
		facets, err := semantic.ParseFacets(searchQuery.Facets)
		if err != nil {
			s.respondError(w, r, semantic.NewHydraError(
				"Invalid Facets",
				err.Error(),
				http.StatusBadRequest,
			))
			return
		}
		opts.Facets = facets
	}

	// Get resource config
	config := s.registry.GetDefault()
	if config == nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Configuration Error",
			"No resource configuration available",
			http.StatusInternalServerError,
		))
		return
	}

	// Validate query options
	if err := config.ValidateQueryOptions(opts); err != nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Validation Error",
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	// Execute search
	result, err := s.store.Search(ctx, opts, config)
	if err != nil {
		s.log.Error().Err(err).Msg("search failed")
		s.respondError(w, r, semantic.NewHydraError(
			"Search Error",
			"An error occurred while executing the search",
			http.StatusInternalServerError,
		))
		return
	}

	// Build response
	collection := s.buildCollection(r, result, opts, config)

	// Respond with JSON-LD
	s.respondJSON(w, r, collection, http.StatusOK)
}

// handleResourceDetail handles detail view requests.
func (s *Service) handleResourceDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		s.respondError(w, r, semantic.NewHydraError(
			"Missing Resource ID",
			"Resource ID is required",
			http.StatusBadRequest,
		))
		return
	}

	// Get resource config (use default for now, could be determined from query param)
	config := s.registry.GetDefault()
	if config == nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Configuration Error",
			"No resource configuration available",
			http.StatusInternalServerError,
		))
		return
	}

	// Check for search context token for navigation
	contextToken := r.URL.Query().Get("context")
	var navContext *semantic.NavigationContext

	if contextToken != "" {
		// Try to get search context for navigation
		searchCtx, err := s.store.GetSearchContext(ctx, contextToken)
		if err != nil {
			// Context not found or expired - that's okay, just no navigation
			s.log.Debug().
				Str("token", contextToken).
				Err(err).
				Msg("search context not found or expired")
		} else {
			// Build navigation context from search context
			navContext = s.buildNavigationContext(r, id, searchCtx)
		}
	}

	// Retrieve the document
	doc, err := s.store.GetByID(ctx, id, config)
	if err != nil {
		s.log.Error().Err(err).Str("id", id).Msg("failed to retrieve document")
		s.respondError(w, r, semantic.NewHydraError(
			"Resource Not Found",
			fmt.Sprintf("Resource with ID '%s' not found", id),
			http.StatusNotFound,
		))
		return
	}

	// Add @context if not present
	if _, ok := doc["@context"]; !ok {
		doc["@context"] = semantic.DefaultContext()
	}

	// Add navigation context if available
	if navContext != nil {
		doc["hub3:navigation"] = navContext
	}

	// Process include sections
	includeRequests := parseIncludeParams(r.URL.Query())
	for name, params := range includeRequests {
		provider, ok := s.includes[name]
		if !ok {
			continue // Unknown provider — silently skip
		}

		req := &semantic.IncludeRequest{
			DocumentID: id,
			Document:   doc,
			Config:     config,
			Params:     params,
		}

		section, err := provider.Provide(ctx, req)
		if err != nil {
			s.log.Warn().
				Err(err).
				Str("provider", name).
				Str("id", id).
				Msg("include provider failed")
			continue // Don't fail the whole request
		}

		if section != nil {
			doc["hub3:"+name] = section
		}
	}

	// Respond with JSON-LD
	s.respondJSON(w, r, doc, http.StatusOK)
}

// handleAPIDocumentation is now implemented in documentation.go

// handleTypeSearch handles type-specific search requests (GET).
func (s *Service) handleTypeSearch(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")

	// Get config for this resource type
	config, ok := s.registry.Get(resourceType)
	if !ok {
		s.respondError(w, r, semantic.NewHydraError(
			"Unknown Resource Type",
			fmt.Sprintf("Resource type '%s' is not configured", resourceType),
			http.StatusNotFound,
		))
		return
	}

	// Parse and execute search (similar to handleSearch but with specific config)
	ctx := r.Context()

	opts, err := parseQueryParams(r)
	if err != nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Invalid Query Parameters",
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	if err := config.ValidateQueryOptions(opts); err != nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Validation Error",
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	result, err := s.store.Search(ctx, opts, config)
	if err != nil {
		s.log.Error().Err(err).Msg("search failed")
		s.respondError(w, r, semantic.NewHydraError(
			"Search Error",
			"An error occurred while executing the search",
			http.StatusInternalServerError,
		))
		return
	}

	collection := s.buildCollection(r, result, opts, config)
	s.respondJSON(w, r, collection, http.StatusOK)
}

// handleTypeSearchPost handles type-specific search requests (POST).
func (s *Service) handleTypeSearchPost(w http.ResponseWriter, r *http.Request) {
	// Similar to handleSearchPost but with specific resource type
	w.WriteHeader(http.StatusNotImplemented)
}

// handleTypeDocumentation is now implemented in documentation.go
