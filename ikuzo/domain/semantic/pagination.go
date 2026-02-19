package semantic

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Pagination represents pagination parameters.
type Pagination struct {
	// Page is the 1-based page number
	Page int `json:"page,omitempty"`

	// Size is the number of items per page
	Size int `json:"size,omitempty"`

	// Offset is the 0-based offset (alternative to Page)
	Offset int `json:"offset,omitempty"`

	// Cursor is an opaque cursor for cursor-based pagination
	Cursor string `json:"cursor,omitempty"`
}

// DefaultPagination returns default pagination settings.
func DefaultPagination() *Pagination {
	return &Pagination{
		Page: 1,
		Size: 20,
	}
}

// GetOffset calculates the offset from page and size.
func (p *Pagination) GetOffset() int {
	if p.Offset > 0 {
		return p.Offset
	}
	if p.Page > 0 {
		return (p.Page - 1) * p.GetSize()
	}
	return 0
}

// GetSize returns the page size, with a default if not set.
func (p *Pagination) GetSize() int {
	if p.Size > 0 {
		return p.Size
	}
	return 20 // Default page size
}

// GetPage returns the page number (1-based).
func (p *Pagination) GetPage() int {
	if p.Page > 0 {
		return p.Page
	}
	return 1
}

// Validate checks if the pagination parameters are valid.
func (p *Pagination) Validate(maxSize int) error {
	if p.Size < 0 {
		return fmt.Errorf("page size cannot be negative")
	}
	if p.Size > maxSize {
		return fmt.Errorf("page size %d exceeds maximum %d", p.Size, maxSize)
	}
	if p.Page < 0 {
		return fmt.Errorf("page number cannot be negative")
	}
	if p.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}

// PartialCollectionView represents Hydra's PartialCollectionView for pagination.
type PartialCollectionView struct {
	TypeValue string `json:"@type"`
	ID        string `json:"@id"`
	First     string `json:"first,omitempty"`
	Previous  string `json:"previous,omitempty"`
	Next      string `json:"next,omitempty"`
	Last      string `json:"last,omitempty"`
}

// Type returns the @type for JSON-LD.
func (pcv *PartialCollectionView) Type() string {
	if pcv.TypeValue == "" {
		return "hydra:PartialCollectionView"
	}
	return pcv.TypeValue
}

// PaginationInfo contains metadata about pagination state.
type PaginationInfo struct {
	CurrentPage  int    `json:"currentPage"`
	PageSize     int    `json:"pageSize"`
	TotalPages   int    `json:"totalPages"`
	TotalItems   int64  `json:"totalItems"`
	HasNext      bool   `json:"hasNext"`
	HasPrevious  bool   `json:"hasPrevious"`
	NextCursor   string `json:"nextCursor,omitempty"`
}

// CalculatePaginationInfo calculates pagination metadata.
func CalculatePaginationInfo(pagination *Pagination, totalItems int64) *PaginationInfo {
	if pagination == nil {
		pagination = DefaultPagination()
	}

	pageSize := pagination.GetSize()
	currentPage := pagination.GetPage()
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))

	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginationInfo{
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		HasNext:      currentPage < totalPages,
		HasPrevious:  currentPage > 1,
	}
}

// BuildPaginationView builds a Hydra PartialCollectionView with links.
func BuildPaginationView(baseURL string, pagination *Pagination, totalItems int64) *PartialCollectionView {
	if pagination == nil {
		pagination = DefaultPagination()
	}

	info := CalculatePaginationInfo(pagination, totalItems)
	currentPage := info.CurrentPage

	view := &PartialCollectionView{
		ID: fmt.Sprintf("%s?page=%d", baseURL, currentPage),
	}

	// First page
	if currentPage > 1 {
		view.First = fmt.Sprintf("%s?page=1", baseURL)
	}

	// Previous page
	if info.HasPrevious {
		view.Previous = fmt.Sprintf("%s?page=%d", baseURL, currentPage-1)
	}

	// Next page
	if info.HasNext {
		view.Next = fmt.Sprintf("%s?page=%d", baseURL, currentPage+1)
	}

	// Last page
	if info.TotalPages > 1 && currentPage < info.TotalPages {
		view.Last = fmt.Sprintf("%s?page=%d", baseURL, info.TotalPages)
	}

	return view
}

