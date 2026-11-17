package rdf

import "strings"

const (
	schemaOrgHTTP  = "http://schema.org/"
	schemaOrgHTTPS = "https://schema.org/"
)

// NormaliseSchemaOrg rewrites http://schema.org/ IRIs in subject, predicate and
// object to https://schema.org/. schema.org accepts both; hub3 maps only https
// to the "schema" prefix, so http gave search labels like sorg_name and missed
// a schema:CreativeWork recDef.
func NormaliseSchemaOrg(t *Triple) *Triple {
	if s, ok := t.Subject.(IRI); ok {
		t.Subject = httpsSchemaOrg(s)
	}
	if p, ok := t.Predicate.(IRI); ok {
		t.Predicate = httpsSchemaOrg(p)
	}
	if o, ok := t.Object.(IRI); ok {
		t.Object = httpsSchemaOrg(o)
	}
	return t
}

func httpsSchemaOrg(iri IRI) IRI {
	if rest, ok := strings.CutPrefix(iri.str, schemaOrgHTTP); ok {
		return IRI{str: schemaOrgHTTPS + rest}
	}
	return iri
}
