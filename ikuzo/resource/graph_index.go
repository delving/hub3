package resource

type index struct {
	dataTypes       map[*IRI]uint64
	objectResources map[*IRI]uint64
	languages       map[string]uint64
	predicates      map[*IRI]uint64
}

// GraphStats returns counts for unique values in Graph
type GraphStats struct {
	Languages      uint64
	ObjectIRIs     uint64
	ObjectLiterals uint64
	Predicates     uint64
	Resources      uint64
	Triples        uint64
}
