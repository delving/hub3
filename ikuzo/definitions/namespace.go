package definitions

import "github.com/delving/hub3/ikuzo/domain"

// NamespaceService allows you to programmatically manage namespaces
type NamespaceService interface {

	// DeletetNamespace deletes a Namespace
	// CAUTION: "You may lose data"
	DeleteNamespace(DeleteNamespaceRequest) DeleteNamespaceResponse

	// GetNamespace gets a Namespace
	GetNamespace(GetNamespaceRequest) GetNamespaceResponse

	// PutNamespace stores a Namespace
	PutNamespace(PutNamespaceRequest) PutNamespaceResponse

	// Search returns a filtered list of Namespaces
	Search(SearchNamespaceRequest) SearchNamespaceResponse
}

// GetNamespaceRequest is the input object for GetNamespaceService.GetNamespace
type GetNamespaceRequest struct {

	// ID is the unique identifier of the namespace
	// example: "123"
	ID string
}

// GetNamespaceResponse is the output object for GetNamespaceService.GetNamespace
type GetNamespaceResponse struct {

	// Namespace is the Namespace
	Namespace *domain.Namespace
}

// DeleteNamespaceRequest is the input object for NamespaceService.DeleteNamespace
type DeleteNamespaceRequest struct {
	// ID is the unique identifier of a Namespace
	ID string
}

// DeleteNamespaceRequest is the output object for NamespaceService.DeleteNamespace
type DeleteNamespaceResponse struct{}

// PutNamespaceRequest is the input object for NamespaceService.PutNamespace
type PutNamespaceRequest struct {
	Namespace *domain.Namespace
}

// PutNamespaceResponse is the output object for NamespaceService.PutNamespace
type PutNamespaceResponse struct{}

// SearchNamespaceRequest is the input object for NamespaceService.Search
type SearchNamespaceRequest struct {
	// Prefix for a Namespace
	Prefix string

	// BaseURI for a Namespace
	BaseURI string
}

// SearchNamespaceResponse is the output object for NamespaceService.Search
type SearchNamespaceResponse struct {
	// Hits returns the list of matching Namespaces
	Hits []*domain.Namespace

	// More indicates that there may be more search results. If true, make the same Search request passing this Cursor.
	More bool
}