const defaultContextTTL = 15 * time.Minute

// NewQueryContext creates a new query context with a short, human-readable ID.
func NewQueryContext(opts *QueryOptions, totalResults int64) *SearchContext {
	id := generateShortID()
	return &SearchContext{
		ID:           id,
		Token:        id,
		Query:        opts,
		TotalResults: totalResults,
		ExpiresAt:    time.Now().Add(defaultContextTTL).Format(time.RFC3339),
	}
}

// SearchContext represents a search session context for detail-level navigation.
type SearchContext struct {
	ID           string        `json:"id"`
	Token        string        `json:"token"`
	Query        *QueryOptions `json:"-"` // Not exposed in JSON
	ResultIDs    []string      `json:"-"` // Ordered list of result IDs
	TotalResults int64         `json:"totalResults"`
	ExpiresAt    string        `json:"expiresAt,omitempty"`
}

// IsExpired checks if the context has expired.
func (sc *SearchContext) IsExpired() bool {
	if sc.ExpiresAt == "" {
		return false
	}
	expires, err := time.Parse(time.RFC3339, sc.ExpiresAt)
	if err != nil {
		return true
	}
	return time.Now().After(expires)
}

// ExtendTTL extends the context's expiration time.
func (sc *SearchContext) ExtendTTL() {
	sc.ExpiresAt = time.Now().Add(defaultContextTTL).Format(time.RFC3339)
}

// NavigationContext represents navigation through search results at item level.
type NavigationContext struct {
	TypeValue      string `json:"@type,omitempty"`
	Position       int    `json:"position"`
	TotalResults   int64  `json:"totalResults"`
	HasNext        bool   `json:"hasNext"`
	HasPrevious    bool   `json:"hasPrevious"`
	First          string `json:"hydra:first,omitempty"`
	Previous       string `json:"hydra:previous,omitempty"`
	Next           string `json:"hydra:next,omitempty"`
	Last           string `json:"hydra:last,omitempty"`
	BackToSearch   string `json:"hub3:backToSearch,omitempty"`
}

// Type returns the @type for JSON-LD.
func (nc *NavigationContext) Type() string {
	if nc.TypeValue == "" {
		return "hub3:SearchResultNavigation"
	}
	return nc.TypeValue
}

// BuildNavigationContext builds navigation context for an item in search results.
func BuildNavigationContext(searchCtx *SearchContext, position int, baseURL string) *NavigationContext {
	if searchCtx == nil || len(searchCtx.ResultIDs) == 0 {
		return nil
	}

	totalResults := int64(len(searchCtx.ResultIDs))
	if searchCtx.TotalResults > 0 {
		totalResults = searchCtx.TotalResults
	}

	nav := &NavigationContext{
		Position:     position,
		TotalResults: totalResults,
		HasNext:      position < len(searchCtx.ResultIDs),
		HasPrevious:  position > 1,
	}

	contextParam := fmt.Sprintf("?ctx=%s", searchCtx.Token)

	// First item
	if position > 1 && len(searchCtx.ResultIDs) > 0 {
		nav.First = fmt.Sprintf("%s/items/%s%s&pos=1", baseURL, searchCtx.ResultIDs[0], contextParam)
	}

	// Previous item
	if nav.HasPrevious && position-2 >= 0 {
		nav.Previous = fmt.Sprintf("%s/items/%s%s&pos=%d", baseURL, searchCtx.ResultIDs[position-2], contextParam, position-1)
	}

	// Next item
	if nav.HasNext && position < len(searchCtx.ResultIDs) {
		nav.Next = fmt.Sprintf("%s/items/%s%s&pos=%d", baseURL, searchCtx.ResultIDs[position], contextParam, position+1)
	}

	// Last item
	lastIdx := len(searchCtx.ResultIDs) - 1
	if position < len(searchCtx.ResultIDs) && lastIdx >= 0 {
		nav.Last = fmt.Sprintf("%s/items/%s%s&pos=%d", baseURL, searchCtx.ResultIDs[lastIdx], contextParam, len(searchCtx.ResultIDs))
	}

	return nav
}

func generateShortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "ctx_" + hex.EncodeToString(b)
}
