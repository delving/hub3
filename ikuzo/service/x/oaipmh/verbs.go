package oaipmh

import "time"

// TimeFormat can be set to change default time.RFC3339 format
var TimeFormat = time.RFC3339

type Verb string

const (
	VerbListIdentifiers     Verb = "ListIdentifiers"
	VerbListRecords         Verb = "ListRecords"
	VerbListSets            Verb = "ListSets"
	VerbGetRecord           Verb = "GetRecord"
	VerbListMetadataFormats Verb = "ListMetadataFormats"
	VerbIdentify            Verb = "Identify"
)

// var validVerbs = map[Verb]
