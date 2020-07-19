package dataset

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

// DaoStats holds the stats for EAD digital objects extracted from METS links.
type DaoStats struct {
	ExtractedLinks uint64         `json:"extractedLinks"`
	RetrieveErrors uint64         `json:"retrieveErrors"`
	DigitalObjects uint64         `json:"digitalObjects"`
	Errors         []string       `json:"errors"`
	UniqueLinks    uint64         `json:"uniqueLinks"`
	DuplicateLinks map[string]int `json:"duplicateLinks"`
}

// DataSetCounter holds value counters for statistics overviews
type DataSetCounter struct {
	Value    string `json:"value"`
	DocCount int    `json:"docCount"`
}

// DataSetRevisions holds the type-frequency data for each revision
type DataSetRevisions struct {
	Number      int `json:"revisionNumber"`
	RecordCount int `json:"recordCount"`
}

// IndexStats hold all Index Statistics for this dataset
type IndexStats struct {
	Enabled        bool               `json:"enabled"`
	Revisions      []DataSetRevisions `json:"revisions"`
	IndexedRecords int                `json:"indexedRecords"`
	Tags           []DataSetCounter   `json:"tags"`
	ContentTags    []DataSetCounter   `json:"contentTags"`
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

// NarthexStats gathers all the record statistics from Narthex
type NarthexStats struct {
	Enabled        bool `json:"enabled"`
	SourceRecords  int  `json:"sourceRecords"`
	ValidRecords   int  `json:"validRecords"`
	InvalidRecords int  `json:"invalidRecords"`
}

// RDFStoreStats hold all the RDFStore Statistics for this dataset
type RDFStoreStats struct {
	Revisions    []DataSetRevisions `json:"revisions"`
	StoredGraphs int                `json:"storedGraphs"`
	Enabled      bool               `json:"enabled"`
}

// WebResourceStats gathers all the MediaManager information for this DataSet
type WebResourceStats struct {
	Enabled           bool `json:"enabled"`
	SourceItems       int  `json:"sourceItems"`
	ThumbnailsCreated int  `json:"thumbnailsCreated"`
	DeepZoomsCreated  int  `json:"deepZoomsCreated"`
	Missing           int  `json:"missing"`
}

// VocabularyEnrichmentStats gathers all counters for the SKOS based enrichment
type VocabularyEnrichmentStats struct {
	LiteralFields        []string `json:"literalFields"`
	TotalConceptsMapped  int      `json:"totalConceptsMapped"`
	UniqueConceptsMapped int      `json:"uniqueConceptsMapped"`
	VocabularyLinked     []string `json:"vocabularyLinked"`
}
