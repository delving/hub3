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

package models

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	wp "github.com/gammazero/workerpool"
	"github.com/rs/zerolog/log"

	elastic "github.com/olivere/elastic/v7"
)

// DataSetRevisions holds the type-frequency data for each revision
type DataSetRevisions struct {
	Number      int `json:"revisionNumber"`
	RecordCount int `json:"recordCount"`
}

// DataSetCounter holds value counters for statistics overviews
type DataSetCounter struct {
	Value    string `json:"value"`
	DocCount int    `json:"docCount"`
}

// IndexStats hold all Index Statistics for this dataset
type IndexStats struct {
	Enabled        bool               `json:"enabled"`
	Revisions      []DataSetRevisions `json:"revisions"`
	IndexedRecords int                `json:"indexedRecords"`
	Tags           []DataSetCounter   `json:"tags"`
	ContentTags    []DataSetCounter   `json:"contentTags"`
}

// RDFStoreStats hold all the RDFStore Statistics for this dataset
type RDFStoreStats struct {
	Revisions    []DataSetRevisions `json:"revisions"`
	StoredGraphs int                `json:"storedGraphs"`
	Enabled      bool               `json:"enabled"`
}

// LODFragmentStats hold all the LODFragment stats for this dataset
type LODFragmentStats struct {
	Enabled         bool               `json:"enabled"`
	Revisions       []DataSetRevisions `json:"revisions"`
	StoredFragments int                `json:"storedFragments"`
	DataType        []DataSetCounter   `json:"dataType"`
	Language        []DataSetCounter   `json:"language"`
	Tags            []DataSetCounter   `json:"tags"`
}

// WebResourceStats gathers all the MediaManager information for this DataSet
type WebResourceStats struct {
	Enabled           bool `json:"enabled"`
	SourceItems       int  `json:"sourceItems"`
	ThumbnailsCreated int  `json:"thumbnailsCreated"`
	DeepZoomsCreated  int  `json:"deepZoomsCreated"`
	Missing           int  `json:"missing"`
}

// NarthexStats gathers all the record statistics from Narthex
type NarthexStats struct {
	Enabled        bool `json:"enabled"`
	SourceRecords  int  `json:"sourceRecords"`
	ValidRecords   int  `json:"validRecords"`
	InvalidRecords int  `json:"invalidRecords"`
}

// VocabularyEnrichmentStats gathers all counters for the SKOS based enrichment
type VocabularyEnrichmentStats struct {
	LiteralFields        []string `json:"literalFields"`
	TotalConceptsMapped  int      `json:"totalConceptsMapped"`
	UniqueConceptsMapped int      `json:"uniqueConceptsMapped"`
	VocabularyLinked     []string `json:"vocabularyLinked"`
}

// DataSetStats holds all gather statistics for a DataSet
type DataSetStats struct {
	Spec                      string `json:"spec"`
	CurrentRevision           int    `json:"currentRevision"`
	IndexStats                `json:"index"`
	RDFStoreStats             `json:"rdfStore"`
	LODFragmentStats          `json:"lodFragmentStats"`
	DaoStats                  `json:"daoStats"`
	WebResourceStats          `json:"webResourceStats"`
	NarthexStats              `json:"narthexStats"`
	VocabularyEnrichmentStats `json:"vocabularyEnrichmentStats"`
}

// DataSet contains all the known informantion for a hub3 metadata dataset
type DataSet struct {
	//MapToPrefix string    `json:"mapToPrefix"`
	Spec             string    `json:"spec" storm:"id,index,unique"`
	URI              string    `json:"uri" storm:"unique,index"`
	Revision         int       `json:"revision"` // revision is used to mark the latest version of ingested RDFRecords
	FragmentRevision int       `json:"fragmentRevision"`
	Modified         time.Time `json:"modified" storm:"index"`
	Created          time.Time `json:"created"`
	Deleted          bool      `json:"deleted"`
	OrgID            string    `json:"orgID"`
	Access           `json:"access" storm:"inline"`
	Tags             []string `json:"tags"`
	RecordType       string   `json:"recordType"` //
	Label            string   `json:"label"`
	Owner            string   `json:"owner"`
	Abstract         []string `json:"abstract"`
	Period           []string `json:"period"`
	Length           string   `json:"length"`
	Files            string   `json:"files"`
	Language         string   `json:"language"`
	Material         string   `json:"material"`
	ArchiveCreator   []string `json:"archiveCreator"`
	MetsFiles        int      `json:"metsFiles"`
	Description      string   `json:"description"`
	Clevels          int      `json:"clevels"`
	DaoStats         `json:"daoStats" storm:"inline"`
	Fingerprint      string `json:"fingerPrint"`
}

