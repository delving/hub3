package oaipmh

var (
	ErrBadVerb = Error{
		Code: "badVerb",
		Message: "Value of the verb argument is not a legal OAI-PMH verb, " +
			"the verb argument is missing, or the verb argument is repeated.",
	}

	ErrBadArgument = Error{
		Code: "badArgument",
		Message: "The request includes illegal arguments, is missing required arguments, " +
			"includes a repeated argument, or values for arguments have an illegal syntax.",
	}

	ErrBadResumptionToken = Error{
		Code:            "badResumptionToken",
		Message:         "The value of the resumptionToken argument is invalid or expired.",
		applicableVerbs: []Verb{VerbListIdentifiers, VerbListRecords, VerbListSets},
	}

	ErrCannotDisseminateFormat = Error{
		Code: "cannotDisseminateFormat",
		Message: "The metadata format identified by the value given for the " +
			"metadataPrefix argument is not supported by the item or by the repository.",
		applicableVerbs: []Verb{VerbGetRecord, VerbListIdentifiers, VerbListRecords},
	}

	ErrIDDoesNotExist = Error{
		Code: "idDoesNotExist",
		Message: "The value of the identifier argument is unknown or illegal " +
			"in this repository.",
		applicableVerbs: []Verb{VerbGetRecord, VerbListMetadataFormats},
	}

	ErrNoRecordsMatch = Error{
		Code: "noRecordsMatch",
		Message: "The combination of the values of the from, until, set " +
			"and metadataPrefix arguments results in an empty list.",
		applicableVerbs: []Verb{VerbListIdentifiers, VerbListRecords},
	}

	ErrNoMetadataFormats = Error{
		Code:            "noMetadataFormats",
		Message:         "There are no metadata formats available for the specified item.",
		applicableVerbs: []Verb{VerbListMetadataFormats},
	}

	ErrNoSetHierachy = Error{
		Code:            "noSetHierarchy",
		Message:         "The repository does not support sets.",
		applicableVerbs: []Verb{VerbListSets, VerbListRecords, VerbListIdentifiers},
	}
)
