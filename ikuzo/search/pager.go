package search

// ScrollPager holds all paging information for a search result.
type ScrollPager struct {
	// scrollID is serialized version SearchRequest
	ScrollID string `json:"scrollID"`
	Cursor   int32  `json:"cursor"`
	Total    int64  `json:"total"`
	Rows     int32  `json:"rows"`
}

// NewScrollPager returns a ScrollPager with defaults set
func NewScrollPager() ScrollPager {
	return ScrollPager{}
}
