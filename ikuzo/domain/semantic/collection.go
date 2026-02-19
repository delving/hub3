package semantic

// Collection represents a Hydra Collection, which is used for search results.
type Collection struct {
	Context       interface{}            `json:"@context"`
	ID            string                 `json:"@id"`
	Type          []string               `json:"@type"`
	TotalItems    int64                  `json:"totalItems"`
	Member        []interface{}          `json:"member"`
	View          *PartialCollectionView `json:"view,omitempty"`
	Search        *SearchTemplate        `json:"search,omitempty"`
	Facets        []Facet                `json:"hub3:facets,omitempty"`
	ActiveFilters *BreadcrumbList        `json:"hub3:activeFilters,omitempty"`
	GeoData       *GeoFeatureCollection  `json:"hub3:geoData,omitempty"`
	GeoClusters   []GeoCluster           `json:"hub3:geoClusters,omitempty"`
	SearchContext *SearchContext         `json:"hub3:searchContext,omitempty"`
	QueryContext  *QueryContextRef      `json:"hub3:queryContext,omitempty"`
	Timing        *Timing               `json:"hub3:timing,omitempty"`
}

// NewCollection creates a new Hydra Collection with default types.
func NewCollection(id string) *Collection {
	return &Collection{
		ID:   id,
		Type: []string{"hydra:Collection"},
	}
}

// QueryContextRef is a reference to a query context in responses.
type QueryContextRef struct {
	ID      string `json:"@id"`
	Expires string `json:"expires,omitempty"`
}

// Timing contains query execution timing.
type Timing struct {
	Took int64  `json:"took"`
	Unit string `json:"unit"`
}

// Facet represents a facet aggregation result.
type Facet struct {
	TypeValue   string        `json:"@type,omitempty"`
	Field       string        `json:"field"`
	Label       string        `json:"label"`
	PropertyURI string        `json:"propertyURI,omitempty"`
	FacetType   string        `json:"facetType"` // enum, range, date, boolean
	TotalValues int64         `json:"totalValues,omitempty"`
	Values      []FacetValue  `json:"values"`
	Next        string        `json:"next,omitempty"` // URL for more values
}

// Type returns the @type for JSON-LD.
func (f *Facet) Type() string {
	if f.TypeValue == "" {
		return "hub3:Facet"
	}
	return f.TypeValue
}

// FacetValue represents a single value in a facet.
type FacetValue struct {
	TypeValue  string                 `json:"@type,omitempty"`
	Value      string                 `json:"value"`
	Label      string                 `json:"label,omitempty"`
	Count      int64                  `json:"count"`
	Selected   bool                   `json:"selected"`
	Filter     string                 `json:"filter,omitempty"`    // Ready-to-use filter string
	ApplyURL   string                 `json:"applyURL,omitempty"`
	RemoveURL  string                 `json:"removeURL,omitempty"`
	Geometry   map[string]interface{} `json:"geometry,omitempty"` // GeoJSON for region facets
}

// Type returns the @type for JSON-LD.
func (fv *FacetValue) Type() string {
	if fv.TypeValue == "" {
		return "hub3:FacetValue"
	}
	return fv.TypeValue
}

// BreadcrumbList represents active filters as a breadcrumb trail.
type BreadcrumbList struct {
	TypeValue       string               `json:"@type,omitempty"`
	ItemListElement []BreadcrumbListItem `json:"itemListElement"`
}

// Type returns the @type for JSON-LD.
func (bl *BreadcrumbList) Type() string {
	if bl.TypeValue == "" {
		return "hub3:BreadcrumbList"
	}
	return bl.TypeValue
}

// BreadcrumbListItem represents a single breadcrumb item.
type BreadcrumbListItem struct {
	TypeValue string `json:"@type,omitempty"`
	Position  int    `json:"position"`
	Name      string `json:"name"`
	RemoveURL string `json:"hub3:removeURL,omitempty"`
}

// Type returns the @type for JSON-LD.
func (bli *BreadcrumbListItem) Type() string {
	if bli.TypeValue == "" {
		return "hub3:ListItem"
	}
	return bli.TypeValue
}

// GeoFeatureCollection represents a GeoJSON FeatureCollection for map display.
type GeoFeatureCollection struct {
	TypeValue string                 `json:"@type,omitempty"`
	Type      string                 `json:"type"` // "FeatureCollection"
	Features  []GeoFeature           `json:"features"`
	BBox      []float64              `json:"bbox,omitempty"` // [west, south, east, north]
}

// GeoFeature represents a single GeoJSON feature.
type GeoFeature struct {
	Type       string                 `json:"type"` // "Feature"
	ID         string                 `json:"id"`
	Geometry   map[string]interface{} `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// GeoCluster represents a geographic cluster for map display.
type GeoCluster struct {
	TypeValue string     `json:"@type,omitempty"`
	Centroid  *GeoPoint  `json:"centroid"`
	Bounds    *GeoBounds `json:"bounds,omitempty"`
	Count     int64      `json:"count"`
	Zoom      int        `json:"zoom,omitempty"`
	FilterURL string     `json:"filterURL,omitempty"`
}

// Type returns the @type for JSON-LD.
func (gc *GeoCluster) Type() string {
	if gc.TypeValue == "" {
		return "hub3:GeoCluster"
	}
	return gc.TypeValue
}

// Operation represents a Hydra operation available on a resource or collection.
type Operation struct {
	TypeValue   interface{} `json:"@type"`
	ID          string      `json:"@id,omitempty"`
	Method      string      `json:"hydra:method"`
	Title       string      `json:"hydra:title"`
	Description string      `json:"description,omitempty"`
	Expects     string      `json:"hydra:expects,omitempty"`
	Returns     string      `json:"hydra:returns,omitempty"`
}

// NewOperation creates a basic operation.
func NewOperation(method, title string) *Operation {
	return &Operation{
		TypeValue: "hydra:Operation",
		Method:    method,
		Title:     title,
	}
}

// NewNavigateOperation creates a navigation operation for detail-level pagination.
func NewNavigateOperation(method, title, url, direction string) *Operation {
	return &Operation{
		TypeValue: []string{"hydra:Operation", "hub3:NavigateOperation"},
		ID:        url,
		Method:    method,
		Title:     title,
	}
}
