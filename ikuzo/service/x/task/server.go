package task

import (
	"github.com/hibiken/asynq"
)

func (s *Service) redisClientOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     s.redisCfg.Address,
		Password: s.redisCfg.Password,
		DB:       0,
	}
}

func (s *Service) asynqServer() *asynq.Server {
	srv := asynq.NewServer(
		s.redisClientOpt(),
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: s.nrWorkers,
			// Optionally specify multiple queues with different priority.
			Queues:         s.queues,
			StrictPriority: true,
		},
	)

	return srv
}
