package trs

import "github.com/delving/hub3/ikuzo/domain"

type Record interface {
	Formats() []string
	Actions() []string
}

type Blob struct {
	// unique identifier for record
	ID domain.HubID
	// TypeID is the registered RecordType
	RecordType     string
	Metadata       map[string]string
	MetadataCustom []byte
	Data           []byte
	Versions       []BlobVersion
}

type BlobVersion struct {
	// sortable identifier
	ID    string // ULID
	Delta []byte
}

type ListOptions struct {
	RecordType string
}

func (s *Service) List(opts ListOptions) error {
	return nil
}

type GetOptions struct {
	Version string // (ulid)
	Diff    string
	Action  string
	Format  string
}

func (s *Service) Get(id string, opts GetOptions) error {
	return nil
}
