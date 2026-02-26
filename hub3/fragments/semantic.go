package fragments

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf"
)

// SemanticView is the root structure for semantic format output.
// It provides a type-safe JSON-LD compatible representation of RDF data.
// Context can be either a string URL or a map[string]string for inline context.
type SemanticView struct {
	Context interface{}              `json:"@context,omitempty"`
	ID      string                   `json:"@id"`
	Type    []string                 `json:"@type,omitempty"`
	Graph   []SemanticResource       `json:"@graph,omitempty"`
	Fields  map[string]SemanticValue `json:"-"` // Custom marshaling spreads these at root
}

// SemanticValue represents any semantic value (literal or resource).
// This is the base interface for all value types in the semantic format.
type SemanticValue interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	IsResource() bool
	GetValue() interface{}
}

// SemanticLiteral represents a simple literal value without language or datatype.
type SemanticLiteral struct {
	Value string `json:"-"`
}

// SemanticTypedLiteral represents a literal with language or datatype.
// This corresponds to JSON-LD's @value/@language/@type pattern.
type SemanticTypedLiteral struct {
	Value    interface{} `json:"@value"`
	Language string      `json:"@language,omitempty"`
	Type     string      `json:"@type,omitempty"`
}

// SemanticLanguageMap represents multilingual literals as a map.
// Keys are language codes, values are the literals in that language.
type SemanticLanguageMap map[string]string

// ReferenceType indicates why a resource appears as an ID-only reference.
type ReferenceType string

const (
	// ReferenceNone means the resource is fully expanded inline.
	ReferenceNone ReferenceType = ""
	// ReferenceCycle means the resource exists in this graph but was already
	// expanded higher in the traversal path (circular reference).
	ReferenceCycle ReferenceType = "cycle"
	// ReferenceExternal means the resource is not part of this graph.
	// Fetch it separately via a LOD (Linked Open Data) request.
	ReferenceExternal ReferenceType = "external"
)

// SemanticResource represents a nested resource with its own properties.
type SemanticResource struct {
	ID            string                   `json:"@id,omitempty"`
	Type          []string                 `json:"@type,omitempty"`
	ReferenceType ReferenceType            `json:"-"` // Serialized as hub3:referenceType
	Fields        map[string]SemanticValue `json:"-"` // Custom marshaling
}

// SemanticArray holds multiple values for a single predicate.
type SemanticArray []SemanticValue

// orderedEntry is used internally to maintain order during processing
type orderedEntry struct {
	order int
	value SemanticValue
}

// MarshalJSON for SemanticView - flattens fields at root level
func (sv *SemanticView) MarshalJSON() ([]byte, error) {
	// Create a map to hold all fields
	m := make(map[string]interface{})

	// Add JSON-LD keywords
	if sv.Context != nil {
		// Handle both string URLs and map contexts
		switch ctx := sv.Context.(type) {
		case string:
			if ctx != "" {
				m["@context"] = ctx
			}
		case map[string]string:
			if len(ctx) > 0 {
				m["@context"] = ctx
			}
		default:
			m["@context"] = sv.Context
		}
	}
	if sv.ID != "" {
		m["@id"] = sv.ID
	}
	if len(sv.Type) > 0 {
		if len(sv.Type) == 1 {
			m["@type"] = sv.Type[0]
		} else {
			m["@type"] = sv.Type
		}
	}
	if len(sv.Graph) > 0 {
		m["@graph"] = sv.Graph
	}

	// Spread fields at root level
	for k, v := range sv.Fields {
		m[k] = v
	}

	return json.Marshal(m)
}

