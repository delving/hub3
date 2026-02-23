package semantic

import (
	"context"

	domainSemantic "github.com/delving/hub3/ikuzo/domain/semantic"
)

// knowledgeContextProvider extracts linked entities from the document graph.
// It scans the document for objects with @id and @type fields and groups them
// by EDM class (edm:Agent, edm:Place, skos:Concept, edm:TimeSpan).
type knowledgeContextProvider struct{}

// NewKnowledgeContextProvider creates a new knowledge context provider.
func NewKnowledgeContextProvider() domainSemantic.IncludeProvider {
	return &knowledgeContextProvider{}
}

// Name returns the provider name used in ?include= query parameter.
func (p *knowledgeContextProvider) Name() string {
	return "knowledgeContext"
}

// Provide generates the knowledge panel for the given document by extracting
// typed entities from the document graph.
func (p *knowledgeContextProvider) Provide(_ context.Context, req *domainSemantic.IncludeRequest) (any, error) {
	if req.Document == nil {
		return nil, nil
	}

	panel := &knowledgePanel{
		Type: "hub3:KnowledgePanel",
	}

	// Scan document for typed entities
	extractEntities(req.Document, panel)

	// Only return if we found something
	if panel.isEmpty() {
		return nil, nil
	}

	return panel.toMap(), nil
}

// edmTypeGroups maps EDM type URIs to panel group keys.
var edmTypeGroups = map[string]string{
	"edm:Agent":    "hub3:agents",
	"edm:Place":    "hub3:places",
	"skos:Concept": "hub3:concepts",
	"edm:TimeSpan": "hub3:timespans",
}

// knowledgePanel collects entities grouped by EDM type.
type knowledgePanel struct {
	Type      string
	Agents    []map[string]any
	Places    []map[string]any
	Concepts  []map[string]any
	Timespans []map[string]any
}

// isEmpty returns true if no entities have been collected.
func (kp *knowledgePanel) isEmpty() bool {
	return len(kp.Agents) == 0 && len(kp.Places) == 0 &&
		len(kp.Concepts) == 0 && len(kp.Timespans) == 0
}

// addEntity adds an entity to the appropriate group based on its EDM type.
func (kp *knowledgePanel) addEntity(typeName string, entity map[string]any) {
	switch typeName {
	case "edm:Agent":
		kp.Agents = append(kp.Agents, entity)
	case "edm:Place":
		kp.Places = append(kp.Places, entity)
	case "skos:Concept":
		kp.Concepts = append(kp.Concepts, entity)
	case "edm:TimeSpan":
		kp.Timespans = append(kp.Timespans, entity)
	}
}

// toMap converts the panel to a map suitable for JSON serialization.
func (kp *knowledgePanel) toMap() map[string]any {
	result := map[string]any{
		"@type": kp.Type,
	}

	if len(kp.Agents) > 0 {
		result["hub3:agents"] = kp.Agents
	}

	if len(kp.Places) > 0 {
		result["hub3:places"] = kp.Places
	}

	if len(kp.Concepts) > 0 {
		result["hub3:concepts"] = kp.Concepts
	}

	if len(kp.Timespans) > 0 {
		result["hub3:timespans"] = kp.Timespans
	}

	return result
}

// extractEntities walks the document looking for typed entities.
// It checks each value in the map: if the value is a map, it checks for
// entity type; if it is a slice, it checks each element.
func extractEntities(obj map[string]any, panel *knowledgePanel) {
	for _, value := range obj {
		switch v := value.(type) {
		case map[string]any:
			checkAndAddEntity(v, panel)
		case []any:
			for _, item := range v {
				if m, ok := item.(map[string]any); ok {
					checkAndAddEntity(m, panel)
				}
			}
		}
	}
}

// checkAndAddEntity checks if an object is a known entity type and adds it
// to the panel. It extracts a summary with @id, @type, and a label predicate.
func checkAndAddEntity(obj map[string]any, panel *knowledgePanel) {
	typeVal, hasType := obj["@type"]
	idVal, hasID := obj["@id"]

	if !hasType || !hasID {
		return
	}

	typeStr, ok := typeVal.(string)
	if !ok {
		return
	}

	// Check if this is a known EDM entity type
	if _, known := edmTypeGroups[typeStr]; !known {
		return
	}

	// Build a summary entity (not the full nested object)
	entity := map[string]any{
		"@id":   idVal,
		"@type": typeStr,
	}

	// Extract label (try common label predicates)
	for _, labelPred := range []string{"skos:prefLabel", "rdfs:label", "dc:title", "foaf:name"} {
		if label, ok := obj[labelPred]; ok {
			entity[labelPred] = label

			break
		}
	}

	panel.addEntity(typeStr, entity)

	// Recurse into nested objects
	extractEntities(obj, panel)
}
