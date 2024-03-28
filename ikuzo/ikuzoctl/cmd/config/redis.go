package config

import "github.com/delving/hub3/ikuzo/storage/x/redis"

type Redis struct {
	Address  string `json:"address,omitempty"`
	Password string `json:"password,omitempty"`
	Database int    `json:"database,omitempty"`
}

func (r *Redis) AddOptions(cfg *Config) error {
	return nil
}

func (r *Redis) RedisConfig() redis.Config {
	return redis.Config{
		Address:  r.Address,
		Password: r.Password,
		Database: r.Database,
	}
}
