package config

import "github.com/delving/hub3/ikuzo/storage/x/redis"

type Redis struct {
	Address            string `json:"address,omitempty"`
	Password           string `json:"password,omitempty"`
	UserName           string `json:"userName,omitempty"`
	Database           int    `json:"database,omitempty"`
	SentinelAddress    string `json:"sentinelAddress"`
	SentinelPassword   string `json:"sentinelPassword"`
	SentinelMasterName string `json:"sentinelMasterName"`
}

func (r *Redis) AddOptions(cfg *Config) error {
	return nil
}

func (r *Redis) RedisConfig() redis.Config {
	cfg := redis.Config{
		Address:  r.Address,
		Password: r.Password,
		UserName: r.UserName,
		Database: r.Database,
	}

	cfg.Sentinel.Address = r.SentinelAddress
	cfg.Sentinel.MasterName = r.SentinelMasterName
	cfg.Sentinel.Password = r.SentinelPassword

	return cfg
}
