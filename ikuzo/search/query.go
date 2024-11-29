package search

type QueryType string

type BoolQuery struct {
	Must   []Query
	Should []Query
	Not    []Query
}

type Query struct {
	Nested    BoolQuery
	Field     string
	Value     string
	QueryType QueryType
}
