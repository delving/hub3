package task

import (
	"github.com/hibiken/asynq"
)

func (s *Service) redisClient() asynq.RedisConnOpt {
	if s.redisCfg.Sentinel.Address != "" {
		return s.RedisFailoverClientOpt()
	}

	return s.redisClientOpt()
}

func (s *Service) redisClientOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     s.redisCfg.Address,
		Password: s.redisCfg.Password,
		Username: s.redisCfg.UserName,
		DB:       0,
	}
}

func (s *Service) RedisFailoverClientOpt() asynq.RedisFailoverClientOpt {
	redis := asynq.RedisFailoverClientOpt{
		MasterName:       s.redisCfg.Sentinel.MasterName,
		SentinelAddrs:    []string{s.redisCfg.Sentinel.Address},
		SentinelPassword: s.redisCfg.Sentinel.Password,
		Password:         s.redisCfg.Password,
		Username:         s.redisCfg.UserName,
		DB:               0,
	}

	return redis
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
