package index

import (
	"testing"
)

func TestGraph_ResolveLabels(t *testing.T) {
	const (
		dcTitle      = "http://purl.org/dc/elements/1.1/title"
		skosPref     = "http://www.w3.org/2004/02/skos/core#prefLabel"
		rdfsLabel    = "http://www.w3.org/2000/01/rdf-schema#label"
		foafName     = "http://xmlns.com/foaf/0.1/name"
		foafKnows    = "http://xmlns.com/foaf/0.1/knows"
		dcCreator    = "http://purl.org/dc/elements/1.1/creator"
		dcSubject    = "http://purl.org/dc/elements/1.1/subject"
		dcRelation   = "http://purl.org/dc/elements/1.1/relation"
		dcDescription = "http://purl.org/dc/elements/1.1/description"
	)

	tests := []struct {
		name           string
		resources      []*Resource
		targetRscIdx   int    // index of resource to check
		targetEntryIdx int    // index of entry to check within that resource
		wantValue      string
		wantLang       string
		wantResolved   string // expected ResolvedFrom
		wantLevel      int32  // expected ResolvedLevel
	}{
		{
			name: "basic resolution via dc:title",
			resources: []*Resource{
				{
					ID: "http://example.org/record/1",
					Entries: []*Entry{
						{
							ID:        "http://example.org/person/1",
							Predicate: dcCreator,
							EntryType: ResourceType,
						},
					},
				},
				{
					ID: "http://example.org/person/1",
					Entries: []*Entry{
						{Predicate: dcTitle, Value: "John Doe", EntryType: Literal},
					},
					Context: []*ContextRef{{Level: 1}},
				},
			},
			targetRscIdx:   0,
			targetEntryIdx: 0,
			wantValue:      "John Doe",
			wantLang:       "",
			wantResolved:   dcTitle,
			wantLevel:      2, // GetLevel returns highestLevel(1) + 1 = 2
		},
		{
			name: "priority order: skos:prefLabel preferred over rdfs:label",
			resources: []*Resource{
				{
					ID: "http://example.org/record/1",
					Entries: []*Entry{
						{
							ID:        "http://example.org/concept/1",
							Predicate: dcSubject,
							EntryType: ResourceType,
						},
					},
				},
				{
					ID: "http://example.org/concept/1",
					Entries: []*Entry{
						{Predicate: rdfsLabel, Value: "Generic Label", EntryType: Literal},
						{Predicate: skosPref, Value: "Preferred Label", EntryType: Literal},
					},
					Context: []*ContextRef{{Level: 1}},
				},
			},
			targetRscIdx:   0,
			targetEntryIdx: 0,
			wantValue:      "Preferred Label",
			wantLang:       "",
			wantResolved:   skosPref,
			wantLevel:      2,
		},
		{
			name: "no overwrite when Value already set",
			resources: []*Resource{
				{
					ID: "http://example.org/record/1",
					Entries: []*Entry{
						{
							ID:        "http://example.org/person/1",
							Predicate: dcCreator,
							EntryType: ResourceType,
							Value:     "Existing Value",
						},
					},
				},
				{
					ID: "http://example.org/person/1",
					Entries: []*Entry{
						{Predicate: dcTitle, Value: "Should Not Appear", EntryType: Literal},
					},
				},
			},
			targetRscIdx:   0,
			targetEntryIdx: 0,
			wantValue:      "Existing Value",
			wantLang:       "",
			wantResolved:   "",
			wantLevel:      0,
		},
		{
			name: "external reference: linked resource not in graph",
			resources: []*Resource{
				{
					ID: "http://example.org/record/1",
					Entries: []*Entry{
						{
							ID:        "http://example.org/missing/1",
							Predicate: dcRelation,
							EntryType: ResourceType,
						},
					},
				},
			},
			targetRscIdx:   0,
			targetEntryIdx: 0,
			wantValue:      "",
			wantLang:       "",
			wantResolved:   "",
			wantLevel:      0,
		},
		{
			name: "bnode resolution",
			resources: []*Resource{
				{
					ID: "http://example.org/record/1",
					Entries: []*Entry{
						{
							ID:        "_:b0",
							Predicate: dcCreator,
							EntryType: Bnode,
						},
					},
				},
				{
					ID: "_:b0",
					Entries: []*Entry{
						{Predicate: foafName, Value: "Jane Smith", EntryType: Literal},
					},
					Context: []*ContextRef{{Level: 1}},
				},
			},
			targetRscIdx:   0,
			targetEntryIdx: 0,
			wantValue:      "Jane Smith",
			wantLang:       "",
			wantResolved:   foafName,
			wantLevel:      2,
		},
		{
			name: "language preservation",
			resources: []*Resource{
				{
					ID: "http://example.org/record/1",
					Entries: []*Entry{
						{
							ID:        "http://example.org/place/1",
							Predicate: dcSubject,
							EntryType: ResourceType,
						},
					},
				},
				{
					ID: "http://example.org/place/1",
					Entries: []*Entry{
						{Predicate: skosPref, Value: "Amsterdam", Language: "nl", EntryType: Literal},
					},
					Context: []*ContextRef{{Level: 2}},
				},
			},
			targetRscIdx:   0,
			targetEntryIdx: 0,
			wantValue:      "Amsterdam",
			wantLang:       "nl",
			wantResolved:   skosPref,
			wantLevel:      3, // highestLevel(2) + 1 = 3
		},
		{
			name: "no label found: linked resource has no label predicates",
			resources: []*Resource{
				{
					ID: "http://example.org/record/1",
					Entries: []*Entry{
						{
							ID:        "http://example.org/thing/1",
							Predicate: dcRelation,
							EntryType: ResourceType,
						},
					},
				},
				{
					ID: "http://example.org/thing/1",
					Entries: []*Entry{
						{Predicate: dcDescription, Value: "Not a label predicate", EntryType: Literal},
					},
				},
			},
			targetRscIdx:   0,
			targetEntryIdx: 0,
			wantValue:      "",
			wantLang:       "",
			wantResolved:   "",
			wantLevel:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Graph{
				Resources: tt.resources,
			}

			g.ResolveLabels()

			entry := g.Resources[tt.targetRscIdx].Entries[tt.targetEntryIdx]

			if entry.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", entry.Value, tt.wantValue)
			}

			if entry.Language != tt.wantLang {
				t.Errorf("Language = %q, want %q", entry.Language, tt.wantLang)
			}

			if entry.ResolvedFrom != tt.wantResolved {
				t.Errorf("ResolvedFrom = %q, want %q", entry.ResolvedFrom, tt.wantResolved)
			}

			if entry.ResolvedLevel != tt.wantLevel {
				t.Errorf("ResolvedLevel = %d, want %d", entry.ResolvedLevel, tt.wantLevel)
			}
		})
	}
}
