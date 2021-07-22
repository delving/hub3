package nde

import (
	"fmt"
	"log"
	"strings"
)

type DistributionCfg struct {
	DatasetType string `json:"datasetType"`
	MimeType    string `json:"mimeType"`
	DownloadFmt string `json:"downloadFmt"`
}

type RegisterConfig struct {
	RDFBaseURL       string
	Description      string
	Name             string
	DefaultLicense   string
	DefaultLanguages []string
	Publisher        struct {
		Name    string
		AltName string
		URL     string
	}
	DataPath      string // for ead support for now. Replace with dataset service later
	DatasetFmt    string
	Distributions []DistributionCfg
}

func (r *RegisterConfig) publisherURL() string {
	return strings.TrimSuffix(r.Publisher.URL, "/")
}

func (r *RegisterConfig) getDatasetURI(datasetID string) string {
	return fmt.Sprintf("%s/id/dataset/%s", r.RDFBaseURL, datasetID)
}

func (r *RegisterConfig) GetAgent() Agent {
	return Agent{
		ID:            r.publisherURL(),
		Type:          "Organization",
		AlternateName: r.Publisher.AltName,
		Name:          r.Publisher.Name,
		SameAs:        "",
	}
}

func (r *RegisterConfig) GetDistributions(spec, datasetType string) []Distribution {
	distributions := []Distribution{}

	log.Printf("%#v", r.Distributions)

	for _, cfg := range r.Distributions {
		if !strings.EqualFold(cfg.DatasetType, datasetType) {
			continue
		}

		distributions = append(distributions, Distribution{
			Type:           "DataDownload",
			ContentURL:     fmt.Sprintf(cfg.DownloadFmt, r.publisherURL(), spec),
			EncodingFormat: cfg.MimeType,
		})
	}

	return distributions
}

func (r *RegisterConfig) newCatalog() *Catalog {
	c := &Catalog{
		Context:     "https://schema.org/",
		ID:          fmt.Sprintf("%s/id/datacatalog", r.RDFBaseURL),
		Type:        "DataCatalog",
		Dataset:     []*Dataset{},
		Description: r.Description,
		Name:        r.Name,
		Publisher:   r.GetAgent(),
	}

	return c
}
