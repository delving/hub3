// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package harvesting

import (
	"github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/models"
	"github.com/kiivihal/goharvest/oai"
)

// ProcessVerb processes different OAI-PMH verbs
func ProcessVerb(r *oai.Request) interface{} {
	switch r.Verb {
	case "Identify":
		return renderIdentify(r)
	case "ListMetadataFormats":
		// TODO add getter from list of record definitions
		formats := []oai.MetadataFormat{
			oai.MetadataFormat{
				MetadataPrefix:    "edm",
				Schema:            "",
				MetadataNamespace: "http://www.europeana.eu/schemas/edm/",
			},
		}

		return oai.ListMetadataFormats{
			MetadataFormat: formats,
		}
	case "ListSets":
		return renderListSets(r)
	case "ListIdentifiers":
		return "identifiers"
	case "ListRecords":
		return "records"
	case "GetRecord":
		return "record"
	default:
		badVerb := oai.OAIError{
			Code: "badVerb",
			Message: `Value of the verb argument is not a legal OAI-PMH verb,
			the verb argument is missing, or the verb argument is repeated.`,
		}
		return badVerb
	}
}

// renderIdentify returns the identify response of the repository
func renderIdentify(r *oai.Request) interface{} {
	return oai.Identify{
		RepositoryName:    config.Config.OAIPMH.RepositoryName,
		BaseURL:           r.BaseURL,
		ProtocolVersion:   "2.0",
		AdminEmail:        config.Config.OAIPMH.AdminEmails,
		DeletedRecord:     "persistent",
		EarliestDatestamp: "1970-01-01T00:00:00Z",
		Granularity:       "YYYY-MM-DDThh:mm:ssZ",
	}
}

// renderListSets returns a list of all the publicly available sets
func renderListSets(r *oai.Request) interface{} {
	sets := []oai.Set{}
	datasets, err := models.ListDataSets()
	if err != nil {
		logger.Errorln("Unable to retrieve datasets from the storage layer.")
		return sets
	}
	for _, ds := range datasets {
		if ds.Access.OAIPMH {
			sets = append(
				sets,
				oai.Set{
					SetSpec:        ds.Spec,
					SetName:        ds.Spec,                                // todo change to name if it has one later
					SetDescription: oai.Description{Body: []byte(ds.Spec)}, // TODO change to description from ds later.
				},
			)
		}
	}
	return oai.ListSets{
		Set: sets,
	}
}
