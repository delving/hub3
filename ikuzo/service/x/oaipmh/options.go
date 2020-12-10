package oaipmh

type Option func(*Service) error

func AddTask(task ...HarvestTask) Option {
	return func(s *Service) error {
		if len(task) > 0 {
			s.tasks = append(s.tasks, task...)
		}
		return nil
	}
}

func SetDelay(delay int) Option {
	return func(s *Service) error {
		s.defaultDelay = delay
		return nil
	}
}
