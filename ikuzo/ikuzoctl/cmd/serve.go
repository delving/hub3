package cmd

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/revision"
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

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(serveCmd)
}

// serve configures and runs the ikuzo server as a silo.
func serve() {
	revisionSvc, err := revision.NewService("/tmp/hub3")
	if err != nil {
		log.Fatal().
			Err(err).
			Stack().
			Msg("unable to initialize revision service")
	}

	svr, err := ikuzo.NewServer(
		ikuzo.SetPort(3001),
		ikuzo.SetRevisionService(revisionSvc),
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
