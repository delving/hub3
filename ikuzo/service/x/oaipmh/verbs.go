package oaipmh

import "time"

// TimeFormat can be set to change default time.RFC3339 format
var TimeFormat = time.RFC3339

type Verb string

const (
	VerbListIdentifiers     Verb = "listIdentifiers"
	VerbListRecords         Verb = "listRecords"
	VerbListSets            Verb = "listSets"
	VerbGetRecord           Verb = "getRecord"
	VerbListMetadataFormats Verb = "listMetadataFormats"
	VerbIdentify            Verb = "identify"
)

// var validVerbs = map[Verb]
