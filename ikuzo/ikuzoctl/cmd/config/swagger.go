package config

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/swagger"
)

type Swagger struct {
	service *swagger.Service
}

func (o *Swagger) NewService(cfg *Config) (*swagger.Service, error) {
	if o.service != nil {
		return o.service, nil
	}

	svc, err := swagger.NewService()
	if err != nil {
		return nil, err
	}

	o.service = svc

	return svc, nil
}

func (o *Swagger) AddOptions(cfg *Config) error {
	svc, err := o.NewService(cfg)
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.RegisterService(svc),
	)

	return nil
}
