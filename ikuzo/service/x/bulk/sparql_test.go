package bulk

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestDiffTriples(t *testing.T) {
	t.Run("inserted and deleted", func(t *testing.T) {
		is := is.New(t)

		a := `<1> <dc:title> "title" .
	<1> <dc:subject> "subject".
	<1> <dc:subject> "subject2".
	<1> <dc:identifier> "123" .
	<1> <dc:isPartOf> <2> .
	`

		b := `<1> <dc:title> "title 2" .
	<1> <dc:subject> "subject".
	<1> <dc:identifier> "123" .
	<1> <dc:isPartOf> <2> .
	<1> <dc:subject> "subject3".
	`

		inserted, deleted := diffTriples(a, b)
		t.Logf("inserted: %#v", inserted)
		t.Logf("deleted: %#v", deleted)
		is.Equal(len(inserted), 2)
		is.Equal(len(deleted), 2)
		is.Equal(inserted, []string{"\t<1> <dc:subject> \"subject3\".", "<1> <dc:title> \"title 2\" ."})
		is.Equal(deleted, []string{"\t<1> <dc:subject> \"subject2\".", "<1> <dc:title> \"title\" ."})
	})
	t.Run("no changes", func(t *testing.T) {
		is := is.New(t)

		a := `<1> <dc:title> "title" .
	<1> <dc:subject> "subject".
	<1> <dc:subject> "subject2".
	<1> <dc:identifier> "123" .
	<1> <dc:isPartOf> <2> .
	`
		inserted, deleted := diffTriples(a, a)
		is.Equal(len(inserted), 0)
		is.Equal(len(deleted), 0)
	})
	t.Run("no changes", func(t *testing.T) {
		is := is.New(t)

		a := `<1> <dc:title> "title" .
	<1> <dc:subject> "subject".
	<1> <dc:subject> "subject2".
	<1> <dc:identifier> "123" .
	<1> <dc:isPartOf> <2> .
	`
		inserted, deleted := diffTriples(``, a)
		is.Equal(len(inserted), 5)
		is.Equal(len(deleted), 0)
	})
}

func TestDiffAsSparqlUpdate(t *testing.T) {
	is := is.New(t)
	a := `<1> <dc:title> "title" .
	<1> <dc:subject> "subject".
	<1> <dc:subject> "subject2".
	<1> <dc:identifier> "123" .
	<1> <dc:isPartOf> <2> .
	`

	b := `<1> <dc:title> "title 2" .
	<1> <dc:subject> "subject".
	<1> <dc:identifier> "123" .
	<1> <dc:isPartOf> <2> .
	<1> <dc:subject> "subject3".
	`

	want := "DELETE DATA {\nGRAPH <urn:123/graph> { \t<1> <dc:subject> \"subject2\".\n<1> <dc:title> \"title\" .\n }\n}\nINSERT DATA {\nGRAPH <urn:123/graph> { <urn:123/graph> <http://schemas.delving.eu/nave/terms/datasetSpec> \"123\" .\n\t<1> <dc:subject> \"subject3\".\n<1> <dc:title> \"title 2\" .\n }\n}\n"

	got, err := diffAsSparqlUpdate(a, b, "urn:123/graph", "123")
	t.Logf("query: %#v", got)
	is.NoErr(err)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("diffAsSparqlUpdate() mismatch (-want +got):\n%s", diff)
	}
}
