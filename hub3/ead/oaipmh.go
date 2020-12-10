package ead

import (
	"github.com/kiivihal/goharvest/oai"
	"github.com/rs/zerolog/log"
)

func ProcessEadFromOai(r *oai.Response) {
	if r.Request.Verb != "listRecords" {
		log.Warn().Str("verb", r.Request.Verb).Msg("verb is not supported for getting ead records")
		return
	}

	for _, record := range r.ListRecords.Records {
		log.Info().
			Str("identifier", record.Header.Identifier).
			Strs("set", record.Header.SetSpec).
			Str("lastModified", record.Header.DateStamp).
			Str("status", record.Header.Status).
			Msg("ead file entry")
	}
}

func ProcessMetsFromOai(r *oai.Response) {
	if r.Request.Verb != "listIdentifiers" {
		log.Warn().Str("verb", r.Request.Verb).Msg("verb is not supported for getting mets headers")
		return
	}

	for _, header := range r.ListIdentifiers.Headers {
		log.Info().
			Str("identifier", header.Identifier).
			Strs("set", header.SetSpec).
			Str("lastModified", header.DateStamp).
			Str("status", header.Status).
			Msg("mets entry")
	}
}
