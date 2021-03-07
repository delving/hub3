package definitions

import "github.com/delving/hub3/ikuzo/domain"

// NamespaceService allows you to programmatically manage namespaces
type NamespaceService interface {

	// GetNamespace gets a Namespace
	GetNamespace(GetNamespaceRequest) GetNamespaceResponse
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
	Namespaces []domain.Namespace
}