// UnmarshalJSON for SemanticView - extracts fields from root level
func (sv *SemanticView) UnmarshalJSON(data []byte) error {
	// First unmarshal to a map
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	sv.Fields = make(map[string]SemanticValue)

	// Process each field
	for k, v := range m {
		switch k {
		case "@context":
			if err := json.Unmarshal(v, &sv.Context); err != nil {
				return err
			}
		case "@id":
			if err := json.Unmarshal(v, &sv.ID); err != nil {
				return err
			}
		case "@type":
			// Handle both string and array
			var typeStr string
			if err := json.Unmarshal(v, &typeStr); err == nil {
				sv.Type = []string{typeStr}
			} else {
				if err := json.Unmarshal(v, &sv.Type); err != nil {
					return err
				}
			}
		case "@graph":
			if err := json.Unmarshal(v, &sv.Graph); err != nil {
				return err
			}
		default:
			// All other fields are semantic values
			val, err := unmarshalSemanticValue(v)
			if err != nil {
				return err
			}
			sv.Fields[k] = val
		}
	}

	return nil
}

// MarshalJSON for SemanticLiteral - returns raw string
func (sl *SemanticLiteral) MarshalJSON() ([]byte, error) {
	return json.Marshal(sl.Value)
}

// UnmarshalJSON for SemanticLiteral
func (sl *SemanticLiteral) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &sl.Value)
}

// IsResource for SemanticLiteral - always false
func (sl *SemanticLiteral) IsResource() bool {
	return false
}

// GetValue for SemanticLiteral
func (sl *SemanticLiteral) GetValue() interface{} {
	return sl.Value
}

// MarshalJSON for SemanticTypedLiteral - standard JSON-LD format
func (stl *SemanticTypedLiteral) MarshalJSON() ([]byte, error) {
	// Use default marshaling for the struct
	type Alias SemanticTypedLiteral
	return json.Marshal((*Alias)(stl))
}

// UnmarshalJSON for SemanticTypedLiteral
func (stl *SemanticTypedLiteral) UnmarshalJSON(data []byte) error {
	type Alias SemanticTypedLiteral
	return json.Unmarshal(data, (*Alias)(stl))
}

// IsResource for SemanticTypedLiteral - always false
func (stl *SemanticTypedLiteral) IsResource() bool {
	return false
}

// GetValue for SemanticTypedLiteral
func (stl *SemanticTypedLiteral) GetValue() interface{} {
	return stl.Value
}

// MarshalJSON for SemanticLanguageMap
func (slm SemanticLanguageMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string(slm))
}

// UnmarshalJSON for SemanticLanguageMap
func (slm *SemanticLanguageMap) UnmarshalJSON(data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*slm = SemanticLanguageMap(m)
	return nil
}

// IsResource for SemanticLanguageMap - always false
func (slm SemanticLanguageMap) IsResource() bool {
	return false
}

// GetValue for SemanticLanguageMap
func (slm SemanticLanguageMap) GetValue() interface{} {
	return map[string]string(slm)
}

// MarshalJSON for SemanticResource - custom field handling
func (sr *SemanticResource) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})

	// Add JSON-LD keywords
	if sr.ID != "" {
		m["@id"] = sr.ID
	}
	if len(sr.Type) > 0 {
		if len(sr.Type) == 1 {
			m["@type"] = sr.Type[0]
		} else {
			m["@type"] = sr.Type
		}
	}

	// Emit reference type so consumers know why this is an ID-only reference
	if sr.ReferenceType != ReferenceNone {
		m["hub3:referenceType"] = string(sr.ReferenceType)
	}

	// Add fields
	for k, v := range sr.Fields {
		m[k] = v
	}

	return json.Marshal(m)
}

