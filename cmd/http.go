// Copyright © 2017 Delving B.V. <info@delving>
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
	"github.com/delving/rapid-saas/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start the webserver process",
	Long: `Starting the webserver https process.

You can find an overview of available endpoints on https://localhost:443/endpoints`,
	Run: func(cmd *cobra.Command, args []string) {
		server.Start(buildInfo)
	},
}

func init() {
	RootCmd.AddCommand(httpCmd)

	httpCmd.Flags().IntP("port", "p", 3001, "Port to run Application server on")
	viper.BindPFlag("port", httpCmd.Flags().Lookup("port"))
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// httpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// httpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