// Access determines the which types of access are enabled for this dataset
type Access struct {
	OAIPMH bool `json:"oaipmh"`
	Search bool `json:"search"`
	LOD    bool `json:"lod"`
}

// DaoStats holds the stats for EAD digital objects extracted from METS links.
type DaoStats struct {
	ExtractedLinks uint64         `json:"extractedLinks"`
	RetrieveErrors uint64         `json:"retrieveErrors"`
	DigitalObjects uint64         `json:"digitalObjects"`
	Errors         []string       `json:"errors"`
	UniqueLinks    uint64         `json:"uniqueLinks"`
	DuplicateLinks map[string]int `json:"duplicateLinks"`
}

// createDatasetURI creates a RDF uri for the dataset based Config RDF BaseUrl
func createDatasetURI(spec string) string {
	uri := fmt.Sprintf("%s/resource/dataset/%s", c.Config.RDF.BaseURL, spec)
	return uri
}

// NewDataset creates a new instance of a DataSet
func NewDataset(spec string) DataSet {
	now := time.Now()
	access := Access{
		OAIPMH: true,
		Search: true,
		LOD:    true,
	}
	dataset := DataSet{
		OrgID:    c.Config.OrgID,
		Spec:     spec,
		URI:      createDatasetURI(spec),
		Created:  now,
		Modified: now,
		Access:   access,
	}

	return dataset
}

// -------------- migrated ---------------------

// GetDataSet returns a DataSet object when found
func GetDataSet(spec string) (*DataSet, error) {
	var ds DataSet
	err := ORM().One("Spec", spec, &ds)

	return &ds, err
}

// CreateDataSet creates and returns a DataSet
func CreateDataSet(spec string) (*DataSet, bool, error) {
	ds := NewDataset(spec)
	ds.Revision = 1
	err := ds.Save()

	return &ds, true, err
}

// GetOrCreateDataSet returns a DataSet object from the Storm ORM.
// If none is present it will create one
func GetOrCreateDataSet(spec string) (*DataSet, bool, error) {
	ds, err := GetDataSet(spec)
	if err != nil {
		return CreateDataSet(spec)
	}
	return ds, false, err
}

// IncrementRevision bumps the latest revision of the DataSet
func (ds *DataSet) IncrementRevision() (*DataSet, error) {
	err := ORM().UpdateField(&DataSet{Spec: ds.Spec}, "Revision", ds.Revision+1)
	if err != nil {
		log.Warn().Err(err).Str("datasetID", ds.Spec).Msg("Unable to update field in dataset")

		return nil, err
	}

	freshDs, err := GetDataSet(ds.Spec)

	return freshDs, err
}

// ListDataSets returns an array of Datasets stored in Storm ORM
func ListDataSets() ([]DataSet, error) {
	var ds []DataSet
	err := ORM().AllByIndex("Spec", &ds)
	return ds, err
}

// Save saves the DataSet to BoltDB
func (ds DataSet) Save() error {
	ds.Modified = time.Now()
	return ORM().Save(&ds)
}

// Delete deletes the DataSet from BoltDB
func (ds DataSet) Delete(ctx context.Context, wp *wp.WorkerPool) error {
	if c.Config.ElasticSearch.Enabled || c.Config.RDF.RDFStoreEnabled {
		_, err := ds.DropAll(ctx, wp)
		return err
	}

	log.Info().
		Str("component", "hub3").
		Str("datasetID", ds.Spec).
		Str("svc", "dataset").
		Msg("deleting dataset")

	return ORM().DeleteStruct(&ds)
}

