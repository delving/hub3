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
	"log"
	"time"

	c "bitbucket.org/delving/rapid/config"
	"bitbucket.org/delving/rapid/hub3/index"
	elastic "gopkg.in/olivere/elastic.v5"
)

// DataSetRevisions holds the type-frequency data for each revision
type DataSetRevisions struct {
	Number      int `json:"revisionNumber"`
	RecordCount int `json:"recordCount"`
}

// DataSetStats holds all gather statistics for a DataSet
type DataSetStats struct {
	Spec            string             `json:"spec"`
	StoredGraphs    int                `json:"storedGraphs"`
	IndexedRecords  int                `json:"indexedRecords"`
	CurrentRevision int                `json:"currentRevision"`
	GraphRevisions  []DataSetRevisions `json:"graphRevisions"`
	IndexRevisions  []DataSetRevisions `json:"indexRevisions"`
}

// DataSet contains all the known informantion for a RAPID metadata dataset
type DataSet struct {
	//MapToPrefix string    `json:"mapToPrefix"`
	Spec     string    `json:"spec" storm:"id,index,unique"`
	URI      string    `json:"uri" storm:"unique,index"`
	Revision int       `json:"revision"` // revision is used to mark the latest version of ingested RDFRecords
	Modified time.Time `json:"modified" storm:"index"`
	Created  time.Time `json:"created"`
	Deleted  bool      `json:"deleted"`
	Access   `json:"access" storm:"inline"`
}

// Access determines the which types of access are enabled for this dataset
type Access struct {
	OAIPMH bool `json:"oaipmh"`
	Search bool `json:"search"`
	LOD    bool `json:"lod"`
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
		Spec:     spec,
		URI:      createDatasetURI(spec),
		Created:  now,
		Modified: now,
		Access:   access,
	}
	return dataset
}

// GetDataSet returns a DataSet object when found
func GetDataSet(spec string) (DataSet, error) {
	var ds DataSet
	err := orm.One("Spec", spec, &ds)
	return ds, err
}

// CreateDataSet creates and returns a DataSet
func CreateDataSet(spec string) (DataSet, error) {
	ds := NewDataset(spec)
	err := ds.Save()
	return ds, err
}

// GetOrCreateDataSet returns a DataSet object from the Storm ORM.
// If none is present it will create one
func GetOrCreateDataSet(spec string) (DataSet, error) {
	ds, err := GetDataSet(spec)
	if err != nil {
		return CreateDataSet(spec)
	}
	return ds, err
}

// IncrementRevision bumps the latest revision of the DataSet
func (ds *DataSet) IncrementRevision() error {
	err := orm.UpdateField(&DataSet{Spec: ds.Spec}, "Revision", ds.Revision+1)
	if err != nil {
		log.Printf("Unable to update field in dataset: %s", ds.Spec)
		return err
	}
	freshDs, err := GetDataSet(ds.Spec)
	ds = &freshDs
	return err
}

// ListDataSets returns an array of Datasets stored in Storm ORM
func ListDataSets() ([]DataSet, error) {
	var ds []DataSet
	err := orm.AllByIndex("Spec", &ds)
	return ds, err
}

// Save saves the DataSet to BoltDB
func (ds DataSet) Save() error {
	ds.Modified = time.Now()
	return orm.Save(&ds)
}

// Delete deletes the DataSet from BoltDB
func (ds DataSet) Delete(ctx context.Context) error {
	if c.Config.ElasticSearch.Enabled || c.Config.RDF.RDFStoreEnabled {
		if _, err := ds.DropAll(ctx); err != nil {
			return err
		}
		return nil
	}
	return orm.DeleteStruct(&ds)
}

// IndexRecordRevisionsBySpec counts all the records stored in the Index for a Dataset
func (ds DataSet) IndexRecordRevisionsBySpec(ctx context.Context) (int, []DataSetRevisions, error) {
	revisions := []DataSetRevisions{}

	if !c.Config.ElasticSearch.Enabled {
		return 0, revisions, fmt.Errorf("IndexRecordRevisionsBySpec should not be called when elasticsearch is not enabled")
	}

	revisionAgg := elastic.NewTermsAggregation().Field("revision").Size(30).OrderByCountDesc()
	q := elastic.NewMatchPhraseQuery("spec", ds.Spec)
	res, err := index.ESClient().Search().
		Index(c.Config.ElasticSearch.IndexName).
		Type("rdfrecord").
		Query(q).
		Size(0).
		Aggregation("revisions", revisionAgg).
		Do(ctx)
	if err != nil {
		logger.WithField("spec", ds.Spec).Errorf("Unable to get IndexRevisionStats for the dataset.")
		return 0, revisions, err
	}
	fmt.Printf("total hits: %d\n", res.Hits.TotalHits)
	if res == nil {
		logger.Errorf("expected response != nil; got: %v", res)
		return 0, revisions, fmt.Errorf("expected response != nil")
	}
	aggs := res.Aggregations
	revAggCount, found := aggs.Terms("revisions")
	if !found {
		logger.Errorf("Expected to find revision aggregations but got: %v", res)
		return 0, revisions, fmt.Errorf("expected revision aggregrations")
	}
	for _, keyCount := range revAggCount.Buckets {
		revisions = append(revisions, DataSetRevisions{
			Number:      int(keyCount.Key.(float64)),
			RecordCount: int(keyCount.DocCount),
		})
	}
	totalHits := res.Hits.TotalHits
	return int(totalHits), revisions, err
}

