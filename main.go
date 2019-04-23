// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import "github.com/delving/hub3/cmd"

var (
	// Version of the application. (Injected at build time)
	Version = "0.1.0-SNAPSHOT"
	// BuildStamp is the timestamp of the application. (Injected at build time)
	BuildStamp = "1970-01-01 UTC"
	// BuildAgent is the agent that created the current build. (Injected at build time)
	BuildAgent string
	// GitHash of the current build. (Injected at build time.)
	GitHash string
	// BuildID is de build ID injected during a Continueus Integration Job
	BuildID string
)

func main() {
	cmd.Execute(Version, BuildStamp, BuildAgent, GitHash, BuildID)
}
