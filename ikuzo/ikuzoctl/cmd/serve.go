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

package cmd

import (
	"embed"

	stdlog "log"

	"github.com/delving/hub3/hub3/server/http/handlers"
	"github.com/delving/hub3/ikuzo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed static
var staticFS embed.FS

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A high performance webserver",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		serve()
	},
}

// nolint:gochecknoinits // cobra requires this init
func init() {
	rootCmd.AddCommand(serveCmd)
}

// serve configures and runs the ikuzo server as a silo.
func serve() {
	stdlog.SetFlags(stdlog.LstdFlags | stdlog.Lshortfile)

	options, err := cfg.Options()
	if err != nil {
		log.Fatal().
			Err(err).
			Stack().
			Msg("unable to create options")
	}

	options = append(
		options,
		ikuzo.SetBuildVersionInfo(
			ikuzo.NewBuildVersionInfo(version, gitHash, buildAgent, buildStamp),
		),
		ikuzo.SetLegacyRouters(
			handlers.RegisterDatasets,
			handlers.RegisterEAD,
			handlers.RegisterSearch,
			handlers.RegisterLinkedDataFragments,
			handlers.RegisterLOD,
			handlers.RegisterSparql,
		),
		ikuzo.SetEnableLegacyConfig(cfgFile),
		ikuzo.SetStaticFS(staticFS),
	)

	svr, err := ikuzo.NewServer(
		options...,
	)
	if err != nil {
		log.Fatal().
			Err(err).
			Stack().
			Msg("unable to initialize ikuzo server")
	}

	err = svr.ListenAndServe()
	if err != nil {
		log.Fatal().
			Err(err).
			Stack().
			Msgf("ikuzo server stopped with an error: %s", err)
	}
}
