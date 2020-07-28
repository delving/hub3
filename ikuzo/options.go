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
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/service/x/bulk"
	"github.com/delving/hub3/ikuzo/service/x/ead"
	"github.com/delving/hub3/ikuzo/service/x/imageproxy"
	"github.com/delving/hub3/ikuzo/service/x/revision"
	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch"
	"github.com/go-chi/chi"
)

const (
	taskIDRoute    = "/api/ead/tasks/{id}"
	datasetIDRoute = "/api/datasets/{spec}"
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

// SetOrganisationService configures the organization service.
// When no service is set a default transient memory-based service is used.
func SetOrganisationService(service *organization.Service) Option {
	return func(s *server) error {
		s.organizations = service
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Mount("/organizations", service.Routes())
			},
		)

		return nil
	}
}

// SetRevisionService configures the organization service.
// When no service is set a default transient memory-based service is used.
func SetRevisionService(service *revision.Service) Option {
	return func(s *server) error {
		s.revision = service
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

func SetEnableLegacyConfig() Option {
	return func(s *server) error {
		// this initializes the hub3 configuration object that has global state
		// TODO(kiivihal): remove this after legacy hub3/server/http/handlers are migrated
		config.InitConfig()

		return nil
	}
}

func SetLegacyRouters(routers ...RouterFunc) Option {
	return func(s *server) error {
		config.InitConfig()

		s.routerFuncs = append(s.routerFuncs, routers...)

		return nil
	}
}

func SetEADService(svc *ead.Service) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Post("/api/ead", svc.Upload)
				r.Get("/api/ead/tasks", svc.Tasks)
				r.Get(taskIDRoute, svc.GetTask)
				r.Delete(taskIDRoute, svc.CancelTask)
			},
		)

		return nil
	}
}

func SetBulkService(svc *bulk.Service) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Post("/api/index/bulk", svc.Handle)
			},
		)

		return nil
	}
}

func SetShutdownHook(name string, hook Shutdown) Option {
	return func(s *server) error {
		if _, ok := s.shutdownHooks[name]; !ok {
			s.shutdownHooks[name] = hook
		}

		return nil
	}
}

func SetImageProxyService(service *imageproxy.Service) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				r.Mount("/", service.Routes())
			},
		)

		return nil
	}
}

type ProxyRoute struct {
	Method  string
	Pattern string
}

// SetDataNodeProxy creates a reverse proxy to the dataNode and set override routes.
//
// The 'proxyRoutes' argument can be used to add additional override routes.
func SetDataNodeProxy(dataNode string, proxyRoutes ...ProxyRoute) Option {
	return func(s *server) error {
		nodeURL, _ := url.Parse(dataNode)
		s.dataNodeProxy = httputil.NewSingleHostReverseProxy(nodeURL)
		s.routerFuncs = append(s.routerFuncs,
			func(r chi.Router) {
				// ead
				r.Post("/api/ead", s.proxyDataNode)
				r.Get("/api/ead/tasks", s.proxyDataNode)
				r.Get(taskIDRoute, s.proxyDataNode)
				r.Delete(taskIDRoute, s.proxyDataNode)
				r.Post("/api/index/bulk", s.proxyDataNode)
				r.Get("/api/ead/{spec}/download", s.proxyDataNode)
				r.Get("/api/ead/{spec}/mets/{inventoryID}", s.proxyDataNode)
				r.Get("/api/ead/{spec}/desc", s.proxyDataNode)
				r.Get("/api/ead/{spec}/desc/index", s.proxyDataNode)
				r.Get("/api/ead/{spec}/meta", s.proxyDataNode)

				// datasets
				r.Get("/api/datasets/", s.proxyDataNode)
				r.Get("/api/datasets/histogram", s.proxyDataNode)
				r.Post("/api/datasets/", s.proxyDataNode)
				r.Get(datasetIDRoute, s.proxyDataNode)
				r.Get("/api/datasets/{spec}/stats", s.proxyDataNode)
				// later change to update dataset
				r.Post(datasetIDRoute, s.proxyDataNode)
				r.Delete(datasetIDRoute, s.proxyDataNode)

				// custom routes
				for _, route := range proxyRoutes {
					switch {
					case strings.EqualFold("get", route.Method):
						r.Get(route.Pattern, s.proxyDataNode)
					case strings.EqualFold("post", route.Method):
						r.Post(route.Pattern, s.proxyDataNode)
					case strings.EqualFold("put", route.Method):
						r.Put(route.Pattern, s.proxyDataNode)
					case strings.EqualFold("delete", route.Method):
						r.Delete(route.Pattern, s.proxyDataNode)
					}
				}
			},
		)

		return nil
	}
}