// UnmarshalJSON for SemanticResource
func (sr *SemanticResource) UnmarshalJSON(data []byte) error {
	// First unmarshal to a map
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	
	sr.Fields = make(map[string]SemanticValue)
	
	// Process each field
	for k, v := range m {
		switch k {
		case "@id":
			if err := json.Unmarshal(v, &sr.ID); err != nil {
				return err
			}
		case "@type":
			// Handle both string and array
			var typeStr string
			if err := json.Unmarshal(v, &typeStr); err == nil {
				sr.Type = []string{typeStr}
			} else {
				if err := json.Unmarshal(v, &sr.Type); err != nil {
					return err
				}
			}
		case "hub3:referenceType":
			var rt string
			if err := json.Unmarshal(v, &rt); err != nil {
				return err
			}
			sr.ReferenceType = ReferenceType(rt)
		default:
			// All other fields are semantic values
			val, err := unmarshalSemanticValue(v)
			if err != nil {
				return err
			}
			sr.Fields[k] = val
		}
	}
	
	return nil
}

// IsResource for SemanticResource - always true
func (sr *SemanticResource) IsResource() bool {
	return true
}

// GetValue for SemanticResource
func (sr *SemanticResource) GetValue() interface{} {
	return sr
}

// MarshalJSON for SemanticArray
func (sa SemanticArray) MarshalJSON() ([]byte, error) {
	if len(sa) == 0 {
		return json.Marshal([]interface{}{})
	}
	if len(sa) == 1 {
		// Single value - unwrap array
		return json.Marshal(sa[0])
	}
	// Multiple values - keep as array
	return json.Marshal([]SemanticValue(sa))
}

// UnmarshalJSON for SemanticArray
func (sa *SemanticArray) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		*sa = make(SemanticArray, len(arr))
		for i, raw := range arr {
			val, err := unmarshalSemanticValue(raw)
			if err != nil {
				return err
			}
			(*sa)[i] = val
		}
		return nil
	}
	
	// Not an array, treat as single value
	val, err := unmarshalSemanticValue(data)
	if err != nil {
		return err
	}
	*sa = SemanticArray{val}
	return nil
}

// IsResource for SemanticArray - checks first element
func (sa SemanticArray) IsResource() bool {
	if len(sa) > 0 {
		return sa[0].IsResource()
	}
	return false
}

// GetValue for SemanticArray
func (sa SemanticArray) GetValue() interface{} {
	if len(sa) == 1 {
		return sa[0].GetValue()
	}
	values := make([]interface{}, len(sa))
	for i, v := range sa {
		values[i] = v.GetValue()
	}
	return values
}

// unmarshalSemanticValue determines the type and unmarshals accordingly
func unmarshalSemanticValue(data []byte) (SemanticValue, error) {
	// Try to unmarshal as a map first to check for JSON-LD keywords
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err == nil {
		// Check for @value (typed literal)
		if _, hasValue := m["@value"]; hasValue {
			var tl SemanticTypedLiteral
			if err := json.Unmarshal(data, &tl); err != nil {
				return nil, err
			}
			return &tl, nil
		}
		
		// Check for @id or @type (resource)
		if _, hasID := m["@id"]; hasID {
			var r SemanticResource
			if err := json.Unmarshal(data, &r); err != nil {
				return nil, err
			}
			return &r, nil
		}
		if _, hasType := m["@type"]; hasType {
			var r SemanticResource
			if err := json.Unmarshal(data, &r); err != nil {
				return nil, err
			}
			return &r, nil
		}
		
		// Check if it's a language map (all keys are language codes)
		isLangMap := true
		for k := range m {
			if !isLanguageCode(k) {
				isLangMap = false
				break
			}
		}
		if isLangMap && len(m) > 0 {
			var lm SemanticLanguageMap
			if err := json.Unmarshal(data, &lm); err != nil {
				return nil, err
			}
			return &lm, nil
		}
		
		// It's a resource without @id or @type
		var r SemanticResource
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, err
		}
		return &r, nil
	}
	
	// Try array
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		var sa SemanticArray
		if err := sa.UnmarshalJSON(data); err != nil {
			return nil, err
		}
		return &sa, nil
	}
	
	// Must be a simple literal
	var sl SemanticLiteral
	if err := json.Unmarshal(data, &sl.Value); err != nil {
		return nil, err
	}
	return &sl, nil
}

