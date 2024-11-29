package fragments

type SearchParams struct {
	// Full text search query with support for phrase queries
	// example: Brabant or "Noord Brabant"
	Query string `json:"q"`

	// Page number for pagination
	// example: 1
	// default: 1
	Page int `json:"page"`
}
