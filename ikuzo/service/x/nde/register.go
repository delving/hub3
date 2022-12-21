package nde

import (
	"fmt"
	"strings"
)

type DistributionCfg struct {
	DatasetType string `json:"datasetType"`
	MimeType    string `json:"mimeType"`
	DownloadFmt string `json:"downloadFmt"`
}

type RegisterConfig struct {
	Default          bool
	URLPrefix        string `json:"urlPrefix"`
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
	DataPath         string // for ead support for now. Replace with dataset service later
	DatasetFmt       string
	Distributions    []DistributionCfg
	RecordTypeFilter string
}

func (r *RegisterConfig) publisherURL() string {
	return strings.TrimSuffix(r.Publisher.URL, "/")
}

func (r *RegisterConfig) getDatasetURI(datasetID string) string {
	return fmt.Sprintf("%s/id/dataset/%s/%s", r.RDFBaseURL, r.URLPrefix, datasetID)
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

	for _, cfg := range r.Distributions {
		if !strings.EqualFold(cfg.DatasetType, datasetType) {
			continue
		}

		contentUrl := cfg.DownloadFmt
		fmtCnt := strings.Count(cfg.DownloadFmt, "%s")

		if fmtCnt > 0 {
			switch fmtCnt {
			case 1:
				contentUrl = fmt.Sprintf(cfg.DownloadFmt, spec)
			case 2:
				contentUrl = fmt.Sprintf(cfg.DownloadFmt, r.Publisher.URL, spec)
			}
		}

		distributions = append(distributions, Distribution{
			Type:           "DataDownload",
			ContentURL:     contentUrl,
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
