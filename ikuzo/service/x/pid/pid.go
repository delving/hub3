package pid

import "time"

type Type int

const (
	Undefined Type = iota
	Ark
)

type PID struct {
	Type       Type
	ID         string
	OrgID      string
	DataSet    string
	HubID      string
	IsShownAt  string
	ReplacedBy string
	Tombstone  bool
	ModifiedAt time.Time
}

func (pid *PID) Merge(newPid PID) error {
	// TODO: which data to retain
	return nil
}
