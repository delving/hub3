package semantic

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain/semantic"
	"github.com/go-chi/chi/v5"
)

// queryContextCreateRequest represents the JSON body for creating a query context.
type queryContextCreateRequest struct {
	Query        *queryInput `json:"query,omitempty"`
	TotalResults int64       `json:"totalResults"`
}

// queryInput represents the query portion of a context create request.
type queryInput struct {
	Text string `json:"text"`
}

// handleCreateQueryContext handles POST /contexts/query.
func (s *Service) handleCreateQueryContext(w http.ResponseWriter, r *http.Request) {
	var req queryContextCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, semantic.NewHydraError(
			"Invalid Request Body",
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	var opts *semantic.QueryOptions
	if req.Query != nil {
		opts = &semantic.QueryOptions{
			Query: &semantic.TextQuery{Value: req.Query.Text},
		}
	} else {
		opts = &semantic.QueryOptions{}
	}

	searchCtx := semantic.NewQueryContext(opts, req.TotalResults)

	if err := s.store.SaveSearchContext(r.Context(), searchCtx); err != nil {
		s.log.Error().Err(err).Msg("failed to save query context")
		s.respondError(w, r, semantic.NewHydraError(
			"Context Creation Failed",
			"Failed to save query context",
			http.StatusInternalServerError,
		))
		return
	}

	resp := map[string]any{
		"@type":        "hub3:QueryContext",
		"id":           searchCtx.ID,
		"totalResults": searchCtx.TotalResults,
		"expiresAt":    searchCtx.ExpiresAt,
	}

	s.respondJSON(w, r, resp, http.StatusCreated)
}

// handleGetQueryContext handles GET /contexts/query/{contextID}.
func (s *Service) handleGetQueryContext(w http.ResponseWriter, r *http.Request) {
	contextID := chi.URLParam(r, "contextID")

	searchCtx, err := s.store.GetSearchContext(r.Context(), contextID)
	if err != nil {
		if errors.Is(err, semantic.ErrContextNotFound) {
			s.respondError(w, r, semantic.NewHydraError(
				"Context Not Found",
				"Query context not found or expired",
				http.StatusNotFound,
			))
			return
		}

		s.log.Error().Err(err).Str("contextID", contextID).Msg("failed to get query context")
		s.respondError(w, r, semantic.NewHydraError(
			"Context Retrieval Failed",
			"Failed to retrieve query context",
			http.StatusInternalServerError,
		))
		return
	}

	// Check if expired
	if searchCtx.IsExpired() {
		_ = s.store.DeleteSearchContext(r.Context(), contextID)
		s.respondError(w, r, semantic.NewHydraError(
			"Context Expired",
			"Query context has expired",
			http.StatusNotFound,
		))
		return
	}

	resp := map[string]any{
		"@type":        "hub3:QueryContext",
		"id":           searchCtx.ID,
		"totalResults": searchCtx.TotalResults,
		"expiresAt":    searchCtx.ExpiresAt,
	}

	s.respondJSON(w, r, resp, http.StatusOK)
}

// handleDeleteQueryContext handles DELETE /contexts/query/{contextID}.
func (s *Service) handleDeleteQueryContext(w http.ResponseWriter, r *http.Request) {
	contextID := chi.URLParam(r, "contextID")

	if err := s.store.DeleteSearchContext(r.Context(), contextID); err != nil {
		s.log.Error().Err(err).Str("contextID", contextID).Msg("failed to delete query context")
	}

	w.WriteHeader(http.StatusNoContent)
}
