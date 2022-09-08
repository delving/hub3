package ikuzo

func (s *server) BackgroundWorkers() error {
	s.logger.Info().Msg("starting hibiken/asynq background workers")
	defer s.shutdown(nil)
	if err := s.ts.StartWorkers(s.ctx); err != nil {
		return err
	}

	return nil
}
