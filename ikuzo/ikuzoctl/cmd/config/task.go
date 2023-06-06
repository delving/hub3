package config

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/task"
)

type Task struct {
	NrWorkers       int
	ExternalWorkers bool
}

func (t *Task) AddOptions(cfg *Config) error {
	svc, err := cfg.GetTaskService()
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.RegisterService(svc),
		ikuzo.SetTaskService(svc),
	)

	return nil
}

func (cfg *Config) GetTaskService() (*task.Service, error) {
	if cfg.ts != nil {
		return cfg.ts, nil
	}

	return cfg.Task.newService(cfg)
}

func (t *Task) newService(cfg *Config) (*task.Service, error) {
	svc, err := task.NewService(
		task.SetRedisConfig(cfg.Redis.redisConfig()),
		task.SetNrWorkers(t.NrWorkers),
		task.SetHasExternalWorkers(t.ExternalWorkers),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to setup task service from cfg; %w", err)
	}

	return svc, nil
}