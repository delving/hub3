/*
Copyright © 2020 Delving B.V. <info@delving.eu>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"log"

	"github.com/delving/hub3/hub3/ead"
	"github.com/spf13/cobra"
)

var (
	eadPath string
)

// eadUpdateCmd represents the eadUpdate command
var eadUpdateCmd = &cobra.Command{
	Use:   "eadUpdate",
	Short: "update ead description from disk",
	Run: func(cmd *cobra.Command, args []string) {
		err := ead.ResaveDescriptions(eadPath)
		if err != nil {
			log.Fatalf("unable to resave descriptions: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(eadUpdateCmd)

	eadUpdateCmd.Flags().StringVarP(&eadPath, "path", "p", "", "full path ead directory")
}