// NewDataSetHistogram returns a histogram for dates that items in the index are modified
func NewDataSetHistogram() ([]*elastic.AggregationBucketHistogramItem, error) {
	ctx := context.Background()
	specAgg := elastic.NewTermsAggregation().Field("meta.spec").Size(100).OrderByCountDesc()
	agg := elastic.NewDateHistogramAggregation().Field("meta.modified").Format("yyyy-MM-dd").Interval("1D").
		SubAggregation("spec", specAgg)
	q := elastic.NewMatchAllQuery()

	res, err := index.ESClient().Search().
		Index(c.Config.ElasticSearch.GetIndexName()).
		TrackTotalHits(c.Config.ElasticSearch.TrackTotalHits).
		Query(q).
		Size(0).
		Aggregation("modified", agg).
		Do(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to render Modified histogram")

		return nil, err
	}

	aggMod, _ := res.Aggregations.DateHistogram("modified")

	return aggMod.Buckets, nil
}

// indexRecordRevisionsBySpec counts all the records stored in the Index for a Dataset
func (ds DataSet) indexRecordRevisionsBySpec(ctx context.Context) (int, []DataSetRevisions, []DataSetCounter, []DataSetCounter, error) {
	revisions := []DataSetRevisions{}
	counter := []DataSetCounter{}
	tagCounter := []DataSetCounter{}

	if !c.Config.ElasticSearch.Enabled {
		err := fmt.Errorf("indexRecordRevisionsBySpec should not be called when elasticsearch is not enabled")
		return 0, revisions, counter, tagCounter, err
	}

	revisionAgg := elastic.NewTermsAggregation().Field("meta.revision").Size(30).OrderByCountDesc()
	tagAgg := elastic.NewTermsAggregation().Field("meta.tags").Size(30).OrderByCountDesc()

	labelAgg := elastic.NewTermsAggregation().Field("resources.entries.tags").Size(30).OrderByCountDesc()
	contentTagAgg := elastic.NewNestedAggregation().Path("resources.entries")
	contentTagAgg = contentTagAgg.SubAggregation("contentTags", labelAgg)
	//contentTagAgg := elastic.NewTermsAggregation().Field("resource.entries.tags").Size(30).OrderByCountDesc()

	q := elastic.NewBoolQuery()
	q = q.Must(
		elastic.NewMatchPhraseQuery(c.Config.ElasticSearch.SpecKey, ds.Spec),
		elastic.NewTermQuery("meta.docType", fragments.FragmentGraphDocType),
		elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, c.Config.OrgID),
	)

	res, err := index.ESClient().Search().
		Index(c.Config.ElasticSearch.GetIndexName()).
		TrackTotalHits(c.Config.ElasticSearch.TrackTotalHits).
		Query(q).
		Size(0).
		Aggregation("revisions", revisionAgg).
		Aggregation("tags", tagAgg).
		Aggregation("contentTags", contentTagAgg).
		Do(ctx)
	if err != nil {
		log.Warn().Msgf("Unable to get IndexRevisionStats for the dataset: %s", err)
		return 0, revisions, counter, tagCounter, err
	}

	log.Info().Msgf("total hits: %d\n", res.Hits.TotalHits.Value)

	if res == nil {
		log.Warn().Msgf("expected response != nil; got: %v", res)
		return 0, revisions, counter, tagCounter, fmt.Errorf("expected response != nil")
	}
	aggs := res.Aggregations

	revAggCount, found := aggs.Terms("revisions")
	if !found {
		log.Warn().Msgf("Expected to find revision aggregations but got: %v", res)
		return 0, revisions, counter, tagCounter, fmt.Errorf("expected revision aggregrations")
	}

	for _, keyCount := range revAggCount.Buckets {
		revisions = append(revisions, DataSetRevisions{
			Number:      int(keyCount.Key.(float64)),
			RecordCount: int(keyCount.DocCount),
		})
	}

	counter, err = createDataSetCounters(aggs, "tags")
	if err != nil {
		log.Warn().Msgf("Unable to get Tag ggregations but got: %v", res)
		return 0, revisions, counter, tagCounter, fmt.Errorf("expected tag aggregrations")
	}

	// contentTags
	ct, _ := aggs.Nested("contentTags")
	ctt, _ := ct.Terms("contentTags")

	for _, keyCount := range ctt.Buckets {

		tagCounter = append(tagCounter, DataSetCounter{
			Value:    fmt.Sprintf("%s", keyCount.Key),
			DocCount: int(keyCount.DocCount),
		})
	}

	totalHits := res.Hits.TotalHits.Value

	return int(totalHits), revisions, counter, tagCounter, err
}

