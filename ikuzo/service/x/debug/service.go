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

package debug

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"

	"github.com/delving/hub3/ikuzo/domain"
)

var _ domain.Service = (*Service)(nil)

// Service provides debug endpoints for testing error tracking.
type Service struct{}

// NewService creates a new debug service.
func NewService() *Service {
	return &Service{}
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

// Routes registers the debug endpoints.
func (s *Service) Routes(pattern string, r chi.Router) {
	r.Get("/debug/trigger-error", s.handleTriggerError())
	r.Get("/debug/trigger-panic", s.handleTriggerPanic())
	r.Get("/debug/trigger-warning", s.handleTriggerWarning())
	r.Get("/debug/health", s.handleHealth())
}

// SetServiceBuilder implements domain.Service.
func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	slog.Info("debug service initialized")
}

// Shutdown implements domain.Service.
func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

// handleTriggerError returns a handler that triggers an error response.
func (s *Service) handleTriggerError() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := errors.New("this is a test error triggered from /debug/trigger-error")

		// Log the error (will be sent to GlitchTip)
		slog.ErrorContext(r.Context(), "test error triggered",
			"error", err,
			"endpoint", r.URL.Path,
			"method", r.Method,
			"user_agent", r.UserAgent(),
			"test_type", "error",
		)

		// Return error response
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   err.Error(),
			"message": "This error has been logged and sent to error tracking",
			"status":  500,
		})
	}
}

// handleTriggerPanic returns a handler that triggers a panic.
func (s *Service) handleTriggerPanic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log before panic
		slog.WarnContext(r.Context(), "about to trigger test panic",
			"endpoint", r.URL.Path,
			"test_type", "panic",
		)

		// This will be caught by the recovery middleware
		panic("this is a test panic triggered from /debug/trigger-panic")
	}
}

// handleTriggerWarning returns a handler that triggers a warning log.
func (s *Service) handleTriggerWarning() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log warning (will be sent to GlitchTip as a warning-level event)
		slog.WarnContext(r.Context(), "test warning triggered",
			"endpoint", r.URL.Path,
			"method", r.Method,
			"user_agent", r.UserAgent(),
			"test_type", "warning",
			"warning_message", "This is a test warning to verify warning capture in error tracking",
		)

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Warning has been logged and sent to error tracking",
			"level":   "warning",
			"status":  200,
		})
	}
}

// handleHealth returns a handler for health checks.
func (s *Service) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"service": "debug",
			"message": "Debug service is healthy",
			"endpoints": []string{
				"/debug/health",
				"/debug/trigger-error",
				"/debug/trigger-panic",
				"/debug/trigger-warning",
			},
		})
	}
}

// TriggerErrorWithContext is a helper function to trigger errors programmatically.
func TriggerErrorWithContext(ctx context.Context, message string) error {
	err := fmt.Errorf("programmatic error: %s", message)
	slog.ErrorContext(ctx, "programmatic error triggered",
		"error", err,
		"stack", string(debug.Stack()),
	)
	return err
}
