package fragments

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemanticValueTypes(t *testing.T) {
	t.Run("SemanticLiteral", func(t *testing.T) {
		lit := &SemanticLiteral{Value: "test value"}
		
		assert.False(t, lit.IsResource())
		assert.Equal(t, "test value", lit.GetValue())
		
		// Test JSON marshaling
		data, err := json.Marshal(lit)
		require.NoError(t, err)
		assert.Equal(t, `"test value"`, string(data))
		
		// Test JSON unmarshaling
		var unmarshaled SemanticLiteral
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, "test value", unmarshaled.Value)
	})
	
	t.Run("SemanticTypedLiteral", func(t *testing.T) {
		// Test with language
		langLit := &SemanticTypedLiteral{
			Value:    "Nederlandse tekst",
			Language: "nl",
		}
		
		assert.False(t, langLit.IsResource())
		assert.Equal(t, "Nederlandse tekst", langLit.GetValue())
		
		data, err := json.Marshal(langLit)
		require.NoError(t, err)
		
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, "Nederlandse tekst", result["@value"])
		assert.Equal(t, "nl", result["@language"])
		
		// Test with datatype
		typedLit := &SemanticTypedLiteral{
			Value: "2024-01-15",
			Type:  "xsd:date",
		}
		
		data, err = json.Marshal(typedLit)
		require.NoError(t, err)
		
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-15", result["@value"])
		assert.Equal(t, "xsd:date", result["@type"])
	})
	
	t.Run("SemanticLanguageMap", func(t *testing.T) {
		langMap := SemanticLanguageMap{
			"nl": "Nederlandse naam",
			"en": "English name",
		}
		
		assert.False(t, langMap.IsResource())
		
		data, err := json.Marshal(langMap)
		require.NoError(t, err)
		
		var result map[string]string
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, "Nederlandse naam", result["nl"])
		assert.Equal(t, "English name", result["en"])
	})
	
	t.Run("SemanticResource", func(t *testing.T) {
		resource := &SemanticResource{
			ID:   "resource/person/123",
			Type: []string{"edm:Agent"},
			Fields: map[string]SemanticValue{
				"skos:prefLabel": &SemanticLiteral{Value: "John Doe"},
			},
		}
		
		assert.True(t, resource.IsResource())
		assert.Equal(t, resource, resource.GetValue())
		
		data, err := json.Marshal(resource)
		require.NoError(t, err)
		
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, "resource/person/123", result["@id"])
		assert.Equal(t, "edm:Agent", result["@type"])
		assert.Equal(t, "John Doe", result["skos:prefLabel"])
	})
	
	t.Run("SemanticArray", func(t *testing.T) {
		// Test single value - should unwrap
		single := SemanticArray{&SemanticLiteral{Value: "single"}}
		data, err := json.Marshal(single)
		require.NoError(t, err)
		assert.Equal(t, `"single"`, string(data))
		
		// Test multiple values - should keep as array
		multiple := SemanticArray{
			&SemanticLiteral{Value: "first"},
			&SemanticLiteral{Value: "second"},
		}
		data, err = json.Marshal(multiple)
		require.NoError(t, err)
		assert.Equal(t, `["first","second"]`, string(data))
	})
}

func TestLanguageCodeValidation(t *testing.T) {
	tests := []struct {
		code  string
		valid bool
	}{
		{"en", true},
		{"nl", true},
		{"en-US", true},
		{"nld", true},
		{"", false},
		{"e", false},
		{"english", false},
		{"en-US-extra", false},
		{"123", false},
		{"EN", false}, // Must be lowercase
	}
	
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.valid, isLanguageCode(tt.code))
		})
	}
}

func TestConvertTypedValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		datatype string
		expected interface{}
	}{
		{"integer", "42", "xsd:integer", 42},
		{"boolean true", "true", "xsd:boolean", true},
		{"boolean false", "false", "xsd:boolean", false},
		{"float", "3.14", "xsd:float", float32(3.14)},
		{"double", "3.14159", "xsd:double", 3.14159},
		{"long", "9223372036854775807", "xsd:long", int64(9223372036854775807)},
		{"unknown type", "value", "custom:type", "value"},
		{"invalid integer", "not-a-number", "xsd:integer", "not-a-number"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTypedValue(tt.value, tt.datatype)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewSemanticValue(t *testing.T) {
	t.Run("creates literal for literal entry", func(t *testing.T) {
		entry := &ResourceEntry{
			EntryType: literal,
			Value:     "test value",
		}
		
		value := NewSemanticValue(entry, nil)
		lit, ok := value.(*SemanticLiteral)
		require.True(t, ok)
		assert.Equal(t, "test value", lit.Value)
	})
	
	t.Run("creates typed literal with language", func(t *testing.T) {
		entry := &ResourceEntry{
			EntryType: literal,
			Value:     "Nederlandse tekst",
			Language:  "nl",
		}
		
		value := NewSemanticValue(entry, nil)
		typedLit, ok := value.(*SemanticTypedLiteral)
		require.True(t, ok)
		assert.Equal(t, "Nederlandse tekst", typedLit.Value)
		assert.Equal(t, "nl", typedLit.Language)
	})
	
	t.Run("creates typed literal with datatype", func(t *testing.T) {
		entry := &ResourceEntry{
			EntryType: literal,
			Value:     "42",
			DataType:  "xsd:integer",
		}
		
		value := NewSemanticValue(entry, nil)
		typedLit, ok := value.(*SemanticTypedLiteral)
		require.True(t, ok)
		assert.Equal(t, 42, typedLit.Value) // Should be converted to int
		assert.Equal(t, "xsd:integer", typedLit.Type)
	})
	
	t.Run("creates resource for resource entry", func(t *testing.T) {
		entry := &ResourceEntry{
			EntryType: resourceType,
			ID:        "resource/123",
		}
		
		value := NewSemanticValue(entry, nil)
		resource, ok := value.(*SemanticResource)
		require.True(t, ok)
		assert.Equal(t, "resource/123", resource.ID)
	})
	
	t.Run("creates resource from inline", func(t *testing.T) {
		inline := &FragmentResource{
			ID:    "inline/resource",
			Types: []string{"edm:Agent"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "skos:prefLabel",
					Predicate:   "http://www.w3.org/2004/02/skos/core#prefLabel",
					Value:       "Inline Name",
					EntryType:   literal,
					Order:       1,
				},
			},
		}
		
		entry := &ResourceEntry{
			EntryType: resourceType,
			ID:        "resource/123",
			Inline:    inline,
		}
		
		rm := &ResourceMap{
			resources: make(map[string]*FragmentResource),
		}
		
		value := NewSemanticValue(entry, rm)
		resource, ok := value.(*SemanticResource)
		require.True(t, ok)
		assert.Equal(t, "inline/resource", resource.ID)
		assert.Equal(t, []string{"edm:Agent"}, resource.Type)
		
		// Check that the inline resource has fields
		prefLabel, exists := resource.Fields["skos:prefLabel"]
		require.True(t, exists)
		lit, ok := prefLabel.(*SemanticLiteral)
		require.True(t, ok)
		assert.Equal(t, "Inline Name", lit.Value)
	})
}

func TestSemanticViewMarshaling(t *testing.T) {
	sv := &SemanticView{
		Context: map[string]string{
			"dc":   "http://purl.org/dc/elements/1.1/",
			"edm":  "http://www.europeana.eu/schemas/edm/",
			"skos": "http://www.w3.org/2004/02/skos/core#",
		},
		ID:   "resource/item/123",
		Type: []string{"edm:ProvidedCHO"},
		Fields: map[string]SemanticValue{
			"dc:title": &SemanticTypedLiteral{
				Value:    "Test Title",
				Language: "en",
			},
			"dc:identifier": &SemanticLiteral{
				Value: "ID123",
			},
			"edm:isRepresentationOf": &SemanticResource{
				ID:   "resource/person/456",
				Type: []string{"edm:Agent"},
				Fields: map[string]SemanticValue{
					"skos:prefLabel": &SemanticLanguageMap{
						"nl": "Nederlandse naam",
						"en": "English name",
					},
				},
			},
		},
	}
	
	data, err := json.Marshal(sv)
	require.NoError(t, err)
	
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	
	// Check @context
	context, ok := result["@context"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "http://purl.org/dc/elements/1.1/", context["dc"])
	
	// Check @id and @type
	assert.Equal(t, "resource/item/123", result["@id"])
	assert.Equal(t, "edm:ProvidedCHO", result["@type"])
	
	// Check fields are at root level
	title, ok := result["dc:title"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test Title", title["@value"])
	assert.Equal(t, "en", title["@language"])
	
	assert.Equal(t, "ID123", result["dc:identifier"])
	
	// Check nested resource
	person, ok := result["edm:isRepresentationOf"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "resource/person/456", person["@id"])
	assert.Equal(t, "edm:Agent", person["@type"])
	
	prefLabel, ok := person["skos:prefLabel"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Nederlandse naam", prefLabel["nl"])
	assert.Equal(t, "English name", prefLabel["en"])
}

func TestFragmentGraphGenerateSemantic(t *testing.T) {
	// Load test data from file
	data, err := os.ReadFile("testdata/nk_fg_grouped.json")
	require.NoError(t, err)
	
	var fg FragmentGraph
	err = json.Unmarshal(data, &fg)
	require.NoError(t, err)
	
	// Generate semantic view
	sv := fg.GenerateSemantic()
	require.NotNil(t, sv)
	
	// Check basic structure
	assert.Equal(t, "https://wo2.collectienederland.nl/id/nk/NK1145", sv.ID)
	assert.Contains(t, sv.Type, "https://wo2.collectienederland.nl/nk/terms/NKRecord")
	
	// Check context has namespaces
	assert.NotEmpty(t, sv.Context)
	
	// Check that fields are present
	assert.NotEmpty(t, sv.Fields)
	
	// Check specific field (thumbnail)
	thumbnail, exists := sv.Fields["nk_thumbnail"]
	require.True(t, exists)
	lit, ok := thumbnail.(*SemanticLiteral)
	require.True(t, ok)
	assert.Contains(t, lit.Value, "https://media.delving.org")
	
	// Check that order is preserved by checking if timeline resource exists
	timeline, exists := sv.Fields["nk_timeline"]
	require.True(t, exists, "nk_timeline field should exist")
	
	// The timeline field should be an array since there are multiple timeline entries
	arr, ok := timeline.(*SemanticArray)
	require.True(t, ok, "timeline should be a *SemanticArray since there are multiple entries")
	require.NotEmpty(t, *arr, "timeline array should not be empty")
	
	// Check the first timeline resource
	firstTimeline, ok := (*arr)[0].(*SemanticResource)
	require.True(t, ok, "first timeline entry should be a *SemanticResource")
	assert.Equal(t, "https://wo2.collectienederland.nl/id/timeline/NK1145/1960-01-01", firstTimeline.ID)
	
	// Verify we have all 7 timeline entries
	assert.Len(t, *arr, 7, "should have 7 timeline entries")
}

func TestOrderPreservation(t *testing.T) {
	fr := &FragmentResource{
		ID:    "test/resource",
		Types: []string{"TestType"},
		Entries: []*ResourceEntry{
			{
				SearchLabel: "field:c",
				Predicate:   "http://example.org/c",
				Value:       "value c",
				EntryType:   literal,
				Order:       3,
			},
			{
				SearchLabel: "field:a",
				Predicate:   "http://example.org/a",
				Value:       "value a",
				EntryType:   literal,
				Order:       1,
			},
			{
				SearchLabel: "field:b",
				Predicate:   "http://example.org/b",
				Value:       "value b",
				EntryType:   literal,
				Order:       2,
			},
		},
	}
	
	sr := fr.ToSemanticResource(nil)
	
	// Check that all fields are present
	assert.Len(t, sr.Fields, 3)
	
	// Marshal to JSON to check order
	_, err := json.Marshal(sr)
	require.NoError(t, err)
	
	// The order in the map doesn't matter for JSON, but when we have multiple
	// values for the same field, they should be ordered
	fr2 := &FragmentResource{
		ID:    "test/resource2",
		Types: []string{"TestType"},
		Entries: []*ResourceEntry{
			{
				SearchLabel: "field:multi",
				Predicate:   "http://example.org/multi",
				Value:       "third",
				EntryType:   literal,
				Order:       3,
			},
			{
				SearchLabel: "field:multi",
				Predicate:   "http://example.org/multi",
				Value:       "first",
				EntryType:   literal,
				Order:       1,
			},
			{
				SearchLabel: "field:multi",
				Predicate:   "http://example.org/multi",
				Value:       "second",
				EntryType:   literal,
				Order:       2,
			},
		},
	}
	
	sr2 := fr2.ToSemanticResource(nil)
	multiField := sr2.Fields["field:multi"]
	require.NotNil(t, multiField)
	
	arr, ok := multiField.(*SemanticArray)
	require.True(t, ok)
	require.Len(t, *arr, 3)
	
	// Check order
	assert.Equal(t, "first", (*arr)[0].(*SemanticLiteral).Value)
	assert.Equal(t, "second", (*arr)[1].(*SemanticLiteral).Value)
	assert.Equal(t, "third", (*arr)[2].(*SemanticLiteral).Value)
}

func TestGetUniqueNamespaces(t *testing.T) {
	fg := &FragmentGraph{
		Resources: []*FragmentResource{
			{
				ID:    "test/resource",
				Types: []string{"http://purl.org/dc/terms/BibliographicResource"},
				Entries: []*ResourceEntry{
					{
						Predicate: "http://purl.org/dc/elements/1.1/title",
						DataType:  "http://www.w3.org/2001/XMLSchema#string",
					},
					{
						Predicate: "http://www.europeana.eu/schemas/edm/isShownAt",
					},
				},
			},
		},
	}
	
	// Note: This test will only work if the namespace manager is properly initialized
	// In a real test environment, you would need to set up the namespace manager first
	namespaces := fg.GetUniqueNamespaces()
	
	// The function should extract unique namespace URIs
	// Since we don't have the namespace manager set up in tests,
	// we just check that the function doesn't panic
	assert.NotNil(t, namespaces)
}

func TestUnmarshalSemanticValue(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected func(t *testing.T, val SemanticValue)
	}{
		{
			name: "simple literal",
			json: `"simple string"`,
			expected: func(t *testing.T, val SemanticValue) {
				lit, ok := val.(*SemanticLiteral)
				require.True(t, ok)
				assert.Equal(t, "simple string", lit.Value)
			},
		},
		{
			name: "typed literal",
			json: `{"@value": "2024-01-15", "@type": "xsd:date"}`,
			expected: func(t *testing.T, val SemanticValue) {
				tl, ok := val.(*SemanticTypedLiteral)
				require.True(t, ok)
				assert.Equal(t, "2024-01-15", tl.Value)
				assert.Equal(t, "xsd:date", tl.Type)
			},
		},
		{
			name: "language literal",
			json: `{"@value": "Dutch text", "@language": "nl"}`,
			expected: func(t *testing.T, val SemanticValue) {
				tl, ok := val.(*SemanticTypedLiteral)
				require.True(t, ok)
				assert.Equal(t, "Dutch text", tl.Value)
				assert.Equal(t, "nl", tl.Language)
			},
		},
		{
			name: "language map",
			json: `{"nl": "Nederlandse tekst", "en": "English text"}`,
			expected: func(t *testing.T, val SemanticValue) {
				lm, ok := val.(*SemanticLanguageMap)
				require.True(t, ok)
				assert.Equal(t, "Nederlandse tekst", (*lm)["nl"])
				assert.Equal(t, "English text", (*lm)["en"])
			},
		},
		{
			name: "resource with @id",
			json: `{"@id": "resource/123", "@type": "Person", "name": "John"}`,
			expected: func(t *testing.T, val SemanticValue) {
				r, ok := val.(*SemanticResource)
				require.True(t, ok)
				assert.Equal(t, "resource/123", r.ID)
				assert.Equal(t, []string{"Person"}, r.Type)
			},
		},
		{
			name: "array of values",
			json: `["first", "second", "third"]`,
			expected: func(t *testing.T, val SemanticValue) {
				arr, ok := val.(*SemanticArray)
				require.True(t, ok)
				assert.Len(t, *arr, 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := unmarshalSemanticValue([]byte(tt.json))
			require.NoError(t, err)
			tt.expected(t, val)
		})
	}
}

func TestCycleDetection(t *testing.T) {
	t.Run("simple cycle A→B→A", func(t *testing.T) {
		// Create resources A and B
		resourceA := &FragmentResource{
			ID:    "resource/A",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/B",
					Order:       1,
				},
			},
		}

		resourceB := &FragmentResource{
			ID:    "resource/B",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/A",
					Order:       1,
				},
			},
		}

		rm := &ResourceMap{
			resources: map[string]*FragmentResource{
				"resource/A": resourceA,
				"resource/B": resourceB,
			},
		}

		// Convert A to semantic - should not stack overflow
		sr := resourceA.ToSemanticResource(rm)
		require.NotNil(t, sr)
		assert.Equal(t, "resource/A", sr.ID)

		// Check that B is nested
		knowsValue, exists := sr.Fields["knows"]
		require.True(t, exists)
		nestedB, ok := knowsValue.(*SemanticResource)
		require.True(t, ok)
		assert.Equal(t, "resource/B", nestedB.ID)

		// Check that A inside B is just an ID reference (cycle detected)
		nestedBKnows, exists := nestedB.Fields["knows"]
		require.True(t, exists)
		cycleRefA, ok := nestedBKnows.(*SemanticResource)
		require.True(t, ok)
		assert.Equal(t, "resource/A", cycleRefA.ID)
		// Should be empty - just an ID reference, no nested fields
		assert.Empty(t, cycleRefA.Fields, "cycle reference should only contain ID, not nested fields")
		assert.Equal(t, ReferenceCycle, cycleRefA.ReferenceType, "cycle reference should have ReferenceType=cycle")
	})

	t.Run("longer cycle A→B→C→A", func(t *testing.T) {
		resourceA := &FragmentResource{
			ID:    "resource/A",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/B",
					Order:       1,
				},
			},
		}

		resourceB := &FragmentResource{
			ID:    "resource/B",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/C",
					Order:       1,
				},
			},
		}

		resourceC := &FragmentResource{
			ID:    "resource/C",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/A",
					Order:       1,
				},
			},
		}

		rm := &ResourceMap{
			resources: map[string]*FragmentResource{
				"resource/A": resourceA,
				"resource/B": resourceB,
				"resource/C": resourceC,
			},
		}

		// Should not stack overflow
		sr := resourceA.ToSemanticResource(rm)
		require.NotNil(t, sr)
		assert.Equal(t, "resource/A", sr.ID)

		// Navigate A→B→C
		b := sr.Fields["knows"].(*SemanticResource)
		assert.Equal(t, "resource/B", b.ID)

		c := b.Fields["knows"].(*SemanticResource)
		assert.Equal(t, "resource/C", c.ID)

		// The reference back to A should be just an ID (cycle detected)
		cycleA := c.Fields["knows"].(*SemanticResource)
		assert.Equal(t, "resource/A", cycleA.ID)
		assert.Empty(t, cycleA.Fields, "cycle reference should only contain ID")
		assert.Equal(t, ReferenceCycle, cycleA.ReferenceType)
	})

	t.Run("cycle to middle A→B→C→B", func(t *testing.T) {
		resourceA := &FragmentResource{
			ID:    "resource/A",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/B",
					Order:       1,
				},
			},
		}

		resourceB := &FragmentResource{
			ID:    "resource/B",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/C",
					Order:       1,
				},
			},
		}

		resourceC := &FragmentResource{
			ID:    "resource/C",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/B", // Cycle back to B, not A
					Order:       1,
				},
			},
		}

		rm := &ResourceMap{
			resources: map[string]*FragmentResource{
				"resource/A": resourceA,
				"resource/B": resourceB,
				"resource/C": resourceC,
			},
		}

		// Should not stack overflow
		sr := resourceA.ToSemanticResource(rm)
		require.NotNil(t, sr)

		// Navigate A→B→C→B (cycle)
		b := sr.Fields["knows"].(*SemanticResource)
		c := b.Fields["knows"].(*SemanticResource)
		cycleB := c.Fields["knows"].(*SemanticResource)

		assert.Equal(t, "resource/B", cycleB.ID)
		assert.Empty(t, cycleB.Fields, "cycle reference should only contain ID")
		assert.Equal(t, ReferenceCycle, cycleB.ReferenceType)
	})

	t.Run("no cycle - normal nested resources", func(t *testing.T) {
		resourceA := &FragmentResource{
			ID:    "resource/A",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/B",
					Order:       1,
				},
			},
		}

		resourceB := &FragmentResource{
			ID:    "resource/B",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "name",
					Predicate:   "http://xmlns.com/foaf/0.1/name",
					EntryType:   literal,
					Value:       "Person B",
					Order:       1,
				},
			},
		}

		rm := &ResourceMap{
			resources: map[string]*FragmentResource{
				"resource/A": resourceA,
				"resource/B": resourceB,
			},
		}

		// Should work normally - full expansion
		sr := resourceA.ToSemanticResource(rm)
		require.NotNil(t, sr)

		b := sr.Fields["knows"].(*SemanticResource)
		assert.Equal(t, "resource/B", b.ID)
		// B should be fully expanded with its name field
		assert.NotEmpty(t, b.Fields, "non-cycle nested resource should have all fields")
		assert.Equal(t, ReferenceNone, b.ReferenceType, "fully expanded resource should have no reference type")

		name := b.Fields["name"].(*SemanticLiteral)
		assert.Equal(t, "Person B", name.Value)
	})

	t.Run("shared resource - not a cycle", func(t *testing.T) {
		// A knows both B and C, and both B and C know D
		// This is not a cycle, D should be fully expanded in both paths
		resourceD := &FragmentResource{
			ID:    "resource/D",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "name",
					Predicate:   "http://xmlns.com/foaf/0.1/name",
					EntryType:   literal,
					Value:       "Person D",
					Order:       1,
				},
			},
		}

		resourceB := &FragmentResource{
			ID:    "resource/B",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/D",
					Order:       1,
				},
			},
		}

		resourceC := &FragmentResource{
			ID:    "resource/C",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/D",
					Order:       1,
				},
			},
		}

		resourceA := &FragmentResource{
			ID:    "resource/A",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/B",
					Order:       1,
				},
				{
					SearchLabel: "knows",
					Predicate:   "http://xmlns.com/foaf/0.1/knows",
					EntryType:   resourceType,
					ID:          "resource/C",
					Order:       2,
				},
			},
		}

		rm := &ResourceMap{
			resources: map[string]*FragmentResource{
				"resource/A": resourceA,
				"resource/B": resourceB,
				"resource/C": resourceC,
				"resource/D": resourceD,
			},
		}

		sr := resourceA.ToSemanticResource(rm)
		require.NotNil(t, sr)

		// A should have an array of two "knows" values
		knowsValues := sr.Fields["knows"].(*SemanticArray)
		require.Len(t, *knowsValues, 2)

		// Both B and C should be fully expanded
		b := (*knowsValues)[0].(*SemanticResource)
		c := (*knowsValues)[1].(*SemanticResource)

		// And D should be fully expanded in both paths since it's not a cycle
		dFromB := b.Fields["knows"].(*SemanticResource)
		dFromC := c.Fields["knows"].(*SemanticResource)

		// Both should have the name field (fully expanded)
		assert.NotEmpty(t, dFromB.Fields)
		assert.NotEmpty(t, dFromC.Fields)
		assert.Equal(t, "Person D", dFromB.Fields["name"].(*SemanticLiteral).Value)
		assert.Equal(t, "Person D", dFromC.Fields["name"].(*SemanticLiteral).Value)
	})

	t.Run("external reference - resource not in graph", func(t *testing.T) {
		resourceA := &FragmentResource{
			ID:    "resource/A",
			Types: []string{"Person"},
			Entries: []*ResourceEntry{
				{
					SearchLabel: "sameAs",
					Predicate:   "http://www.w3.org/2002/07/owl#sameAs",
					EntryType:   resourceType,
					ID:          "http://external.example.org/person/123",
					Order:       1,
				},
			},
		}

		rm := &ResourceMap{
			resources: map[string]*FragmentResource{
				"resource/A": resourceA,
				// Note: external resource is NOT in the map
			},
		}

		sr := resourceA.ToSemanticResource(rm)
		require.NotNil(t, sr)

		extRef := sr.Fields["sameAs"].(*SemanticResource)
		assert.Equal(t, "http://external.example.org/person/123", extRef.ID)
		assert.Empty(t, extRef.Fields, "external reference should only contain ID")
		assert.Equal(t, ReferenceExternal, extRef.ReferenceType, "external reference should have ReferenceType=external")
	})

	t.Run("reference type appears in JSON output", func(t *testing.T) {
		// Cycle reference
		cycleRef := &SemanticResource{
			ID:            "resource/A",
			ReferenceType: ReferenceCycle,
		}
		data, err := json.Marshal(cycleRef)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))
		assert.Equal(t, "cycle", m["hub3:referenceType"])

		// External reference
		extRef := &SemanticResource{
			ID:            "http://external.example.org/thing",
			ReferenceType: ReferenceExternal,
		}
		data, err = json.Marshal(extRef)
		require.NoError(t, err)

		require.NoError(t, json.Unmarshal(data, &m))
		assert.Equal(t, "external", m["hub3:referenceType"])

		// Fully expanded resource (no reference type in output)
		expanded := &SemanticResource{
			ID:   "resource/B",
			Type: []string{"Person"},
			Fields: map[string]SemanticValue{
				"name": &SemanticLiteral{Value: "Test"},
			},
		}
		data, err = json.Marshal(expanded)
		require.NoError(t, err)

		var expandedMap map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &expandedMap))
		_, hasRefType := expandedMap["hub3:referenceType"]
		assert.False(t, hasRefType, "fully expanded resource should not have hub3:referenceType")
	})
}