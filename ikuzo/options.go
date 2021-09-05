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

package ikuzo

import (
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/service/x/bulk"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
	"github.com/delving/hub3/ikuzo/service/x/revision"
	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch"
	"github.com/delving/hub3/ikuzo/webapp"
	"github.com/go-chi/chi"
	"github.com/pacedotdev/oto/otohttp"
)

// RouterFunc is a callback that registers routes to the ikuzo.Server.
type RouterFunc func(router chi.Router)

// Option is a closure to configure the Server.
// It is used in NewServer.
type Option func(*server) error

// SetPort sets the TCP port for the Server.
//
// The Server listens on :3000 by default.
func SetPort(port int) Option {
	return func(s *server) error {
		s.port = port
		return nil
	}
}

// RegisterService registers a Service with the ikuzo server
func RegisterService(svc domain.Service) Option {
	return func(s *server) error {
		s.services = append(s.services, svc)
		return nil
	}
}

// SetMetricsPort sets the TCP port for the metrics server.
//
// No default. When set to 0 the metrics server is not started
func SetMetricsPort(port int) Option {
	return func(s *server) error {
		s.metricsPort = port
		return nil
	}
}

// SetTLS sets the TLS key and certificate.
//
// When both are set the server starts in TLS mode.
func SetTLS(cert, key string) Option {
	return func(s *server) error {
		s.certFile = cert
		s.keyFile = key

		return nil
	}
}

// SetLogger configures the global logger for the server.
func SetLogger(l *logger.CustomLogger) Option {
	return func(s *server) error {
		s.logger = l
		return nil
	}
}

// SetDisableRequestLogger disables logging of HTTP request
func SetDisableRequestLogger() Option {
	return func(s *server) error {
		s.disableRequestLogger = true
		return nil
	}
}

// SetMiddleware configures the global middleware for the HTTP router.
func SetMiddleware(middleware ...func(next http.Handler) http.Handler) Option {
	return func(s *server) error {
		s.middleware = append(s.middleware, middleware...)
		return nil
	}
}

// SetRouters adds all HTTP routes for the server.
func SetRouters(rb ...RouterFunc) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs, rb...)
		return nil
	}
}

// RegisterOtoServer registers an otohttp.Server.
//
// This enables the server to expose RPC on the '/oto/' endpoint
func RegisterOtoServer(otoServer *otohttp.Server) Option {
	return func(s *server) error {
		s.oto = otoServer
		return nil
	}
}

// SetOrganisationService configures the organization service.
func SetOrganisationService(svc *organization.Service) Option {
	return func(s *server) error {
		s.organizations = svc
		s.services = append(s.services, svc)
		s.middleware = append(s.middleware, svc.ResolveOrgByDomain)

		return nil
	}
}

// SetRevisionService configures the organization service.
// When no service is set a default transient memory-based service is used.
func SetRevisionService(service *revision.Service) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs, func(r chi.Router) {
			r.HandleFunc("/git/{user}/{collection}.git/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := strings.TrimPrefix(r.URL.Path, "/git")
				if !service.BareRepo {
					p = strings.ReplaceAll(p, ".git/", "/.git/")
				}
				r2 := new(http.Request)
				*r2 = *r
				r2.URL = new(url.URL)
				*r2.URL = *r.URL
				r2.URL.Path = p
				service.ServeHTTP(w, r2)
			}))
		})

		return nil
	}
}

func SetElasticSearchProxy(proxy *elasticsearch.Proxy) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Handle("/{index}/_search", proxy)
				r.Handle("/{index}/{documentType}/_search", proxy)
			},
		)

		return nil
	}
}

func SetBuildVersionInfo(info *BuildVersionInfo) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
					s.respond(w, r, info, http.StatusOK)
				})
			},
		)

		return nil
	}
}

// SetStaticFS registers an fs.FS as a static fileserver.
//
// It is mounts '/static/*' and '/favicon.ico'.
// Note: it can only be set once. So you can register multiple fs.FS.
func SetStaticFS(static fs.FS) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Get("/static/*", webapp.NewStaticHandler(static))
				r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, "/static/favicon.ico", http.StatusMovedPermanently)
				})
			},
		)

		return nil
	}
}

func SetOAIPMHService(svc *oaipmh.Service) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Get("/oai/!open_oai.OAIHandler", svc.ServeHTTP)
				r.Post("/oai/harvest-now", svc.HarvestNow)
			},
		)

		s.addShutdown("oai-pmh service", svc)

		return nil
	}
}

func SetBulkService(svc *bulk.Service) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Post("/api/index/bulk", svc.Handle)
				r.Post("/api/index/rdf", svc.HandleRDF)
			},
		)

		return nil
	}
}

// SetShutdownHook adds a shutdown hook to the ikuzo.Server.
//
// This should not be used for domain.Service implementations.
// Their shutdownHooks are registered automatically.
func SetShutdownHook(name string, hook domain.Shutdown) Option {
	return func(s *server) error {
		if _, ok := s.shutdownHooks[name]; !ok {
			s.shutdownHooks[name] = hook
		}

		return nil
	}
}

func SetEnableLegacyConfig(cfgFile string) Option {
	return func(s *server) error {
		// this initializes the hub3 configuration object that has global state
		// TODO(kiivihal): remove this after legacy hub3/server/http/handlers are migrated
		config.SetCfgFile(cfgFile)
		config.InitConfig()

		return nil
	}
}

func SetLegacyRouters(routers ...RouterFunc) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs, routers...)

		return nil
	}
}
