package sparql

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestHarvest(t *testing.T) {
	is := is.New(t)

	cfg := RepoConfig{
		Host:      "https://api.kleksi.com",
		QueryPath: "/apps/thorn/datasets/default/sparql",
	}
	repo, err := NewRepo(cfg)
	is.NoErr(err)

	query := `PREFIX%20rdf%3A%20%3Chttp%3A%2F%2Fwww.w3.org%2F1999%2F02%2F22-rdf-syntax-ns%23%3E%0APREFIX%20rdfs%3A%20%3Chttp%3A%2F%2Fwww.w3.org%2F2000%2F01%2Frdf-schema%23%3E%0ASELECT%20*%20WHERE%20%7B%0A%20%20%3Fs%20a%20%20%3Chttps%3A%2F%2Fschema.org%2FCreativeWork%3E.%0A%7D%20LIMIT%2010`
	responses, err := Harvest(context.TODO(), repo, query)
	is.NoErr(err)
	_ = responses
}
