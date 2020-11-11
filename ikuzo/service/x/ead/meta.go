// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	DaoErrors             uint64
	DaoErrorLinks         []string
	Tags                  []string
	RecordsPublished      uint64
	DigitalObjects        uint64
	FileSize              uint64
	Revision              int32
	ProcessDigital        bool
	ProcessAccessTime     time.Time
	Created               bool
	ProcessingDuration    time.Duration `json:"processingDuration,omitempty"`
	ProcessingDurationFmt string        `json:"processingDurationFmt,omitempty"`
}

// getSourcePath returns full path to the source EAD file
func (m *Meta) getSourcePath() string {
	return fmt.Sprintf("%s/%s.xml", m.basePath, m.DatasetID)
}
