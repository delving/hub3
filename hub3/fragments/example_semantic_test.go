package fragments

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSemanticFormatExample(t *testing.T) {
	// Create a simple FragmentGraph
	fg := &FragmentGraph{
		Meta: &Header{
			OrgID:         "museum",
			Spec:          "artworks",
			HubID:         "museum_artwork_123",
			EntryURI:      "https://museum.org/artwork/123",
			NamedGraphURI: "https://museum.org/graph/artwork/123",
		},
		Resources: []*FragmentResource{
			{
				ID:    "https://museum.org/artwork/123",
				Types: []string{"http://www.europeana.eu/schemas/edm/ProvidedCHO"},
				Entries: []*ResourceEntry{
					{
						SearchLabel: "dc:title",
						Predicate:   "http://purl.org/dc/elements/1.1/title",
						Value:       "Sunflowers",
						Language:    "en",
						EntryType:   literal,
						Order:       1,
					},
					{
						SearchLabel: "dc:creator",
						Predicate:   "http://purl.org/dc/elements/1.1/creator",
						ID:          "https://museum.org/artist/456",
						EntryType:   resourceType,
						Order:       2,
						Inline: &FragmentResource{
							ID:    "https://museum.org/artist/456",
							Types: []string{"http://www.europeana.eu/schemas/edm/Agent"},
							Entries: []*ResourceEntry{
								{
									SearchLabel: "skos:prefLabel",
									Predicate:   "http://www.w3.org/2004/02/skos/core#prefLabel",
									Value:       "Vincent van Gogh",
									EntryType:   literal,
									Order:       1,
								},
								{
									SearchLabel: "edm:begin",
									Predicate:   "http://www.europeana.eu/schemas/edm/begin",
									Value:       "1853",
									DataType:    "xsd:integer",
									EntryType:   literal,
									Order:       2,
								},
							},
						},
					},
					{
						SearchLabel: "dcterms:created",
						Predicate:   "http://purl.org/dc/terms/created",
						Value:       "1888",
						DataType:    "xsd:integer",
						EntryType:   literal,
						Order:       3,
					},
					{
						SearchLabel: "dc:description",
						Predicate:   "http://purl.org/dc/elements/1.1/description",
						Value:       "A vibrant painting of sunflowers",
						Language:    "en",
						EntryType:   literal,
						Order:       4,
					},
					{
						SearchLabel: "dc:description",
						Predicate:   "http://purl.org/dc/elements/1.1/description",
						Value:       "Un tableau vibrant de tournesols",
						Language:    "fr",
						EntryType:   literal,
						Order:       5,
					},
				},
			},
		},
	}

	// Generate semantic view
	sv := fg.GenerateSemantic()
	require.NotNil(t, sv)

	// Marshal to JSON for visualization
	jsonData, err := json.MarshalIndent(sv, "", "  ")
	require.NoError(t, err)

	// Print the JSON output (only visible when running with -v flag)
	t.Logf("Semantic Format Output:\n%s", string(jsonData))

	// Verify the structure
	require.Equal(t, "https://museum.org/artwork/123", sv.ID)
	require.Contains(t, sv.Type, "http://www.europeana.eu/schemas/edm/ProvidedCHO")

	// Check context
	require.NotEmpty(t, sv.Context)
	require.Contains(t, sv.Context, "dc")
	require.Contains(t, sv.Context, "dcterms")
	require.Contains(t, sv.Context, "edm")
	require.Contains(t, sv.Context, "skos")

	// Check title field
	title, exists := sv.Fields["dc:title"]
	require.True(t, exists)
	typedTitle, ok := title.(*SemanticTypedLiteral)
	require.True(t, ok)
	require.Equal(t, "Sunflowers", typedTitle.Value)
	require.Equal(t, "en", typedTitle.Language)

	// Check creator field (nested resource)
	creator, exists := sv.Fields["dc:creator"]
	require.True(t, exists)
	creatorResource, ok := creator.(*SemanticResource)
	require.True(t, ok)
	require.Equal(t, "https://museum.org/artist/456", creatorResource.ID)
	require.Contains(t, creatorResource.Type, "http://www.europeana.eu/schemas/edm/Agent")

	// Check nested resource fields
	prefLabel, exists := creatorResource.Fields["skos:prefLabel"]
	require.True(t, exists)
	prefLabelLit, ok := prefLabel.(*SemanticLiteral)
	require.True(t, ok)
	require.Equal(t, "Vincent van Gogh", prefLabelLit.Value)

	// Check typed literal (integer)
	created, exists := sv.Fields["dcterms:created"]
	require.True(t, exists)
	createdTyped, ok := created.(*SemanticTypedLiteral)
	require.True(t, ok)
	require.Equal(t, 1888, createdTyped.Value) // Converted to integer
	require.Equal(t, "xsd:integer", createdTyped.Type)

	// Check multiple descriptions (language variants)
	descriptions, exists := sv.Fields["dc:description"]
	require.True(t, exists)
	descArray, ok := descriptions.(*SemanticArray)
	require.True(t, ok)
	require.Len(t, *descArray, 2)

	// Verify both language variants
	enDesc := (*descArray)[0].(*SemanticTypedLiteral)
	require.Equal(t, "A vibrant painting of sunflowers", enDesc.Value)
	require.Equal(t, "en", enDesc.Language)

	frDesc := (*descArray)[1].(*SemanticTypedLiteral)
	require.Equal(t, "Un tableau vibrant de tournesols", frDesc.Value)
	require.Equal(t, "fr", frDesc.Language)

	fmt.Println("\nExample semantic format test completed successfully!")
}