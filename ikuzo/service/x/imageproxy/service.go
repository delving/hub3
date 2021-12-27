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
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/singleflight"

	lru "github.com/hashicorp/golang-lru"
)

// var _ domain.Service = (*Service)(nil)

type Service struct {
	client          http.Client
	lruCache        *lru.ARCCache
	cacheDir        string // The path to the imageCache
	maxSizeCacheDir int    //  max size of the cache directory on disK
	timeOut         int    // timelimit for request served by this proxy. 0 is for no timeout
	proxyPrefix     string // The prefix where we mount the imageproxy. default: imageproxy. default: imageproxy.
	referrers       []string
	allowList       []string
	refuselist      []string
	m               Metrics
	log             zerolog.Logger
	enableResize    bool
	singleSetCache  singleflight.Group
	// orgs         domain.OrgConfigRetriever
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		cacheDir:    "/tmp/imageproxy",
		timeOut:     10,
		proxyPrefix: "imageproxy",
		log:         zerolog.Nop(),
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

	s.client = http.Client{Timeout: time.Duration(s.timeOut) * time.Second}

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

	log.Printf("deepzoom command: %s", cmd.String())

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

	if s.lruCache != nil {
		if s.lruCache.Contains(req.CacheKey) {
			data, ok := s.lruCache.Get(req.CacheKey)
			if ok {
				req.CacheType = Lru

				_, err := fmt.Fprintf(w, "%s", data)

				s.m.IncLruCache()

				return err
			}
		}
	}

	isCached := existsInCache(req.cacheKeyPath())

	if !isCached {
		hasSource := existsInCache(req.downloadedSourcePath())
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
					return nil, resizeExternally(req.downloadedSourcePath(), req.cacheKeyPath(), req.thumbnailOpts)
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
					return nil, deepZoomExternally(req.downloadedSourcePath())
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

		_, err = io.Copy(w, tee)
		if err != nil {
			s.log.Error().Err(err).Str("url", req.SourceURL).Msg("error copying data from cache")
			s.m.IncError()

			return err
		}

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
	if err != nil || resp.StatusCode != http.StatusOK {
		s.log.Error().Err(err).Str("url", req.SourceURL).Msg("unable to make remote request")
		s.m.IncRemoteRequestError()

		return err
	}

	defer resp.Body.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	contentType := resp.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "text/xml") && bytes.Contains(buf.Bytes(), []byte("adlibXML")) {
		// don't cache adlib error messages
		s.log.Warn().Str("url", req.SourceURL).Msg("adlib error retrieving image")
		s.m.IncRemoteRequestError()

		return fmt.Errorf("unable to retrieve adlib result")
	}

	// check for adlib error when content type is xml
	err = req.Write(req.downloadedSourcePath(), tee)
	if err != nil {
		// do not return error here or cache write error
		s.log.Error().Err(err).Msgf("unable to write remote file to cache; %s", err)

		s.m.IncRemoteRequestError()
	}

	s.m.IncSource()

	return nil
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

// func (s *Service) SetOrganizationService(svc domain.Service) error {
// // s.organizations = svc
// // do nothing because can't set itself
// return nil
// }

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// needed to implement ikuzo service interface
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

// func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
// s.log = b.Logger.With().Str("svc", "imageproxy").Logger()
// s.orgs = b.Orgs
// }
