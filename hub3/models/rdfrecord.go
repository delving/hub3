package models

import (
	"fmt"
	"log"
	"strings"
	"time"

	"bitbucket.org/delving/rapid/config"
	"github.com/asdine/storm/q"
)

// RDFRecord contains all the information about a grouping of RDF triples
// that are considered a single search record.
// RDFRecord can be stored in various backends. The default is a Boltdb database
type RDFRecord struct {
	HubID         string    `json:"hubId" storm:"id,index"`
	ContentHash   string    `json:"contentHash"`
	Spec          string    `json:"spec" storm:"index"`
	Graph         string    `json:"graph"`
	NamedGraphURI string    `json:"NamedGraphURI" storm:"unique"`
	Modified      time.Time `json:"modified" storm:"index"`
	Created       time.Time `json:"created"`
	Revision      int       `json:"revision" storm:"index"` // the revision is used to mark records as orphans. it is autoincremented on each full save of the dataset
	Deleted       bool      `json:"deleted"`                // Deleted marks a record as an orphan
}

// createSourceURI creates a RDF uri for the RDFRecord based Config RDF BaseUrl
func (r RDFRecord) createSourceURI() string {
	_, spec, localID, err := r.ExtractHubID()
	if err != nil {
		log.Printf("Unable to extract hubId for %s", r.HubID)
	}
	uri := fmt.Sprintf("%s/resource/%s/%s", config.Config.RDF.BaseUrl, spec, localID)
	return uri
}

// NewRDFRecord creates a new RDFRecord
func NewRDFRecord(hubID string, spec string) RDFRecord {
	return RDFRecord{
		HubID:   hubID,
		Created: time.Now(),
		Spec:    spec,
	}
}

// CountRDFRecords returns an int with the records count for spec.
// If the spec is empty it should return a count for all
func CountRDFRecords(spec string) int {
	var record RDFRecord
	var count int
	var err error
	if spec != "" {
		count, err = orm.Select(q.Eq("Spec", spec)).Count(&record)
	} else {
		count, err = orm.Count(&record)
	}
	if err != nil {
		log.Printf("Unable to count for spec: %s", spec)
	}
	return count
}

// GetOrCreateRDFRecord returns a RDFRecord object from the Storm ORM.
// If none is present it will create one
func GetOrCreateRDFRecord(hubID, spec string) (RDFRecord, error) {
	var record RDFRecord
	err := orm.One("HubID", hubID, &record)
	if err != nil {
		record = NewRDFRecord(hubID, spec)
		err = record.Save()
		if err != nil {
			log.Println(err)
		}
	}
	return record, nil
}

// Save saves the RDFRecord to Boltdb
func (record RDFRecord) Save() error {
	record.Modified = time.Now()
	return orm.Save(&record)
}

// ExtractHubID extracts the orgId, spec and localId from the HubID
func (record RDFRecord) ExtractHubID() (orgID string, spec string, localID string, err error) {
	parts := strings.Split(record.HubID, "_")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("%s is not properly formatted. It should have three parts: orgid_spec_localid.", record.HubID)
	}
	return parts[0], parts[1], parts[2], nil
}

//// TODO: func toRdfSearchRecord
