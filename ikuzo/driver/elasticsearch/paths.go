package elasticsearch

type MappingPath string

const (
	OrgID     MappingPath = "meta.orgID"
	DatasetID MappingPath = "meta.spec"
	Revision  MappingPath = "meta.revision"
)
