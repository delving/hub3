package ikuzo

func (s *server) BackgroundWorkers() error {
	// TODO(kiivihal): implement the mux for hibiken workers
	s.logger.Info().Msg("starting hibiken/asynq background workers")

	return nil
}
