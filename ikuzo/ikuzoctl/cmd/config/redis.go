package config

import "github.com/delving/hub3/ikuzo/storage/x/redis"

type Redis struct {
	Address  string
	Password string
}

func (r *Redis) AddOptions(cfg *Config) error {
	return nil
}

func (r *Redis) redisConfig() redis.Config {
	return redis.Config{
		Address:  r.Address,
		Password: r.Password,
	}
}
