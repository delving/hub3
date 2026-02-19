package semantic

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// handleIntrospectClasses handles requests to list all RDF classes in the data.
func (s *Service) handleIntrospectClasses(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	classes, err := s.introspect.IntrospectClasses(r.Context(), nil)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect classes")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":        "hub3:IntrospectionResult",
		"hub3:scope":   map[string]any{"type": "index"},
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

	props, err := s.introspect.IntrospectProperties(r.Context(), classURI, nil)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect properties")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":           "hub3:PropertyIntrospection",
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

	prop, err := s.introspect.IntrospectField(r.Context(), field, nil)
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

	paths, err := s.introspect.IntrospectPaths(r.Context(), nil)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect paths")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":      "hub3:IntrospectionResult",
		"hub3:paths": paths,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}
