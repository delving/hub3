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

package imageproxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"golang.org/x/sync/singleflight"

	"github.com/delving/hub3/ikuzo/domain"

	lru "github.com/hashicorp/golang-lru/v2"
)

var ErrRemoteResourceNotFound = errors.New("remote resource not found")

// var _ domain.Service = (*Service)(nil)

type Service struct {
	client           http.Client
	lruCache         *lru.Cache[string, []byte]
	cacheDir         string // The path to the imageCache
	maxSizeCacheDir  int    //  max size of the cache directory on disk in kb
	timeOut          int    // timelimit for request served by this proxy. 0 is for no timeout
	proxyPrefix      string // The prefix where we mount the imageproxy. default: imageproxy. default: imageproxy.
	referrers        []string
	allowList        []string
	allowPorts       []string
	refuselist       []string
	allowedMimeTypes []string
	m                RequestMetrics
	cm               CacheMetrics
	log              zerolog.Logger
	enableResize     bool
	singleSetCache   singleflight.Group
	cancelWorker     context.CancelFunc
	orgs             domain.OrgConfigRetriever
	defaultImagePath string // The path to the defaultimage
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		cacheDir:    "/tmp/imageproxy",
		timeOut:     10,
		proxyPrefix: "imageproxy",
		log:         zerolog.Nop(),
		cm:          newCacheMetrics(),
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.enableResize {
		s.enableResize = s.checkForVips()
	}

	if s.maxSizeCacheDir > 0 {
		s.startCacheWorker()
	}

	s.client = http.Client{Timeout: time.Duration(s.timeOut) * time.Second}

	if err := os.MkdirAll(s.cacheDir, os.ModePerm); err != nil {
		return s, err
	}

	return s, nil
}

func (s *Service) checkForVips() bool {
	_, err := exec.LookPath("vips")
	if err != nil {
		s.log.Warn().Msg("libvips is not installed so disabling imageproxy resize options")
		return false
	}

	return true
}

func deepZoomExternally(from string) error {
	cleanFrom := strings.TrimSuffix(from, ".dzi")
	args := []string{
		"dzsave",
		cleanFrom,
		cleanFrom,
	}

	path, err := exec.LookPath("vips")
	if err != nil {
		return err
	}

	cmd := exec.Command(path, args...)

	return cmd.Run()
}

func resizeExternally(from, to, size string) error {
	args := []string{
		"--size", size,
		"--output", to,
		from,
	}

	path, err := exec.LookPath("vipsthumbnail")
	if err != nil {
		return err
	}

	cmd := exec.Command(path, args...)

	return cmd.Run()
}

func (s *Service) Do(ctx context.Context, req *Request, w io.Writer) error {
	_ = ctx

	s.log.Debug().
		Str("cacheKey", req.CacheKey).
		Str("sourceURL", req.SourceURL).
		Msg("processing cache key")

	if s.lruCache != nil {
		if s.lruCache.Contains(req.CacheKey) {
			data, ok := s.lruCache.Get(req.CacheKey)
			if ok {
				req.CacheType = Lru

				written, err := fmt.Fprintf(w, "%s", data)

				s.m.IncBytesServed(int64(written))

				s.m.IncLruCache()

				return err
			}
		}
	}

	_, isCached := existsInCache(req.cacheKeyPath())

	if !isCached {
		_, hasSource := existsInCache(req.downloadedSourcePath())
		if !hasSource {
			_, err, shared := s.singleSetCache.Do(
				req.SourceURL,
				func() (interface{}, error) {
					s.log.Info().Str("path", req.downloadedSourcePath()).Msg("started storing source")
					err := s.storeSource(req)
					if err != nil {
						s.log.Error().Err(err).Str("url", req.SourceURL).
							Str("sourcePath", req.downloadedSourcePath()).
							Str("storePath", req.cacheKeyPath()).
							Msgf("unexpected error saving source; %s", err)
						return nil, err
					}
					s.log.Info().Str("path", req.downloadedSourcePath()).Msg("finished storing source")
					return nil, nil
				},
			)

			if err != nil {
				return err
			}

			if !shared {
				req.CacheType = Source
			}

			hasSource = true
		}

		if !isCached && req.thumbnailOpts != "" && hasSource {
			_, err, _ := s.singleSetCache.Do(
				req.CacheKey,
				func() (interface{}, error) {
					err := resizeExternally(req.downloadedSourcePath(), req.cacheKeyPath(), req.thumbnailOpts)
					if err == nil {
						info, ok := existsInCache(req.cacheKeyPath())
						if ok {
							if cacheErr := s.updateCacheMetrics("", info, false); cacheErr != nil {
								return nil, cacheErr
							}
						}
					}
					return nil, err
				},
			)
			if err != nil {
				s.log.Error().Err(err).Str("url", req.SourceURL).
					Str("sourcePath", req.downloadedSourcePath()).
					Str("storePath", req.cacheKeyPath()).
					Msgf("unexpected error creating thumbnail; %s", err)

				req.CacheKey = encodeURL(req.SourceURL)
			}

			s.m.IncResize()
		}

		if !isCached && req.SubPath == deepZoomSuffix {
			_, err, _ := s.singleSetCache.Do(
				req.CacheKey,
				func() (interface{}, error) {
					err := deepZoomExternally(req.downloadedSourcePath())
					if err == nil {
						info, ok := existsInCache(req.cacheKeyPath())
						if ok {
							if cacheErr := s.updateCacheMetrics("", info, false); cacheErr != nil {
								return nil, cacheErr
							}
							tiles, size := s.countTiles(strings.ReplaceAll(req.cacheKeyPath(), ".dzi", "_files"))
							s.cm.addDeepZoomTiles(tiles, size)
						}
					}
					return nil, err
				},
			)
			if err != nil {
				s.log.Error().Err(err).Str("url", req.SourceURL).Msg("unexpected error creating deepzoom")
				s.m.IncError()

				return err
			}

			s.m.IncDeepZoom()
		}
	}

	// check cache
	r, err := req.Read(req.cacheKeyPath())
	if err != nil && !errors.Is(err, ErrCacheKeyNotFound) {
		s.log.Error().Err(err).Str("url", req.SourceURL).Msg("unexpected error reading from cache")
		s.m.IncError()

		return err
	}

	if !errors.Is(err, ErrCacheKeyNotFound) {
		defer r.Close()

		var buf bytes.Buffer
		tee := io.TeeReader(r, &buf)

		written, err := io.Copy(w, tee)
		if err != nil {
			s.log.Error().Err(err).Str("url", req.SourceURL).Msg("error copying data from cache")
			s.m.IncError()

			return err
		}

		s.m.IncBytesServed(written)

		if req.CacheType == "" {
			req.CacheType = Cache
		}

		s.lruCache.Add(req.CacheKey, buf.Bytes())

		s.m.IncCache()
	}

	return nil
}

