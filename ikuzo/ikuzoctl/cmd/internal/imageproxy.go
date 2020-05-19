package internal

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/imageproxy"
)

type ImageProxy struct {
	Enabled     bool
	CacheDir    string
	ProxyPrefix string
	Timeout     int
}

func (ip *ImageProxy) AddOptions(cfg *Config) error {
	if !ip.Enabled {
		return nil
	}

	s, err := imageproxy.NewService(
		imageproxy.SetCacheDir(ip.CacheDir),
		imageproxy.SetProxyPrefix(ip.ProxyPrefix),
		imageproxy.SetTimeout(ip.Timeout),
	)

	if err != nil {
		return err
	}

	cfg.options = append(cfg.options, ikuzo.SetImageProxyService(s))

	return nil
}
