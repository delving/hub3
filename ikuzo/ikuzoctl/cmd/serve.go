package cmd

import (
	"github.com/delving/hub3/hub3/server/http/handlers"
	"github.com/delving/hub3/ikuzo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A high performance webserver",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

// nolint:gochecknoinits cobra requires this init
func init() {
	rootCmd.AddCommand(serveCmd)
}

// serve configures and runs the ikuzo server as a silo.
func serve() {
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
			handlers.RegisterEAD,
			handlers.RegisterSearch,
		),
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
			Msg("ikuzo server stopped with an error")
	}
}
