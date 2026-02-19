package semantic

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// resolveIntrospectOpts resolves QueryOptions from a ?context= query parameter.
// Returns (opts, contextID, error). opts is nil if no context param provided.
func (s *Service) resolveIntrospectOpts(r *http.Request) (*semantic.QueryOptions, string, error) {
	contextID := r.URL.Query().Get("context")
	if contextID == "" {
		return nil, "", nil
	}

	searchCtx, err := s.store.GetSearchContext(r.Context(), contextID)
	if err != nil {
		return nil, contextID, err
	}

	if searchCtx.IsExpired() {
		_ = s.store.DeleteSearchContext(r.Context(), contextID)
		return nil, contextID, semantic.ErrContextNotFound
	}

	return searchCtx.Query, contextID, nil
}

// buildIntrospectScope builds the scope object for an introspection response.
func buildIntrospectScope(contextID string) map[string]any {
	if contextID != "" {
		return map[string]any{"type": "query", "context": contextID}
	}

	return map[string]any{"type": "index"}
}

// handleIntrospectClasses handles requests to list all RDF classes in the data.
func (s *Service) handleIntrospectClasses(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	opts, contextID, err := s.resolveIntrospectOpts(r)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to resolve query context")
		http.Error(w, "query context not found", http.StatusNotFound)
		return
	}

	classes, err := s.introspect.IntrospectClasses(r.Context(), opts)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect classes")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":        "hub3:IntrospectionResult",
		"hub3:scope":   buildIntrospectScope(contextID),
		"hub3:classes": classes,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}

// handleIntrospectProperties handles requests to list properties for a given class.
func (s *Service) handleIntrospectProperties(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	classURI := chi.URLParam(r, "class")

	opts, contextID, err := s.resolveIntrospectOpts(r)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to resolve query context")
		http.Error(w, "query context not found", http.StatusNotFound)
		return
	}

	props, err := s.introspect.IntrospectProperties(r.Context(), classURI, opts)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect properties")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":           "hub3:PropertyIntrospection",
		"hub3:scope":      buildIntrospectScope(contextID),
		"hub3:class":      classURI,
		"hub3:properties": props,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}

// handleIntrospectField handles requests to get value distribution for a specific field.
func (s *Service) handleIntrospectField(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	field := chi.URLParam(r, "field")

	opts, contextID, err := s.resolveIntrospectOpts(r)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to resolve query context")
		http.Error(w, "query context not found", http.StatusNotFound)
		return
	}

	prop, err := s.introspect.IntrospectField(r.Context(), field, opts)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect field")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	if prop == nil {
		http.Error(w, "field not found", http.StatusNotFound)
		return
	}

	resp := map[string]any{
		"@type":         "hub3:PropertyIntrospection",
		"hub3:scope":    buildIntrospectScope(contextID),
		"hub3:property": prop,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}

// handleIntrospectPaths handles requests to list predicate paths between classes.
func (s *Service) handleIntrospectPaths(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	opts, contextID, err := s.resolveIntrospectOpts(r)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to resolve query context")
		http.Error(w, "query context not found", http.StatusNotFound)
		return
	}

	paths, err := s.introspect.IntrospectPaths(r.Context(), opts)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect paths")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":      "hub3:IntrospectionResult",
		"hub3:scope": buildIntrospectScope(contextID),
		"hub3:paths": paths,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}