// createDataSetCounters creates counters from an ElasticSearch aggregation
func createDataSetCounters(aggs elastic.Aggregations, name string) ([]DataSetCounter, error) {
	counters := []DataSetCounter{}

	aggCount, found := aggs.Terms(name)
	if !found {
		log.Warn().Msgf("Expected to find %s aggregations but got: %v", name, aggs)
		return counters, fmt.Errorf("expected %s aggregrations", name)
	}

	for _, keyCount := range aggCount.Buckets {
		counters = append(counters, DataSetCounter{
			Value:    fmt.Sprintf("%s", keyCount.Key),
			DocCount: int(keyCount.DocCount),
		})
	}
	return counters, nil
}

// createLodFragmentStats queries the Fragment Store and returns LODFragmentStats struct
func (ds DataSet) createLodFragmentStats(ctx context.Context) (LODFragmentStats, error) {
	revisions := []DataSetRevisions{}
	fStats := LODFragmentStats{Enabled: c.Config.ElasticSearch.Fragments}

	if !c.Config.ElasticSearch.Enabled {
		return fStats, fmt.Errorf("FragmentStatsBySpec should not be called when elasticsearch is not enabled")
	}

	revisionAgg := elastic.NewTermsAggregation().Field("meta.revision").Size(30).OrderByCountDesc()
	languageAgg := elastic.NewTermsAggregation().Field("language").Size(50).OrderByCountDesc()
	dataType := elastic.NewTermsAggregation().Field("dataType").Size(50).OrderByCountDesc()
	tagsAgg := elastic.NewTermsAggregation().Field("meta.tags").Size(50).OrderByCountDesc()
	q := elastic.NewBoolQuery()
	q = q.Must(
		elastic.NewMatchPhraseQuery(c.Config.ElasticSearch.SpecKey, ds.Spec),
		elastic.NewTermQuery("meta.docType", fragments.FragmentDocType),
		elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, c.Config.OrgID),
	)
	res, err := index.ESClient().Search().
		Index(c.Config.ElasticSearch.FragmentIndexName()).
		TrackTotalHits(c.Config.ElasticSearch.TrackTotalHits).
		Query(q).
		Size(0).
		Aggregation("revisions", revisionAgg).
		Aggregation("language", languageAgg).
		Aggregation("dataType", dataType).
		Aggregation("tags", tagsAgg).
		Do(ctx)
	if err != nil {
		log.Warn().Msgf("Unable to get FragmentStatsBySpec for the dataset: %s", ds.Spec)
		return fStats, err
	}
	log.Info().Msgf("total hits: %d\n", res.Hits.TotalHits.Value)

	if res == nil {
		log.Warn().Msgf("expected response != nil; got: %v", res)
		return fStats, fmt.Errorf("expected response != nil")
	}

	aggs := res.Aggregations

	revAggCount, found := aggs.Terms("revisions")
	if !found {
		log.Warn().Msgf("Expected to find revision aggregations but got: %v", res)
		return fStats, fmt.Errorf("expected revision aggregrations")
	}

	for _, keyCount := range revAggCount.Buckets {
		revisions = append(revisions, DataSetRevisions{
			Number:      int(keyCount.Key.(float64)),
			RecordCount: int(keyCount.DocCount),
		})
	}

	buckets := []string{
		"language", "dataType",
		"tags",
	}
	for _, a := range buckets {
		counter, err := createDataSetCounters(aggs, a)
		if err != nil {
			return fStats, err
		}

		switch a {
		case "language":
			fStats.Language = counter
		case "dataType":
			fStats.DataType = counter
		case "tags":
			fStats.Tags = counter
		}
	}

	fStats.StoredFragments = int(res.Hits.TotalHits.Value)
	fStats.Revisions = revisions
	return fStats, nil
}

func getCount(counter []DataSetCounter) int {
	links := 0
	if len(counter) == 1 {
		links = counter[0].DocCount
	}
	return links
}

// createIndexStats queries ElasticSearch and returns the IndexStats struct
func (ds DataSet) createIndexStats(ctx context.Context) (IndexStats, error) {
	if !c.Config.ElasticSearch.Enabled {
		return IndexStats{Enabled: false}, nil
	}

	hits, indexRevisionCount, tags, contentTags, err := ds.indexRecordRevisionsBySpec(ctx)
	if err != nil {
		log.Warn().Msgf("Unable to get Index Revisions from ElasticSearch.")
		return IndexStats{}, err
	}
	return IndexStats{
		Revisions:      indexRevisionCount,
		Enabled:        c.Config.ElasticSearch.Enabled,
		IndexedRecords: hits,
		Tags:           tags,
		ContentTags:    contentTags,
	}, nil
}

