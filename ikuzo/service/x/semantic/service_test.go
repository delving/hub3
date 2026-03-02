package semantic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestNewService(t *testing.T) {
	// Test with no store (should fail)
	_, err := NewService()
	if err == nil {
		t.Error("Expected error when creating service without store")
	}

	// Test with mock store (should succeed)
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Errorf("Expected no error with mock store, got: %v", err)
	}

	if svc == nil {
		t.Fatal("Expected service to be created")
	}

	if svc.store == nil {
		t.Error("Expected store to be set")
	}

	if svc.registry == nil {
		t.Error("Expected registry to be initialized")
	}
}

func TestServiceOptions(t *testing.T) {
	store := semantic.NewMockStore()
	customRegistry := semantic.NewConfigRegistry()
	customRegistry.Register(semantic.CHOConfig())

	svc, err := NewService(
		WithStore(store),
		WithRegistry(customRegistry),
		WithBaseURL("/api/v2/semantic"),
	)

	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if svc.baseURL != "/api/v2/semantic" {
		t.Errorf("baseURL = %v, want /api/v2/semantic", svc.baseURL)
	}

	if svc.registry != customRegistry {
		t.Error("Expected custom registry to be used")
	}
}

func TestServiceRoutes(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET search",
			method:         "GET",
			path:           "/api/semantic/v1/search",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST search",
			method:         "POST",
			path:           "/api/semantic/v1/search",
			expectedStatus: http.StatusBadRequest, // Empty body
		},
		{
			name:           "API docs",
			method:         "GET",
			path:           "/api/semantic/v1/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Resource detail",
			method:         "GET",
			path:           "/api/semantic/v1/resource/123",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			svc.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestServiceShutdown(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	if err := svc.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestGetRegistry(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	registry := svc.GetRegistry()
	if registry == nil {
		t.Error("Expected registry to be available")
	}

	// Should have default resource types
	types := registry.ListResourceTypes()
	if len(types) < 3 {
		t.Errorf("Expected at least 3 default resource types, got %d", len(types))
	}
}

func TestGetStore(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if svc.GetStore() != store {
		t.Error("Expected GetStore to return the configured store")
	}
}

func TestResolveStore(t *testing.T) {
	primary := semantic.NewMockStore()
	alternate := semantic.NewMockStore()

	svc, err := NewService(
		WithStore(primary),
		WithStoreName("v2"),
		WithAlternateStore("es8", alternate),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	t.Run("empty backend returns primary", func(t *testing.T) {
		opts := &semantic.QueryOptions{}
		if svc.resolveStore(opts) != primary {
			t.Error("expected primary store for empty backend")
		}
	})

	t.Run("primary name returns primary", func(t *testing.T) {
		opts := &semantic.QueryOptions{Backend: "v2"}
		if svc.resolveStore(opts) != primary {
			t.Error("expected primary store for 'v2' backend")
		}
	})

	t.Run("alternate name returns alternate", func(t *testing.T) {
		opts := &semantic.QueryOptions{Backend: "es8"}
		if svc.resolveStore(opts) != alternate {
			t.Error("expected alternate store for 'es8' backend")
		}
	})

	t.Run("unknown name returns primary", func(t *testing.T) {
		opts := &semantic.QueryOptions{Backend: "unknown"}
		if svc.resolveStore(opts) != primary {
			t.Error("expected primary store for unknown backend")
		}
	})
}

func TestResolveStore_NoAlternate(t *testing.T) {
	primary := semantic.NewMockStore()

	svc, err := NewService(WithStore(primary))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// When no alternate is configured, always return primary
	opts := &semantic.QueryOptions{Backend: "es8"}
	if svc.resolveStore(opts) != primary {
		t.Error("expected primary store when no alternate configured")
	}

	if svc.HasAlternateStore() {
		t.Error("expected HasAlternateStore to return false")
	}
}

func TestHasAlternateStore(t *testing.T) {
	primary := semantic.NewMockStore()
	alternate := semantic.NewMockStore()

	svc, err := NewService(
		WithStore(primary),
		WithAlternateStore("es8", alternate),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if !svc.HasAlternateStore() {
		t.Error("expected HasAlternateStore to return true")
	}
}
