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

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/service/x/ead"
	"github.com/spf13/cobra"
)

var (
	eadPath string
)

var eadResyncCmd = &cobra.Command{
	Use:   "eadResync",
	Short: "update ead from EAD cache",
	Run: func(cmd *cobra.Command, args []string) {

		config.InitConfig()

		svc, err := ead.NewService(
			ead.SetDataDir(eadPath),
			ead.SetWorkers(4),
			ead.SetProcessDigital(true),
		)
		if err != nil {
			log.Fatalf("unable to start EAD service: %s", err)
			return
		}

		if err := svc.StartWorkers(); err != nil {
			log.Fatalf("unable to start EAD service workers; %s", err)
			return
		}

		if err := svc.ResyncCacheDir(); err != nil {
			log.Fatalf("unable to sync ead cache directories; %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(eadResyncCmd)

	eadResyncCmd.Flags().StringVarP(&eadPath, "path", "p", "", "full path ead directory")
}
