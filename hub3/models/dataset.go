package models

import (
	"fmt"
	"time"

	. "bitbucket.org/delving/rapid/config"
)

// DataSet contains all the known informantion for a RAPID metadata dataset
type DataSet struct {
	Spec string `json:"spec" storm:"id,index"`
	URI  string `json:"uri" storm:"unique,index"`
	//MapToPrefix string    `json:"mapToPrefix"`
	Revision int       `json:"revision"` // revision is used to mark the latest version of ingested RDFRecords
	Modified time.Time `json:"modified" storm:"index"`
	Created  time.Time `json:"created"`
	Deleted  bool      `json:"deleted"`
}

// createDatasetURI creates a RDF uri for the dataset based Config RDF BaseUrl
func createDatasetURI(spec string) string {
	uri := fmt.Sprintf("%s/resource/dataset/%s", Config.RDF.BaseUrl, spec)
	return uri
}

// NewDataset creates a new instance of a DataSet
func NewDataset(spec string) DataSet {
	now := time.Now()
	dataset := DataSet{
		Spec:     spec,
		URI:      createDatasetURI(spec),
		Created:  now,
		Modified: now,
	}
	return dataset
}

// Save saves the DataSet to BoltDB
func (ds DataSet) Save() error {
	ds.Modified = time.Now()
	return orm.Save(&ds)
}