// createRDFStoreStats queries the RDFstore and returns the RDFStoreStats struct
func (ds DataSet) createRDFStoreStats() (RDFStoreStats, error) {
	if !c.Config.RDF.RDFStoreEnabled {
		return RDFStoreStats{Enabled: false}, nil
	}

	storedGraphs, err := CountGraphsBySpec(ds.Spec)
	if err != nil {
		return RDFStoreStats{}, err
	}
	revisionCount, err := CountRevisionsBySpec(ds.Spec)
	if err != nil {
		return RDFStoreStats{}, err
	}
	return RDFStoreStats{
		Enabled:      c.Config.RDF.RDFStoreEnabled,
		Revisions:    revisionCount,
		StoredGraphs: storedGraphs,
	}, nil
}

// CreateDataSetStats returns DataSetStats that contain all relevant counts from the storage layer
func CreateDataSetStats(ctx context.Context, spec string) (DataSetStats, error) {
	ds, err := GetDataSet(spec)
	if err != nil {
		log.Warn().Msgf("Unable to retrieve dataset %s: %s", spec, err)
		return DataSetStats{}, err
	}
	indexStats, err := ds.createIndexStats(ctx)
	if err != nil {
		log.Warn().Msgf("Unable to create indexStats for %s; %#v", spec, err)
		return DataSetStats{}, err
	}
	storeStats, err := ds.createRDFStoreStats()
	if err != nil {
		log.Warn().Msgf("Unable to create rdfStoreStats for %s; %#v", spec, err)
		return DataSetStats{}, err
	}
	lodFragmentStats, err := ds.createLodFragmentStats(ctx)
	if err != nil {
		log.Warn().Msgf("Unable to create LODFragmentStats for %s; %#v", spec, err)
		return DataSetStats{}, err
	}
	return DataSetStats{
		Spec:             spec,
		IndexStats:       indexStats,
		RDFStoreStats:    storeStats,
		LODFragmentStats: lodFragmentStats,
		CurrentRevision:  ds.Revision,
		DaoStats:         ds.DaoStats,
	}, nil
}

// DeleteGraphsOrphans deletes all the orphaned graphs from the Triple Store linked to this dataset
func (ds DataSet) deleteGraphsOrphans() (bool, error) {
	return DeleteGraphsOrphansBySpec(ds.Spec, ds.Revision)
}

// DeleteAllGraphs deletes all the graphs linked to this dataset
func (ds DataSet) deleteAllGraphs() (bool, error) {
	return DeleteAllGraphsBySpec(ds.Spec)
}

// DeleteIndexOrphans deletes all the Orphaned records from the Search Index linked to this dataset
func (ds DataSet) deleteIndexOrphans(ctx context.Context, wp *wp.WorkerPool) (int, error) {

	v2 := elastic.NewBoolQuery()
	v2 = v2.MustNot(elastic.NewMatchQuery(c.Config.ElasticSearch.RevisionKey, ds.Revision))
	v2 = v2.Must(elastic.NewTermQuery(c.Config.ElasticSearch.SpecKey, ds.Spec))
	v2 = v2.Must(elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, c.Config.OrgID))

	v1 := elastic.NewBoolQuery()
	v1 = v1.MustNot(elastic.NewMatchQuery("revision", ds.Revision))
	v1 = v1.Must(elastic.NewTermQuery("spec.raw", ds.Spec))
	v1 = v1.Must(elastic.NewTermQuery("orgID", c.Config.OrgID))

	queries := map[*elastic.BoolQuery][]string{
		v1: []string{c.Config.ElasticSearch.GetV1IndexName()},
		v2: []string{
			c.Config.ElasticSearch.GetIndexName(),
			c.Config.ElasticSearch.FragmentIndexName(),
		},
	}

	go func() {
		// block for 15 seconds to allow cluster to be in sync
		timer := time.NewTimer(time.Second * 15)
		<-timer.C
		//log.Print("Orphan wait timer expired")

		for q, indices := range queries {
			res, err := index.ESClient().DeleteByQuery().
				Index(indices...).
				Query(q).
				Conflicts("proceed"). // default is abort
				Do(ctx)
			if err != nil {
				log.Warn().Msgf("Unable to delete orphaned dataset records from index: %s.", err)
				return
			}

			if res == nil {
				log.Warn().Msgf("expected response != nil; got: %v", res)
				return
			}

			log.Warn().Msgf(
				"Removed %d records for spec %s in index %s with older revision than %d",
				res.Deleted,
				indices,
				ds.Spec,
				ds.Revision,
			)
		}
	}()

	return 0, nil
}

