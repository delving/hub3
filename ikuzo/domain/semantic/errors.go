package semantic

import "fmt"

// Error types for semantic API.
var (
	// ErrInvalidOperator is returned when an invalid operator is used.
	ErrInvalidOperator = fmt.Errorf("invalid operator")

	// ErrInvalidFilter is returned when a filter is invalid.
	ErrInvalidFilter = fmt.Errorf("invalid filter")

	// ErrFieldNotAllowed is returned when filtering on a field that is not allowed.
	ErrFieldNotAllowed = fmt.Errorf("field not allowed for filtering")

	// ErrOperatorNotAllowed is returned when using an operator not allowed for a field.
	ErrOperatorNotAllowed = fmt.Errorf("operator not allowed for field")

	// ErrInvalidPagination is returned when pagination parameters are invalid.
	ErrInvalidPagination = fmt.Errorf("invalid pagination parameters")

	// ErrInvalidGeoCoordinates is returned when geographic coordinates are invalid.
	ErrInvalidGeoCoordinates = fmt.Errorf("invalid geographic coordinates")

	// ErrContextNotFound is returned when a search context is not found.
	ErrContextNotFound = fmt.Errorf("search context not found")

	// ErrContextExpired is returned when a search context has expired.
	ErrContextExpired = fmt.Errorf("search context has expired")
)

// ValidationError represents a validation error with details.
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error implements the error interface.
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", ve.Field, ve.Message)
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string, value interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// HydraError represents a Hydra-compliant error response.
type HydraError struct {
	TypeValue   string                 `json:"@type"`
	Title       string                 `json:"hydra:title"`
	Description string                 `json:"hydra:description"`
	StatusCode  int                    `json:"hub3:statusCode,omitempty"`
	Details     map[string]interface{} `json:"hub3:details,omitempty"`
	Documentation string               `json:"hub3:documentation,omitempty"`
}

// Type returns the @type for JSON-LD.
func (he *HydraError) Type() string {
	if he.TypeValue == "" {
		return "hydra:Error"
	}
	return he.TypeValue
}

// Error implements the error interface.
func (he *HydraError) Error() string {
	return fmt.Sprintf("%s: %s", he.Title, he.Description)
}

// NewHydraError creates a new Hydra error.
func NewHydraError(title, description string, statusCode int) *HydraError {
	return &HydraError{
		TypeValue:   "hydra:Error",
		Title:       title,
		Description: description,
		StatusCode:  statusCode,
	}
}

// WithDetails adds details to a Hydra error.
func (he *HydraError) WithDetails(details map[string]interface{}) *HydraError {
	he.Details = details
	return he
}

// WithDocumentation adds documentation URL to a Hydra error.
func (he *HydraError) WithDocumentation(url string) *HydraError {
	he.Documentation = url
	return he
}