// CreateDataSetStats returns DataSetStats that contain all relevant counts from the storage layer
func CreateDataSetStats(ctx context.Context, spec string) (DataSetStats, error) {
	storedGraphs, err := CountGraphsBySpec(spec)
	if err != nil {
		return DataSetStats{}, err
	}
	revisionCount, err := CountRevisionsBySpec(spec)
	if err != nil {
		return DataSetStats{}, err
	}
	ds, err := GetDataSet(spec)
	if err != nil {
		log.Printf("Unable to retrieve dataset %s: %s", spec, err)
		return DataSetStats{}, err
	}
	hits, indexRevisionCount, err := ds.IndexRecordRevisionsBySpec(ctx)
	if err != nil {
		log.Printf("Unable to get Index Revisions from ElasticSearch.")
		return DataSetStats{}, err
	}
	return DataSetStats{
		Spec:            spec,
		StoredGraphs:    storedGraphs,
		CurrentRevision: ds.Revision,
		GraphRevisions:  revisionCount,
		IndexRevisions:  indexRevisionCount,
		IndexedRecords:  hits,
	}, nil
}

// DeleteGraphsOrphans deletes all the orphaned graphs from the Triple Store linked to this dataset
func (ds DataSet) DeleteGraphsOrphans() (bool, error) {
	return DeleteGraphsOrphansBySpec(ds.Spec, ds.Revision)
}

// DeleteAllGraphs deletes all the graphs linked to this dataset
func (ds DataSet) DeleteAllGraphs() (bool, error) {
	return DeleteAllGraphsBySpec(ds.Spec)
}

// DeleteIndexOrphans deletes all the Orphaned records from the Search Index linked to this dataset
func (ds DataSet) DeleteIndexOrphans(ctx context.Context) (int, error) {
	q := elastic.NewBoolQuery()
	q = q.MustNot(elastic.NewMatchQuery("revision", ds.Revision))
	q = q.Must(elastic.NewMatchPhraseQuery("spec", ds.Spec))
	logger.Infof("%#v", q)
	res, err := index.ESClient().DeleteByQuery().
		Index(c.Config.ElasticSearch.IndexName).
		Type("rdfrecord").
		Query(q).
		Do(ctx)
	if err != nil {
		logger.WithField("spec", ds.Spec).Errorf("Unable to delete orphaned dataset records from index.")
		return 0, err
	}
	if res == nil {
		logger.Errorf("expected response != nil; got: %v", res)
		return 0, fmt.Errorf("expected response != nil")
	}
	logger.Infof("Removed %d records for spec %s with older revision than %d", res.Deleted, ds.Spec, ds.Revision)
	return int(res.Deleted), err
}

// DeleteAllIndexRecords deletes all the records from the Search Index linked to this dataset
func (ds DataSet) DeleteAllIndexRecords(ctx context.Context) (int, error) {
	q := elastic.NewMatchQuery("spec", ds.Spec)
	logger.Infof("%#v", q)
	res, err := index.ESClient().DeleteByQuery().
		Index(c.Config.ElasticSearch.IndexName).
		Type("rdfrecord").
		Query(q).
		Do(ctx)
	if err != nil {
		logger.WithField("spec", ds.Spec).Errorf("Unable to delete dataset records from index.")
		return 0, err
	}
	if res == nil {
		logger.Errorf("expected response != nil; got: %v", res)
		return 0, fmt.Errorf("expected response != nil")
	}
	logger.Infof("Removed %d records for spec %s", res.Deleted, ds.Spec)
	return int(res.Deleted), err
}

// DropOrphans

// DropRecords Drops all records linked to the dataset from the storage layers
func (ds DataSet) DropRecords(ctx context.Context) (bool, error) {
	var err error
	ok := false
	if c.Config.RDF.RDFStoreEnabled {
		ok, err = ds.DeleteAllGraphs()
		if !ok || err != nil {
			logger.Errorf("Unable to drop all graphs for %s", ds.Spec)
			return ok, err
		}
	}
	// todo add deleting all records from elastic search
	if c.Config.ElasticSearch.Enabled {
		_, err = ds.DeleteAllIndexRecords(ctx)
		if err != nil {
			logger.Errorf("Unable to drop all index records for %s: %#v", ds.Spec, err)
			return false, err
		}
	}
	return ok, err
}

// DropAll drops the dataset from the Rapid storages completely (BoltDB, Triple Store, Search Index)
func (ds DataSet) DropAll(ctx context.Context) (bool, error) {
	ok, err := ds.DropRecords(ctx)
	if !ok || err != nil {
		logger.Errorf("Unable to drop all records for spec %s: %#v", ds.Spec, err)
		return ok, err
	}
	err = orm.DeleteStruct(&ds)
	if err != nil {
		logger.Errorf("Unable to delete dataset %s from storage")
		return false, err
	}
	return ok, err
}
