package semantic

// Context represents a JSON-LD @context.
// It can be a simple URL string, a map, or an array of both.
type Context interface{}

// DefaultContext returns the default @context for the semantic API.
func DefaultContext() map[string]interface{} {
	return map[string]interface{}{
		"hydra":   "http://www.w3.org/ns/hydra/core#",
		"dc":      "http://purl.org/dc/elements/1.1/",
		"dcterms": "http://purl.org/dc/terms/",
		"edm":     "http://www.europeana.eu/schemas/edm/",
		"rdf":     "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"rdfs":    "http://www.w3.org/2000/01/rdf-schema#",
		"skos":    "http://www.w3.org/2004/02/skos/core#",
		"ore":     "http://www.openarchives.org/ore/terms/",
		"foaf":    "http://xmlns.com/foaf/0.1/",
		"owl":     "http://www.w3.org/2002/07/owl#",
		"cc":      "https://creativecommons.org/ns#",
		"schema":  "https://schema.org/",
		"vcard":   "http://www.w3.org/2006/vcard/ns#",
		"gn":      "http://www.geonames.org/ontology#",
		"ebucore": "http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#",
		"rdaGr2":  "http://rdvocab.info/ElementsGr2/",
		"nave":    "http://schemas.delving.eu/nave/terms/",
		"hub3":    "https://hub3.delving.org/vocab/",
		"geojson": "https://purl.org/geojson/vocab#",
		"geo":     "http://www.w3.org/2003/01/geo/wgs84_pos#",

		// Hydra vocabulary shortcuts
		"Collection":            "hydra:Collection",
		"PartialCollectionView": "hydra:PartialCollectionView",
		"member":                "hydra:member",
		"totalItems":            "hydra:totalItems",
		"view":                  "hydra:view",
		"search":                "hydra:search",
		"first":                 "hydra:first",
		"previous":              "hydra:previous",
		"next":                  "hydra:next",
		"last":                  "hydra:last",
		"operation":             "hydra:operation",
		"method":                "hydra:method",
		"expects":               "hydra:expects",
		"returns":               "hydra:returns",

		// Hub3-specific vocabulary
		"SearchQuery":            "hub3:SearchQuery",
		"TextQuery":              "hub3:TextQuery",
		"PropertyFilter":         "hub3:PropertyFilter",
		"RangeFilter":            "hub3:RangeFilter",
		"ExistsFilter":           "hub3:ExistsFilter",
		"GeoBBoxFilter":          "hub3:GeoBBoxFilter",
		"GeoDistanceFilter":      "hub3:GeoDistanceFilter",
		"GeoPolygonFilter":       "hub3:GeoPolygonFilter",
		"FacetRequest":           "hub3:FacetRequest",
		"Facet":                  "hub3:Facet",
		"FacetValue":             "hub3:FacetValue",
		"GeoCluster":             "hub3:GeoCluster",
		"SearchResultNavigation": "hub3:SearchResultNavigation",
		"NavigateOperation":      "hub3:NavigateOperation",
	}
}

// CompactContext returns a compact @context with just URLs.
func CompactContext() []string {
	return []string{
		"http://www.w3.org/ns/hydra/core",
		"http://www.europeana.eu/schemas/edm",
		"https://hub3.delving.org/contexts/search",
	}
}

// ContextManager manages @context generation and caching.
type ContextManager struct {
	defaultContext map[string]interface{}
	customContexts map[string]map[string]interface{} // orgID -> custom context
}

// NewContextManager creates a new context manager.
func NewContextManager() *ContextManager {
	return &ContextManager{
		defaultContext: DefaultContext(),
		customContexts: make(map[string]map[string]interface{}),
	}
}

// GetContext returns the context for an organization.
func (cm *ContextManager) GetContext(orgID string) Context {
	if custom, ok := cm.customContexts[orgID]; ok {
		// Merge default and custom contexts
		merged := make(map[string]interface{})
		for k, v := range cm.defaultContext {
			merged[k] = v
		}
		for k, v := range custom {
			merged[k] = v
		}
		return merged
	}
	return cm.defaultContext
}

// SetCustomContext sets a custom context for an organization.
func (cm *ContextManager) SetCustomContext(orgID string, ctx map[string]interface{}) {
	cm.customContexts[orgID] = ctx
}

// AddVocabulary adds a vocabulary namespace to an organization's context.
func (cm *ContextManager) AddVocabulary(orgID, prefix, uri string) {
	if _, ok := cm.customContexts[orgID]; !ok {
		cm.customContexts[orgID] = make(map[string]interface{})
	}
	cm.customContexts[orgID][prefix] = uri
}
