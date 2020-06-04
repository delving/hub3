package config

import "github.com/delving/hub3/ikuzo"

type HTTP struct {
	Port        int    `json:"port" mapstructure:"port"`
	MetricsPort int    `json:"metricsPort"`
	CertFile    string `json:"certFile"`
	KeyFile     string `json:"keyFile"`
}

func (http *HTTP) AddOptions(cfg *Config) error {
	cfg.options = append(
		cfg.options,
		ikuzo.SetPort(http.Port),
		ikuzo.SetTLS(http.CertFile, http.KeyFile),
	)

	if http.MetricsPort != 0 {
		cfg.options = append(cfg.options, ikuzo.SetMetricsPort(http.MetricsPort))
	}

	return nil
}
