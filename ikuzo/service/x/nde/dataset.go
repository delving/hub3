package nde

import (
	"errors"
	"fmt"

	"github.com/asdine/storm/q"
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

func (s *Service) createDataSet(ds *models.DataSet) (*Dataset, error) {
	spec := ds.Spec

	if ds.RecordType == "" {
		ds.RecordType = "narthex"
	}

	r, ok := s.recordTypeLookup[ds.RecordType]
	if !ok {
		r = s.defaultCfg
	}

	d := &Dataset{
		Context:               "https://schema.org/",
		ID:                    r.getDatasetURI(spec),
		Type:                  "Dataset",
		Creator:               r.GetAgent(),
		InLanguage:            r.DefaultLanguages,
		IncludedInDataCatalog: fmt.Sprintf("%s/id/datacatalog/%s", r.RDFBaseURL, r.URLPrefix),
		Keywords:              []string{},
		License:               r.DefaultLicense,
		MainEntityOfPage:      fmt.Sprintf(r.DatasetFmt, r.publisherURL(), spec),
		Name:                  spec,
		Publisher:             r.GetAgent(),
	}

	layoutISO := "2006-01-02"

	if ds.RecordType == "ead" {
		meta, err := ead.GetMeta(spec)
		if !errors.Is(err, ead.ErrFileNotFound) && meta != nil {
			d.DateCreated = meta.Created.Format(layoutISO)
			d.DateModified = meta.Updated.Format(layoutISO)
			d.DatePublished = meta.Updated.Format(layoutISO)
			d.Description = meta.Label
			d.Distribution = r.GetDistributions(spec, "ead")

			return d, nil
		}
	}

	d.DateCreated = ds.Created.Format(layoutISO)
	d.DateModified = ds.Modified.Format(layoutISO)
	d.DatePublished = ds.Modified.Format(layoutISO)
	d.Description = ds.Label
	d.Distribution = r.GetDistributions(spec, ds.RecordType)

	return d, nil
}

func (s *Service) getDataset(orgID, spec string) (*Dataset, error) {
	ds, err := models.GetDataSet(orgID, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to find dataset: %s [%s]", spec, orgID)
	}

	d, err := s.createDataSet(ds)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (s *Service) getDatasets(orgID string, tag string) ([]*models.DataSet, error) {
	query := models.ORM().Select(q.And(
		q.Eq("OrgID", orgID),
		q.Eq("RecordType", tag),
	))

	var sets []*models.DataSet
	if err := query.Find(&sets); err != nil {
		return sets, fmt.Errorf("no sets match query %s %s; %w", orgID, tag, err)
	}

	return sets, nil
}

func (s *Service) AddDatasets(orgID string, catalog *Catalog, tag string) error {
	datasets, err := s.getDatasets(orgID, tag)
	if err != nil {
		return fmt.Errorf("unable to get datasets from store: %w", err)
	}

	for _, ds := range datasets {
		dataset, err := s.createDataSet(ds)
		if err != nil {
			return err
		}

		catalog.Dataset = append(catalog.Dataset, dataset)
	}

	return nil
}
