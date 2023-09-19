package nde

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/asdine/storm"
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
	Identifier            string         `json:"identifier,omitempty"`
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
	Context     []any      `json:"@context,omitempty"`
	ID          string     `json:"@id,omitempty"`
	Type        []string   `json:"@type,omitempty"`
	HydraView   *HydraView `json:"hydra:view,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Publisher   Agent      `json:"publisher,omitempty"`
	Dataset     []*Dataset `json:"dataset,omitempty"`
}

func (c *Catalog) addHydraView(currentPage string, totalSize int) error {
	if currentPage == "" {
		currentPage = "1"
	}

	page, err := strconv.Atoi(currentPage)
	if err != nil {
		return fmt.Errorf("unable to convert %q to integer", currentPage)
	}
	c.HydraView = &HydraView{
		baseID:      c.ID,
		Type:        "hydra:PartialCollectionView",
		TotalItems:  totalSize,
		currentPage: page,
	}

	c.HydraView.setPager()
	return nil
}

const hydraObjectPerPage = 500

type HydraView struct {
	ID          string            `json:"@id,omitempty"`
	Type        string            `json:"@type,omitempty"`
	First       map[string]string `json:"hydra:first,omitempty"`
	Next        map[string]string `json:"hydra:next,omitempty"`
	Last        map[string]string `json:"hydra:last,omitempty"`
	TotalItems  int               `json:"hydra:totalItems,omitempty"`
	baseID      string
	currentPage int
}

func (hv *HydraView) getBounds() (lower int, upper int) {
	lower = (hv.currentPage - 1) * hydraObjectPerPage
	upper = (hv.currentPage * hydraObjectPerPage) - 1
	return lower, upper
}

func (hv *HydraView) hydraPage(page int) string {
	return fmt.Sprintf("%s?page=%d", hv.baseID, page)
}

func (hv *HydraView) setPager() {
	if hv.currentPage == 0 {
		hv.currentPage = 1
	}
	hv.ID = hv.hydraPage(hv.currentPage)

	hv.First = map[string]string{"@id": hv.hydraPage(1)}
	next := hv.currentPage + 1
	totalPages := int(hv.TotalItems / hydraObjectPerPage)
	if next < totalPages {
		hv.Next = map[string]string{"@id": hv.hydraPage(next)}
	}
	if totalPages > 1 {
		hv.Last = map[string]string{"@id": hv.hydraPage(totalPages)}
	}
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
		Name:                  spec,
		Identifier:            spec,
		Publisher:             r.GetAgent(),
	}

	if r.DatasetFmt != "" {
		d.MainEntityOfPage = fmt.Sprintf(r.DatasetFmt, r.publisherURL(), spec)
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
			d.Name = meta.Label

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

func (s *Service) getDatasetsCount(orgID string, tag string) (int, error) {
	return s.datasetQuery(orgID, tag).Count(&models.DataSet{})
}

func (s *Service) datasetQuery(orgID string, tag string) storm.Query {
	return models.ORM().Select(q.And(
		q.Eq("OrgID", orgID),
		q.Eq("RecordType", tag),
	))
}

func (s *Service) getDatasets(orgID string, tag string) ([]*models.DataSet, error) {
	var sets []*models.DataSet
	if err := s.datasetQuery(orgID, tag).Find(&sets); err != nil {
		return sets, fmt.Errorf("no sets match query %s %s; %w", orgID, tag, err)
	}

	return sets, nil
}

func (s *Service) AddDatasets(orgID string, catalog *Catalog, tag string) error {
	datasets, err := s.getDatasets(orgID, tag)
	if err != nil {
		return fmt.Errorf("unable to get datasets from store: %w", err)
	}

	lower, upper := catalog.HydraView.getBounds()

	for idx, ds := range datasets {
		if idx >= lower && (idx <= upper || upper == 0) {
			dataset, err := s.createDataSet(ds)
			if err != nil {
				return err
			}
			catalog.Dataset = append(catalog.Dataset, dataset)
		}
	}

	return nil
}
