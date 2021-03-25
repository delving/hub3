package domain

// DataSetRevisions holds the type-frequency data for each revision for a given DataSet
type DataSetRevisions struct {
	Number      int `json:"revisionNumber"`
	RecordCount int `json:"recordCount"`
}
