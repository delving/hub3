package ginger

import (
	"fmt"

	"github.com/araddon/dateparse"
	ld "github.com/linkeddata/gojsonld"
)

var (
	ns = struct {
		rdf, rdfs, acl, cert, foaf, stat, dc, dcterms, nave, rdagr2, edm ld.NS
	}{
		rdf:     ld.NewNS("http://www.w3.org/1999/02/22-rdf-syntax-ns#"),
		rdfs:    ld.NewNS("http://www.w3.org/2000/01/rdf-schema#"),
		acl:     ld.NewNS("http://www.w3.org/ns/auth/acl#"),
		cert:    ld.NewNS("http://www.w3.org/ns/auth/cert#"),
		foaf:    ld.NewNS("http://xmlns.com/foaf/0.1/"),
		stat:    ld.NewNS("http://www.w3.org/ns/posix/stat#"),
		dc:      ld.NewNS("http://purl.org/dc/elements/1.1/"),
		dcterms: ld.NewNS("http://purl.org/dc/terms/"),
		nave:    ld.NewNS("http://schemas.delving.eu/nave/terms/"),
		rdagr2:  ld.NewNS("http://rdvocab.info/ElementsGr2/"),
		edm:     ld.NewNS("http://www.europeana.eu/schemas/edm/"),
	}
)

var dateFields = []ld.Term{
	ns.dcterms.Get("created"),
	ns.dcterms.Get("issued"),
	ns.nave.Get("creatorBirthYear"),
	ns.nave.Get("creatorDeathYear"),
	ns.nave.Get("date"),
	ns.dc.Get("date"),
	ns.nave.Get("dateOfBurial"),
	ns.nave.Get("dateOfDeath"),
	ns.nave.Get("productionEnd"),
	ns.nave.Get("productionStart"),
	ns.nave.Get("productionPeriod"),
	ns.rdagr2.Get("dateOfBirth"),
	ns.rdagr2.Get("dateOfDeath"),
}

func reverseDates(date string) (string, error) {
	// cleanDate := strings.ReplaceAll(date, "-", "/")

	t, err := dateparse.ParseLocal(date)
	if err != nil {
		return "", err
	}

	// fmt.Printf("%s", t.Format("2006-01-02"))
	return t.Format("2006-01-02"), nil
}

func cleanDateURI(uri string) string {
	return fmt.Sprintf("%sRaw", uri)
}
