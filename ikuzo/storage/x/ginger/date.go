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

var dateFields = map[string]bool{
	ns.dcterms.Get("created").RawValue():       true,
	ns.dcterms.Get("issued").RawValue():        true,
	ns.nave.Get("creatorBirthYear").RawValue(): true,
	ns.nave.Get("creatorDeathYear").RawValue(): true,
	ns.nave.Get("date").RawValue():             true,
	ns.dc.Get("date").RawValue():               true,
	ns.nave.Get("dateOfBurial").RawValue():     true,
	ns.nave.Get("dateOfDeath").RawValue():      true,
	ns.nave.Get("productionEnd").RawValue():    true,
	ns.nave.Get("productionStart").RawValue():  true,
	ns.nave.Get("productionPeriod").RawValue(): true,
	ns.rdagr2.Get("dateOfBirth").RawValue():    true,
	ns.rdagr2.Get("dateOfDeath").RawValue():    true,
}

func reverseDates(date string) (string, error) {
	t, err := dateparse.ParseLocal(date)
	if err != nil {
		return "", err
	}

	// TODO(kiivihal): check if to support forward slash
	// cleanDate := strings.ReplaceAll(date, "-", "/")

	return t.Format("2006-01-02"), nil
}

func cleanDateURI(uri string) string {
	return fmt.Sprintf("%sRaw", uri)
}
