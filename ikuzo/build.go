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

package ikuzo

import (
	"encoding/json"
	"fmt"
)

// BuildVersionInfo holds all the version information
type BuildVersionInfo struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	BuildAgent string `json:"buildAgent"`
	BuildDate  string `json:"buildDate"`
	BuildID    string `json:"buildID"`
}

// NewBuildVersionInfo creates a BuildVersionInfo struct
func NewBuildVersionInfo(version, commit, buildagent, builddate string) *BuildVersionInfo {
	if version == "" {
		version = "devBuild"
	}

	return &BuildVersionInfo{
		Version:    version,
		Commit:     commit,
		BuildAgent: buildagent,
		BuildDate:  builddate,
	}
}

func (info *BuildVersionInfo) String() string {
	b, err := json.MarshalIndent(info, "", "\t")
	if err != nil {
		return fmt.Sprintln("Unable to marshal build information")
	}

	return string(b)
}
