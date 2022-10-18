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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// workerCmd represents the worker command
//
// This starts hibiken/asynq workers
var workerCmd = &cobra.Command{
	Use:   "workers",
	Short: "Ikuzo background workers",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		workers()
	},
}

// nolint:gochecknoinits // cobra requires this init
func init() {
	rootCmd.AddCommand(workerCmd)
}

// workers configures and runs the ikuzo background workers.
func workers() {
	svr, err := setupIkuzo(true)
	if err != nil {
		log.Fatal().
			Err(err).
			Stack().
			Msg("unable to initialize ikuzo server")
	}

	err = svr.BackgroundWorkers()
	if err != nil {
		log.Fatal().
			Err(err).
			Stack().
			Msgf("ikuzo server stopped with an error: %s", err)
	}
}
