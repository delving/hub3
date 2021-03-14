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

	"github.com/delving/hub3/ikuzo/service/x/revision"
)

type Meta struct {
	basePath              string
	repo                  *revision.Repository
	OrgID                 string
	DatasetID             string
	Title                 string
	Clevels               uint64
	DaoLinks              uint64
	DaoErrors             uint64
	DaoErrorLinks         map[string]string
	Tags                  []string
	TotalRecordsPublished uint64
	DigitalObjects        uint64
	FileSize              uint64
	Revision              int32
	ProcessDigital        bool
	ProcessAccessTime     time.Time
	Created               bool
	ProcessingDuration    time.Duration `json:"processingDuration,omitempty"`
	ProcessingDurationFmt string        `json:"processingDurationFmt,omitempty"`
	RecordsUpdated        uint64
	RecordsDeleted        uint64
	PublishedCommitID     string
}

// getSourcePath returns full path to the source EAD file
func (m *Meta) getSourcePath() string {
	return getEADPath(m.DatasetID)
}

// getDaoLinkErrors returns the error messages for each retrieved error and its unitId
func (m *Meta) getDaoLinkErrors() (messages []string) {
	for id, errMsg := range m.DaoErrorLinks {
		messages = append(messages, fmt.Sprintf("Inventory %s => %s", id, errMsg))
	}
	return messages
}
