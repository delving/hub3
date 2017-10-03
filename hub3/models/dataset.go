package models

import (
	"fmt"
	"time"

	. "bitbucket.org/delving/rapid/config"
)

// DataSet contains all the known informantion for a RAPID metadata dataset
type DataSet struct {
	Spec string `json:"spec" storm:"id,indexed"`
	Uri  string `json:"uri" storm:"unique,indexed"`
	//MapToPrefix string    `json:"mapToPrefix"`
	Revision int       `json:"revision"` // revision is used to mark the latest version of ingested RDFRecords
	Modified time.Time `json:"modified" storm:"indexed"`
	Created  time.Time `json:"created"`
	Deleted  bool      `json:"deleted"`
}

func createDatasetURI(spec string) string {
	uri := fmt.Sprintf("%s/resource/dataset/%s", Config.RDF.BaseUrl, spec)
	return uri
}

func NewDataset(spec string) DataSet {
	now := time.Now()
	dataset := DataSet{
		Spec:     spec,
		Uri:      createDatasetURI(spec),
		Created:  now,
		Modified: now,
	}
	return dataset
}
