package ikuzo

func (s *server) BackgroundWorkers() error {
	s.logger.Info().Msg("starting hibiken/asynq background workers")
	if s.ts == nil {
		s.logger.Warn().Msg("cannot start background workers when task.Service is not initialised")
		return nil
	}
	defer s.shutdown(nil)
	if err := s.ts.StartWorkers(s.ctx); err != nil {
		return err
	}

	return nil
}
