package dataset

import (
	"time"

	"github.com/delving/hub3/ikuzo/domain"
)

// DataSet contains all the known informantion for a hub3 metadata dataset
type DataSet struct {
	OrgID            string    `json:"orgID"`
	Spec             string    `json:"spec" storm:"id,index,unique"`
	URI              string    `json:"uri" storm:"unique"`
	Revision         int       `json:"revision"` // revision is used to mark the latest version of ingested RDFRecords
	FragmentRevision int       `json:"fragmentRevision"`
	Modified         time.Time `json:"modified" storm:"index"`
	Created          time.Time `json:"created"`
	Deleted          bool      `json:"deleted"`
	Tags             []string  `json:"tags"`
	RecordType       string    `json:"recordType"`
	Owner            string    `json:"owner"`
	Access           `json:"access" storm:"inline"`
	EAD              `json:"ead" storm:"inline"`
}

type EAD struct {
	Abstract       []string `json:"abstract"`
	ArchiveCreator []string `json:"archiveCreator"`
	Clevels        int      `json:"clevels"`
	DaoStats       `json:"daoStats" storm:"inline"`
	Description    string   `json:"description"`
	Files          string   `json:"files"`
	Fingerprint    string   `json:"fingerPrint"`
	Label          string   `json:"label"`
	Language       string   `json:"language"`
	Length         string   `json:"length"`
	Material       string   `json:"material"`
	MetsFiles      int      `json:"metsFiles"`
	Period         []string `json:"period"`
}

// Access determines which types of access are enabled for this dataset
type Access struct {
	OAIPMH bool `json:"oaipmh"`
	Search bool `json:"search"`
	LOD    bool `json:"lod"`
}

// NewDataset creates a new instance of a DataSet
func NewDataset(org *domain.Organization, spec string) DataSet {
	now := time.Now()
	access := Access{
		OAIPMH: true,
		Search: true,
		LOD:    true,
	}

	dataset := DataSet{
		OrgID:    org.RawID(),
		Spec:     spec,
		URI:      org.NewDatasetURI(spec),
		Created:  now,
		Modified: now,
		Access:   access,
		Revision: 1,
	}

	return dataset
}