// isLanguageCode checks if a string is a valid language code
func isLanguageCode(code string) bool {
	// Simple check for language codes (e.g., "en", "en-US", "nld")
	if len(code) < 2 || len(code) > 7 {
		return false
	}
	
	parts := strings.Split(code, "-")
	if len(parts) > 2 {
		return false
	}
	
	// Check first part (language)
	lang := parts[0]
	if len(lang) < 2 || len(lang) > 3 {
		return false
	}
	for _, r := range lang {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	
	// Check second part if exists (region)
	if len(parts) == 2 {
		region := parts[1]
		if len(region) != 2 {
			return false
		}
		for _, r := range region {
			if r < 'A' || r > 'Z' {
				return false
			}
		}
	}
	
	return true
}

// NewSemanticValue creates appropriate SemanticValue based on ResourceEntry
func NewSemanticValue(entry *ResourceEntry, rm *ResourceMap) SemanticValue {
	// Initialize empty visited map for top-level calls
	visited := make(map[string]bool)
	return newSemanticValueWithCycleDetection(entry, rm, visited)
}

// newSemanticValueWithCycleDetection creates appropriate SemanticValue with cycle detection
func newSemanticValueWithCycleDetection(entry *ResourceEntry, rm *ResourceMap, visited map[string]bool) SemanticValue {
	switch entry.EntryType {
	case literal:
		return newSemanticLiteralValue(entry)
	case resourceType, bnode:
		return newSemanticResourceValue(entry, rm, visited)
	default:
		return &SemanticLiteral{Value: entry.Value}
	}
}

// newSemanticLiteralValue creates literal with proper type
func newSemanticLiteralValue(entry *ResourceEntry) SemanticValue {
	// Simple literal - no language or datatype
	if entry.Language == "" && entry.DataType == "" {
		return &SemanticLiteral{Value: entry.Value}
	}
	
	// Typed or language literal
	lit := &SemanticTypedLiteral{Value: entry.Value}
	
	if entry.Language != "" {
		lit.Language = entry.Language
	}
	
	if entry.DataType != "" {
		lit.Type = entry.DataType
		// Convert value based on datatype
		lit.Value = convertTypedValue(entry.Value, entry.DataType)
	}
	
	return lit
}

// newSemanticResourceValue creates nested resource with cycle detection
func newSemanticResourceValue(entry *ResourceEntry, rm *ResourceMap, visited map[string]bool) SemanticValue {
	// Check for inline resource first
	if entry.Inline != nil {
		return entry.Inline.toSemanticResourceWithCycleDetection(rm, visited)
	}

	// Then check resource map
	if rm != nil {
		if resource, exists := rm.resources[entry.ID]; exists {
			return resource.toSemanticResourceWithCycleDetection(rm, visited)
		}
	}

	// External resource - not part of this graph, fetch via LOD request
	return &SemanticResource{
		ID:            entry.ID,
		ReferenceType: ReferenceExternal,
	}
}

// convertTypedValue converts string to proper type based on XSD datatype
func convertTypedValue(value, datatype string) interface{} {
	// Handle common XSD datatypes
	switch datatype {
	case "xsd:integer", "http://www.w3.org/2001/XMLSchema#integer":
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	case "xsd:boolean", "http://www.w3.org/2001/XMLSchema#boolean":
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	case "xsd:float", "http://www.w3.org/2001/XMLSchema#float":
		if f, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(f)
		}
	case "xsd:double", "http://www.w3.org/2001/XMLSchema#double",
		"xsd:decimal", "http://www.w3.org/2001/XMLSchema#decimal":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	case "xsd:long", "http://www.w3.org/2001/XMLSchema#long":
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	// Return as string for unknown types or parsing errors
	return value
}

// GenerateSemantic creates type-safe semantic view from FragmentGraph
func (fg *FragmentGraph) GenerateSemantic() *SemanticView {
	sv := &SemanticView{
		Context: fg.buildSemanticContext(),
		Fields:  make(map[string]SemanticValue),
	}

	// Build resource map for lookups
	rm := &ResourceMap{
		resources: make(map[string]*FragmentResource),
	}
	for _, r := range fg.Resources {
		rm.resources[r.ID] = r
	}
	
	// Find main resource (usually the first one or the one matching EntryURI)
	mainResource := fg.findMainResource()
	if mainResource != nil {
		sv.ID = mainResource.ID
		sv.Type = mainResource.Types
		
		// Group entries by SearchLabel (predicate) with order preservation
		predicateMap := make(map[string][]orderedEntry)
		
		for _, entry := range mainResource.Entries {
			// Skip entries without SearchLabel
			if entry.SearchLabel == "" {
				continue
			}
			
			// Skip rdf:type as it's already in @type
			if entry.Predicate == RDFType {
				continue
			}
			
			value := NewSemanticValue(entry, rm)
			predicateMap[entry.SearchLabel] = append(
				predicateMap[entry.SearchLabel],
				orderedEntry{order: entry.Order, value: value},
			)
		}
		
		// Convert to fields with proper typing and order
		for predicate, orderedValues := range predicateMap {
			// Sort by order field
			sort.Slice(orderedValues, func(i, j int) bool {
				return orderedValues[i].order < orderedValues[j].order
			})
			
			// Extract values in order
			values := make([]SemanticValue, len(orderedValues))
			for i, ov := range orderedValues {
				values[i] = ov.value
			}
			
			if len(values) == 1 {
				sv.Fields[predicate] = values[0]
			} else if len(values) > 1 {
				arr := SemanticArray(values)
				sv.Fields[predicate] = &arr
			}
		}
	}
	
	return sv
}

// ToSemanticResource converts FragmentResource to SemanticResource
func (fr *FragmentResource) ToSemanticResource(rm *ResourceMap) *SemanticResource {
	// Initialize empty visited map for top-level calls
	visited := make(map[string]bool)
	return fr.toSemanticResourceWithCycleDetection(rm, visited)
}

// toSemanticResourceWithCycleDetection converts FragmentResource to SemanticResource with cycle detection
func (fr *FragmentResource) toSemanticResourceWithCycleDetection(rm *ResourceMap, visited map[string]bool) *SemanticResource {
	// Check for cycles - if we're already processing this resource, return just the ID
	if visited[fr.ID] {
		return &SemanticResource{
			ID:            fr.ID,
			ReferenceType: ReferenceCycle,
		}
	}

	// Mark this resource as being processed
	visited[fr.ID] = true
	defer func() {
		// Unmark when done processing this branch
		delete(visited, fr.ID)
	}()

	sr := &SemanticResource{
		ID:     fr.ID,
		Type:   fr.Types,
		Fields: make(map[string]SemanticValue),
	}

	// Group entries by SearchLabel with order preservation
	predicateMap := make(map[string][]orderedEntry)

	for _, entry := range fr.Entries {
		// Skip entries without SearchLabel
		if entry.SearchLabel == "" {
			continue
		}

		// Skip rdf:type as it's already in @type
		if entry.Predicate == RDFType {
			continue
		}

		value := newSemanticValueWithCycleDetection(entry, rm, visited)
		predicateMap[entry.SearchLabel] = append(
			predicateMap[entry.SearchLabel],
			orderedEntry{order: entry.Order, value: value},
		)
	}

	// Convert to fields with order preserved
	for predicate, orderedValues := range predicateMap {
		// Sort by order field
		sort.Slice(orderedValues, func(i, j int) bool {
			return orderedValues[i].order < orderedValues[j].order
		})

		// Extract values in order
		values := make([]SemanticValue, len(orderedValues))
		for i, ov := range orderedValues {
			values[i] = ov.value
		}

		if len(values) == 1 {
			sr.Fields[predicate] = values[0]
		} else if len(values) > 1 {
			arr := SemanticArray(values)
			sr.Fields[predicate] = &arr
		}
	}

	return sr
}

// GetUniqueNamespaces extracts all unique namespaces used in the FragmentGraph.
// It efficiently gathers predicates and RDF classes, deduplicates them,
// and extracts their namespaces using domain.SplitURI.
func (fg *FragmentGraph) GetUniqueNamespaces() map[string]string {
	// Use a map to track unique namespace URIs
	namespaceURIs := make(map[string]bool)
	
	// Collect all predicates and types
	for _, resource := range fg.Resources {
		// Add types (RDF classes)
		for _, typeURI := range resource.Types {
			base, _ := domain.SplitURI(typeURI)
			if base != "" {
				namespaceURIs[base] = true
			}
		}
		
		// Add predicates from entries
		for _, entry := range resource.Entries {
			base, _ := domain.SplitURI(entry.Predicate)
			if base != "" {
				namespaceURIs[base] = true
			}
			
			// Also check DataType for typed literals
			if entry.DataType != "" {
				base, _ := domain.SplitURI(entry.DataType)
				if base != "" {
					namespaceURIs[base] = true
				}
			}
		}
	}
	
	// Now resolve the URIs to prefixes using the namespace manager
	namespaces := make(map[string]string)
	
	if rdf.DefaultNamespaceManager != nil {
		for uri := range namespaceURIs {
			// Try to get the namespace with this base URI
			if ns, err := rdf.DefaultNamespaceManager.GetWithBase(uri); err == nil && ns != nil {
				namespaces[ns.Prefix] = ns.URI
			}
		}
	}
	
	return namespaces
}

// buildSemanticContext builds the @context for the semantic view
// Returns either a string URL reference or a map of namespace prefixes
func (fg *FragmentGraph) buildSemanticContext() interface{} {
	// For EDM-based records, reference the EDM context URL
	// Check if this is an EDM record by looking at the types
	isEDM := false
	if len(fg.Resources) > 0 && len(fg.Resources[0].Types) > 0 {
		for _, t := range fg.Resources[0].Types {
			if strings.Contains(t, "europeana.eu/schemas/edm") ||
			   strings.Contains(t, "ProvidedCHO") ||
			   strings.Contains(t, "Aggregation") ||
			   strings.Contains(t, "WebResource") {
				isEDM = true
				break
			}
		}
	}

	// If EDM record, return reference to EDM context URL
	// This will be a relative URL that resolves to the deployed context
	if isEDM {
		return "/contexts/edm/latest/context.jsonld"
	}

	// For non-EDM records, build context from namespaces
	context := fg.GetUniqueNamespaces()

	// Add common namespaces if not present
	commonNamespaces := map[string]string{
		"xsd":     "http://www.w3.org/2001/XMLSchema#",
		"rdf":     "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"rdfs":    "http://www.w3.org/2000/01/rdf-schema#",
		"dc":      "http://purl.org/dc/elements/1.1/",
		"dcterms": "http://purl.org/dc/terms/",
		"skos":    "http://www.w3.org/2004/02/skos/core#",
		"foaf":    "http://xmlns.com/foaf/0.1/",
		"nave":    "http://schemas.delving.eu/nave/terms/",
	}

	for prefix, uri := range commonNamespaces {
		if _, exists := context[prefix]; !exists {
			context[prefix] = uri
		}
	}

	return context
}

// findMainResource finds the main resource in the graph
func (fg *FragmentGraph) findMainResource() *FragmentResource {
	if len(fg.Resources) == 0 {
		return nil
	}
	
	// Try to find resource matching EntryURI
	if fg.Meta != nil && fg.Meta.EntryURI != "" {
		for _, r := range fg.Resources {
			if r.ID == fg.Meta.EntryURI {
				return r
			}
		}
	}
	
	// Fallback to first resource
	return fg.Resources[0]
}