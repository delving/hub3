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
