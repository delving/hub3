// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"expvar"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/imageproxy"
)

type ImageProxy struct {
	Enabled         bool
	CacheDir        string
	MaxSizeCacheDir int
	ProxyPrefix     string
	Timeout         int
	ProxyReferrer   []string
	RefuseList      []string
	AllowList       []string
	LruCacheSize    int
	EnableResize    bool
}

func (ip *ImageProxy) AddOptions(cfg *Config) error {
	if !ip.Enabled {
		return nil
	}

	s, err := imageproxy.NewService(
		imageproxy.SetCacheDir(ip.CacheDir),
		imageproxy.SetMaxSizeCacheDir(ip.MaxSizeCacheDir),
		imageproxy.SetProxyPrefix(ip.ProxyPrefix),
		imageproxy.SetTimeout(ip.Timeout),
		imageproxy.SetProxyReferrer(ip.ProxyReferrer),
		imageproxy.SetRefuseList(ip.RefuseList),
		imageproxy.SetAllowList(ip.AllowList),
		imageproxy.SetLruCacheSize(ip.LruCacheSize),
		imageproxy.SetEnableResize(ip.EnableResize),
		imageproxy.SetLogger(cfg.logger.Logger),
	)
	if err != nil {
		return err
	}

	cfg.options = append(cfg.options, ikuzo.SetImageProxyService(s))

	expvar.Publish("hub3-imageproxy-requests", expvar.Func(func() interface{} { m := s.RequestMetrics(); return m }))
	expvar.Publish("hub3-imageproxy-cache", expvar.Func(func() interface{} { m := s.CacheMetrics(); return m }))

	return nil
}
