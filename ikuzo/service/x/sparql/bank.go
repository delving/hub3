package sparql

const queries = `
# SPARQL queries that are loaded as a QueryBank

# ask returns a boolean
# tag: ask_subject
ASK { <{{ .URI }}> ?p ?o }

# tag: ask_predicate
ASK { ?s <{{ .URI }}> ?o }

# tag: ask_object
ASK { ?s <{{ .URI }}> ?o }

# tag: ask_query
ASK { {{ .Query }} }

# The DESCRIBE form returns a single result RDF graph containing RDF data about resources.
# tag: describe
DESCRIBE <{{.URI}}>

# tag: countGraphPerSpec
SELECT (count(?subject) as ?count)
WHERE {
  ?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}"
}
LIMIT 1

# tag: countRevisionsBySpec
SELECT ?revision (COUNT(?revision) as ?rCount)
WHERE
{
  ?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}";
		<http://schemas.delving.eu/nave/terms/specRevision> ?revision .
}
GROUP BY ?revision

# tag: deleteAllGraphsBySpec
DELETE {
	GRAPH ?g {
	?s ?p ?o .
	}
}
WHERE {
	GRAPH ?g {
	?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}".
	}
	GRAPH ?g {
	?s ?p ?o .
	}
};

# tag: deleteOrphanGraphsBySpec
DELETE {
	GRAPH ?g {
	?s ?p ?o .
	}
}
WHERE {
	GRAPH ?g {
	?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}";
		<http://schemas.delving.eu/nave/terms/specRevision> ?revision .
		FILTER (?revision != {{.RevisionNumber}}).
	}
	GRAPH ?g {
	?s ?p ?o .
	}
};

# tag: countAllTriples
SELECT (count(?s) as ?count)
WHERE {
  ?s ?p ?o .
};

# tag: harvestTriples
SELECT *
WHERE {
  ?s ?p ?o .
} LIMIT {{.Limit}} OFFSET {{.Offset}}
`
