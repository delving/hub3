package nde

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/delving/hub3/hub3/ead"
)

type Agent struct {
	ID            string `json:"@id,omitempty"`
	Type          string `json:"@type,omitempty"`
	AlternateName string `json:"alternateName,omitempty"`
	Name          string `json:"name,omitempty"`
	SameAs        string `json:"sameAs,omitempty"`
}

type Distribution struct {
	Type           string `json:"@type,omitempty"`
	ContentSize    string `json:"contentSize,omitempty"`
	ContentURL     string `json:"contentUrl,omitempty"`
	DateModified   string `json:"dateModified,omitempty"`
	DatePublished  string `json:"datePublished,omitempty"`
	EncodingFormat string `json:"encodingFormat,omitempty"`
	Name           string `json:"name,omitempty"`
}

type DatasetLink struct {
	ID   string `json:"@id,omitempty"`
	Type string `json:"@type,omitempty"`
}

type Dataset struct {
	Context               string         `json:"@context,omitempty"`
	Type                  string         `json:"@type,omitempty"`
	ID                    string         `json:"@id,omitempty"`
	Name                  string         `json:"name,omitempty"`
	Description           string         `json:"description,omitempty"`
	License               string         `json:"license,omitempty"`
	DateCreated           string         `json:"dateCreated,omitempty"`
	DateModified          string         `json:"dateModified,omitempty"`
	DatePublished         string         `json:"datePublished,omitempty"`
	Keywords              []string       `json:"keywords,omitempty"`
	IncludedInDataCatalog string         `json:"includedInDataCatalog,omitempty"`
	InLanguage            []string       `json:"inLanguage,omitempty"`
	MainEntityOfPage      string         `json:"mainEntityOfPage,omitempty"`
	Publisher             Agent          `json:"publisher,omitempty"`
	Creator               Agent          `json:"creator,omitempty"`
	Distribution          []Distribution `json:"distribution,omitempty"`
}

type Catalog struct {
	Context     string        `json:"@context,omitempty"`
	ID          string        `json:"@id,omitempty"`
	Type        string        `json:"@type,omitempty"`
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Publisher   Agent         `json:"publisher,omitempty"`
	Dataset     []DatasetLink `json:"dataset,omitempty"`
}

func (s *Service) getDataset(spec string) (*Dataset, error) {
	r := s.cfg

	meta, err := ead.GetMeta(spec)
	if err != nil {
		return nil, err
	}

	layoutISO := "2006-01-02"

	// only support ead for now
	datasetType := "ead"

	d := &Dataset{
		Context:               "https://schema.org/",
		ID:                    r.getDatasetURI(spec),
		Type:                  "Dataset",
		Creator:               r.GetAgent(),
		DateCreated:           meta.Created.Format(layoutISO),
		DateModified:          meta.Updated.Format(layoutISO),
		DatePublished:         meta.Updated.Format(layoutISO),
		Description:           meta.Label,
		Distribution:          r.GetDistributions(spec, datasetType),
		InLanguage:            r.DefaultLanguages,
		IncludedInDataCatalog: fmt.Sprintf("%s/id/catalog", r.RDFBaseURL),
		Keywords:              []string{},
		License:               r.DefaultLicense,
		MainEntityOfPage:      fmt.Sprintf(r.DatasetFmt, r.publisherURL(), spec),
		Name:                  spec,
		Publisher:             r.GetAgent(),
	}

	return d, err
}

func (s *Service) getDatasets() ([]string, error) {
	datasets := []string{}

	// TODO(kiivihal): change with dataset API later
	dirs, err := ioutil.ReadDir(s.cfg.DataPath)
	if err != nil {
		return datasets, err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		datasets = append(datasets, dir.Name())
	}

	return datasets, nil
}

func (s *Service) AddShortDatasetLinks(catalog *Catalog) error {
	datasets, err := s.getDatasets()
	if err != nil {
		return err
	}

	for _, spec := range datasets {
		// TODO(kiivihal): for now only return EAD datasets
		// change with dataset later
		_, err := ead.GetMeta(spec)
		if err != nil {
			if errors.Is(err, ead.ErrFileNotFound) {
				continue
			}

			return err
		}

		catalog.Dataset = append(catalog.Dataset, DatasetLink{
			ID:   s.cfg.getDatasetURI(spec),
			Type: "Dataset",
		})
	}

	return nil
}
