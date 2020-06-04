package ead

import (
	"fmt"
	"time"
)

type Meta struct {
	basePath              string
	OrgID                 string
	DatasetID             string
	Title                 string
	Clevels               uint64
	DaoLinks              uint64
	RecordsPublished      uint64
	DigitalObjects        uint64
	FileSize              uint64
	ProcessingDuration    time.Duration `json:"processingDuration,omitempty"`
	ProcessingDurationFmt string        `json:"processingDurationFmt,omitempty"`
}

// getSourcePath returns full path to the source EAD file
func (m *Meta) getSourcePath() string {
	return fmt.Sprintf("%s/%s.xml", m.basePath, m.DatasetID)
}
