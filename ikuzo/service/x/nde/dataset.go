package nde

import (
	"errors"
	"fmt"

	"github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/hub3/models"
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
	Context     string     `json:"@context,omitempty"`
	ID          string     `json:"@id,omitempty"`
	Type        string     `json:"@type,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Publisher   Agent      `json:"publisher,omitempty"`
	Dataset     []*Dataset `json:"dataset,omitempty"`
}

func (s *Service) getDataset(orgID, spec string) (*Dataset, error) {
	r, ok := s.lookUp[spec]
	if !ok {
		return nil, fmt.Errorf("dataset not found: %q", spec)
	}

	d := &Dataset{
		Context:               "https://schema.org/",
		ID:                    r.getDatasetURI(spec),
		Type:                  "Dataset",
		Creator:               r.GetAgent(),
		InLanguage:            r.DefaultLanguages,
		IncludedInDataCatalog: fmt.Sprintf("%s/id/datacatalog", r.RDFBaseURL),
		Keywords:              []string{},
		License:               r.DefaultLicense,
		MainEntityOfPage:      fmt.Sprintf(r.DatasetFmt, r.publisherURL(), spec),
		Name:                  spec,
		Publisher:             r.GetAgent(),
	}

	layoutISO := "2006-01-02"

	meta, err := ead.GetMeta(spec)
	if !errors.Is(err, ead.ErrFileNotFound) && meta != nil {
		d.DateCreated = meta.Created.Format(layoutISO)
		d.DateModified = meta.Updated.Format(layoutISO)
		d.DatePublished = meta.Updated.Format(layoutISO)
		d.Description = meta.Label
		d.Distribution = r.GetDistributions(spec, "ead")

		return d, nil
	}

	ds, err := models.GetDataSet(orgID, spec)
	if err != nil {
		return nil, err
	}

	if ds.RecordType == "" {
		ds.RecordType = "narthex"
	}

	d.DateCreated = ds.Created.Format(layoutISO)
	d.DateModified = ds.Modified.Format(layoutISO)
	d.DatePublished = ds.Modified.Format(layoutISO)
	d.Description = ds.Label
	d.Distribution = r.GetDistributions(spec, ds.RecordType)

	return d, nil
}

func (s *Service) getDatasets(orgID string) ([]string, error) {
	datasets := []string{}

	sets, err := models.ListDataSets(orgID)
	if err != nil {
		return datasets, err
	}

	for _, set := range sets {
		datasets = append(datasets, set.Spec)
	}

	return datasets, nil
}

func (s *Service) AddDatasets(orgID string, catalog *Catalog) error {
	datasets, err := s.getDatasets(orgID)
	if err != nil {
		return err
	}

	for _, spec := range datasets {
		dataset, err := s.getDataset(orgID, spec)
		if err != nil {
			return err
		}

		catalog.Dataset = append(catalog.Dataset, dataset)
	}

	return nil
}
