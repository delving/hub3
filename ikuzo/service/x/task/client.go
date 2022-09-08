package task

import "github.com/hibiken/asynq"

func (s *Service) asynqClient() *asynq.Client {
	return asynq.NewClient(
		s.redisClientOpt(),
	)
}
