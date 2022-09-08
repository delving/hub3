package cmd

import (
	stdlog "log"

	"github.com/rs/zerolog/log"

	"github.com/delving/hub3/hub3/server/http/handlers"
	"github.com/delving/hub3/ikuzo"
)

func setupIkuzo(background bool) (ikuzo.Server, error) {
	stdlog.SetFlags(stdlog.LstdFlags | stdlog.Lshortfile)

	if background {
		cfg.Nats.Enabled = false
	}

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

	return ikuzo.NewServer(
		options...,
	)
}
