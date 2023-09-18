package task

import (
	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynqmon"
)

func (s *Service) Routes(pattern string, r chi.Router) {
	h := asynqmon.New(asynqmon.Options{
		RootPath:     "/monitoring", // RootPath specifies the root for asynqmon app
		RedisConnOpt: s.redisClientOpt(),
	})

	r.Handle("/monitoring", h)
	r.Handle("/monitoring/*", h)
}