// DeleteAllIndexRecords deletes all the records from the Search Index linked to this dataset
func (ds DataSet) deleteAllIndexRecords(ctx context.Context, wp *wp.WorkerPool) (int, error) {
	q := elastic.NewBoolQuery().Should(
		elastic.NewTermQuery(c.Config.ElasticSearch.SpecKey, ds.Spec),
		elastic.NewTermQuery("spec.raw", ds.Spec),
	)

	log.Warn().Msgf("%#v", q)
	res, err := index.ESClient().DeleteByQuery().
		Index(
			c.Config.ElasticSearch.GetIndexName(),
			c.Config.ElasticSearch.GetV1IndexName(),
			c.Config.ElasticSearch.FragmentIndexName(),
		).
		Query(q).
		Do(ctx)
	if err != nil {
		log.Warn().Msgf("Unable to delete dataset records from index.")
		return 0, err
	}

	if res == nil {
		log.Warn().Msgf("expected response != nil; got: %v", res)
		return 0, fmt.Errorf("expected response != nil")
	}

	log.Warn().Msgf("Removed %d records for spec %s", res.Deleted, ds.Spec)

	return int(res.Deleted), err
}

//DropOrphans removes all records of different revision that the current from the attached datastores
func (ds DataSet) DropOrphans(ctx context.Context, p *elastic.BulkProcessor, wp *wp.WorkerPool) (bool, error) {
	ok := true

	// TODO(kiivihal): replace flush with TRS
	// err := p.Flush()
	// if err != nil {
	// log.Warn().Msgf("Unable to Flush ElasticSearch index before deleting orphans.")
	// return false, err
	// }
	// log.Warn().Msgf("Flushed remaining items on the index queue.")

	if c.Config.RDF.RDFStoreEnabled {
		ok, err := ds.deleteGraphsOrphans()
		if !ok || err != nil {
			log.Warn().Msgf("Unable to remove RDF orphan graphs from spec %s: %s", ds.Spec, err)
			return false, err
		}
	}
	if c.Config.ElasticSearch.Enabled {
		_, err := ds.deleteIndexOrphans(ctx, wp)
		if err != nil {
			log.Warn().Msgf("Unable to remove RDF orphan graphs from spec %s: %s", ds.Spec, err)
			return false, err
		}
	}
	return ok, nil
}

// DropRecords Drops all records linked to the dataset from the storage layers
func (ds DataSet) DropRecords(ctx context.Context, wp *wp.WorkerPool) (bool, error) {
	var err error
	ok := true
	if c.Config.RDF.RDFStoreEnabled {
		ok, err = ds.deleteAllGraphs()
		if !ok || err != nil {
			log.Warn().Msgf("Unable to drop all graphs for %s", ds.Spec)
			return ok, err
		}
	}
	// todo add deleting all records from elastic search
	if c.Config.ElasticSearch.Enabled {
		_, err = ds.deleteAllIndexRecords(ctx, wp)
		if err != nil {
			log.Warn().Msgf("Unable to drop all index records for %s: %#v", ds.Spec, err)
			return false, err
		}
	}
	return ok, err
}

// DropAll drops the dataset from the Hub3 storages completely (BoltDB, Triple Store, Search Index)
func (ds DataSet) DropAll(ctx context.Context, wp *wp.WorkerPool) (bool, error) {
	ok, err := ds.DropRecords(ctx, wp)
	if !ok || err != nil {
		log.Warn().Msgf("Unable to drop all records for spec %s: %#v", ds.Spec, err)
		return ok, err
	}
	err = ORM().DeleteStruct(&ds)
	if err != nil {
		log.Warn().Msgf("Unable to delete dataset %s from storage", ds.Spec)
		return false, err
	}

	cachePath := filepath.Join(c.Config.EAD.CacheDir, ds.Spec)

	err = os.RemoveAll(cachePath)
	if err != nil {
		return false, fmt.Errorf("unable to delete EAD cache at %s; %#w", cachePath, err)
	}

	return ok, err
}