func (s *Service) storeSource(req *Request) error {
	// make request
	proxyRequest, err := req.GET()
	if err != nil {
		s.log.Error().Err(err).Str("url", req.SourceURL).Msg("unable to create GET request")
		s.m.IncRemoteRequestError()

		return err
	}

	resp, err := s.client.Do(proxyRequest)
	if err != nil {
		s.log.Error().Err(err).Str("url", req.SourceURL).Msg("unable to make remote request")
		s.m.IncRemoteRequestError()

		return err
	}

	if resp.StatusCode != http.StatusOK {
		s.log.Error().Err(err).Int("status_code", resp.StatusCode).Str("url", req.SourceURL).Msg("retrieve object")
		s.m.IncRemoteRequestError()

		return fmt.Errorf("status_code: %d; %w", resp.StatusCode, ErrRemoteResourceNotFound)
	}

	contentType := resp.Header.Get("Content-Type")

	if len(s.allowedMimeTypes) != 0 {
		var allowed bool

		for _, mimeType := range s.allowedMimeTypes {
			if strings.EqualFold(mimeType, contentType) {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("mimeType %s is not allowed", contentType)
		}
	}

	defer resp.Body.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	if strings.HasPrefix(contentType, "text/xml") && bytes.Contains(buf.Bytes(), []byte("adlibXML")) {
		// don't cache adlib error messages
		s.log.Warn().Str("url", req.SourceURL).Msg("adlib error retrieving image")
		s.m.IncRemoteRequestError()

		return fmt.Errorf("unable to retrieve adlib result")
	}

	// check for adlib error when content type is xml
	size, err := req.Write(req.downloadedSourcePath(), tee)
	if err != nil {
		// do not return error here or cache write error
		s.log.Error().Err(err).Msgf("unable to write remote file to cache; %s", err)

		s.m.IncRemoteRequestError()
	}

	s.cm.addSourceFile(size)

	s.m.IncSource()

	return nil
}

func (s *Service) portsAllowed(targetURL string) (bool, error) {
	if len(s.allowPorts) == 0 {
		return true, nil
	}

	var allowed bool

	u, err := url.Parse(targetURL)
	if err != nil {
		s.log.Error().Err(err).Str("target_url", targetURL).Msg("unable to parse target url")
		return false, err
	}

	if u.Port() == "" {
		return true, nil
	}

	for _, target := range s.allowPorts {
		if strings.EqualFold(u.Port(), target) {
			log.Printf("port %s", u.Port())
			allowed = true
			break
		}
	}

	if !allowed {
		return false, nil
	}

	return true, nil
}

func (s *Service) domainAllowed(targetURL string) (bool, error) {
	if len(s.allowList) == 0 {
		return true, nil
	}

	var allowed bool

	for _, target := range s.allowList {
		u, err := url.Parse(targetURL)
		if err != nil {
			s.log.Error().Err(err).Str("target_url", targetURL).Msg("unable to parse target url")
			return false, err
		}

		if strings.HasSuffix(u.Host, target) {
			allowed = true
			break
		}
	}

	if !allowed {
		return false, nil
	}

	return true, nil
}

func (s *Service) reffererAllowed(referrer string) bool {
	if len(s.referrers) == 0 {
		return true
	}

	var allowed bool

	for _, allowedReferrer := range s.referrers {
		if strings.Contains(referrer, allowedReferrer) {
			allowed = true
			break
		}
	}

	return allowed
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

func (s *Service) Shutdown(ctx context.Context) error {
	if s.cancelWorker != nil {
		s.cancelWorker()
	}

	return nil
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "imageproxy").Logger()
	s.orgs = b.Orgs
}
