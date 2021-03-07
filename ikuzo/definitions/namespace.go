package definitions

import "github.com/delving/hub3/ikuzo/domain"

// NamespaceService allows you to programmatically manage namespaces
type NamespaceService interface {

	// GetNamespace gets a Namespace
	GetNamespace(GetNamespaceRequest) GetNamespaceResponse

	// ListNamespace returns a list of all namespaces
	ListNamespace(ListNamespaceRequest) ListNamespaceResponse
}

// GetNamespaceRequest is the input object for GetNamespaceRequest.
type GetNamespaceRequest struct {

	// Prefix is the prefix of the Namespace
	// example: "dc"
	Prefix string
}

// GetNamespaceResponse is the output object for GetNamespaceRequest.
type GetNamespaceResponse struct {

	// Namespaces are the namespaces that match the GetNamespaceRequest.Prefix
	Namespaces []*domain.Namespace
}

// ListNamespaceRequest is the input object for ListNamespaceRequest.
type ListNamespaceRequest struct {

	// Prefix is the prefix of the Namespace
	// example: "dc"
	Prefix string

	// Base is the base URI of the Namespace
	// example: "http://purl.org/dc/elements/1.1/"
	Base string
}

// ListNamespaceResponse is the output object for ListNamespaceRequest.
type ListNamespaceResponse struct {

	// Namespaces are the namespaces that match the ListNamespaceRequest Prefix or Base
	Namespaces []*domain.Namespace
}
