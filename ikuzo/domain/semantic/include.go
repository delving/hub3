package semantic

import "context"

// IncludeProvider provides an optional section for the detail endpoint.
// Providers are registered by name and invoked when their name appears
// in the ?include= query parameter.
type IncludeProvider interface {
	// Name returns the provider name used in ?include= (e.g., "relatedItems").
	Name() string
	// Provide generates the include section for the given document.
	// The doc is the full document as returned by GetByID.
	// Returns the section data to be added to the response, or nil to skip.
	Provide(ctx context.Context, req *IncludeRequest) (any, error)
}

// IncludeRequest contains everything an IncludeProvider needs.
type IncludeRequest struct {
	// DocumentID is the ID of the resource being viewed.
	DocumentID string
	// Document is the full document as returned by GetByID.
	Document map[string]any
	// Config is the resource configuration.
	Config *ResourceConfig
	// Params are provider-specific parameters from the query string.
	// For example, relatedItems.count=5 would be Params["count"] = "5".
	Params map[string]string
}
