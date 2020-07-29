// Copyright © 2019 Delving B.V. <info@delving>
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
	"log"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/server/http"
	"github.com/delving/hub3/hub3/server/http/handlers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start the webserver process",
	Long:  `Starting the webserver http process.`,
	Run: func(cmd *cobra.Command, args []string) {

		routers := []http.RouterCallBack{
			handlers.RegisterBulkIndexer,
			handlers.RegisterCSV,
			handlers.RegisterDatasets,
			handlers.RegisterEAD,
			handlers.RegisterElasticSearchProxy,
			handlers.RegisterLOD,
			handlers.RegisterLinkedDataFragments,
			handlers.RegisterSparql,
			handlers.RegisterSearch,
		}
		server, err := http.NewServer(
			http.SetBuildInfo(buildInfo),
			http.SetIntroSpection(true),
			http.SetRouters(routers...),
			http.SetPort(config.Config.HTTP.Port),
		)
		if err != nil {
			log.Fatal(err)
		}
		err = server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(httpCmd)

	httpCmd.Flags().IntP("port", "p", 3001, "Port to run Application server on")
	viper.BindPFlag("port", httpCmd.Flags().Lookup("port"))
}
