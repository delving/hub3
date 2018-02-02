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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Verbose logs extra information when the version command is called.
var Verbose bool

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version and build information.",
	Run: func(cmd *cobra.Command, args []string) {
		if !Verbose {
			fmt.Println(RootCmd.Use + " " + buildInfo.Version)
		} else {
			info, err := buildInfo.JSON(true)
			if err != nil {
				fmt.Printf("%+v\n", buildInfo)
			} else {
				fmt.Printf("%s\n", info)
			}
		}
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	RootCmd.AddCommand(versionCmd)

}
