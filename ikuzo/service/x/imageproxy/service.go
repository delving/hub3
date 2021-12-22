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

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"

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
	memoryCache     string
	referrers       []string
	allowList       []string
	refuselist      []string
	m               Metrics
	log             zerolog.Logger
	enableResize    bool
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

	s.client = http.Client{Timeout: time.Duration(s.timeOut) * time.Second}

	return s, nil
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

	log.Printf("sourcePath: %s", req.downloadedSourcePath())
	log.Printf("cacheKey: %s", req.CacheKey)
	log.Printf("sourceURL: %s", req.SourceURL)
	log.Printf("key: %#v", req)

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

	hasSource := existsInCache(req.downloadedSourcePath())
	if !hasSource {
		if err := s.storeSource(req); err != nil {
			return err
		}

		s.log.Info().Str("path", req.downloadedSourcePath()).Msg("storing source path")
		req.CacheType = Source

		hasSource = true
	}

	isCached := existsInCache(req.cacheKeyPath())

	if !isCached && req.thumbnailOpts != "" && hasSource {
		err := resizeExternally(req.downloadedSourcePath(), req.cacheKeyPath(), req.thumbnailOpts)
		if err != nil {
			s.log.Error().Err(err).Str("url", req.SourceURL).
				Str("sourcePath", req.downloadedSourcePath()).
				Str("storePath", req.cacheKeyPath()).
				Msgf("unexpected error creating thumbnail; %s", err)
		}
		req.CacheKey = encodeURL(req.SourceURL)

		s.m.IncResize()
	}

	if !isCached && req.SubPath == deepZoomSuffix {
		err := deepZoomExternally(req.downloadedSourcePath())
		if err != nil {
			s.log.Error().Err(err).Str("url", req.SourceURL).Msg("unexpected error creating deepzoom")
			s.m.IncError()
			return err
		}

		s.m.IncDeepZoom()
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
		s.log.Error().Err(err).Msg("unable to write remote file to cache")

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

// create handler fuction to serve the proxied images
func (s *Service) proxyImage(w http.ResponseWriter, r *http.Request) {
	targetURL := chi.URLParam(r, "*")
	options := chi.URLParam(r, "options")

	allowed, err := s.domainAllowed(targetURL)
	if err != nil {
		s.m.IncError()
		s.log.Error().Err(err).Str("url", targetURL).Msg("unable to check allowed domains")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if !allowed {
		s.m.IncRejectDomain()
		s.log.Error().Err(err).Str("url", targetURL).Msg("domain not allowed")
		http.Error(w, "domain is now allowed", http.StatusForbidden)

		return
	}

	allowed = s.reffererAllowed(r.Referer())
	if !allowed {
		s.m.IncRejectReferrer()
		s.log.Error().Err(err).Str("url", targetURL).Str("referrer", r.Referer()).Msg("domain not allowed")
		http.Error(w, fmt.Sprintf("referrer not allowed: %s", r.Referer()), http.StatusForbidden)

		return
	}

	req, err := NewRequest(
		targetURL,
		SetRawQueryString(r.URL.RawQuery),
		SetTransform(options),
		SetService(s),
		SetEnableTransform(s.enableResize),
	)
	if err != nil {
		s.log.Error().Err(err).Str("url", targetURL).Msg("unable to create proxy request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.m.IncError()

		return
	}

	switch req.TransformOptions {
	case "explain":
		explain := fmt.Sprintf("%s => %s", req.SourceURL, req.downloadedSourcePath())
		render.PlainText(w, r, explain)

		return
	case "metrics":
		render.JSON(w, r, s.m)
		return
	case "request":
		render.JSON(w, r, req)
		return
	}

	if len(s.refuselist) != 0 {
		for _, uri := range s.refuselist {
			if strings.Contains(req.SourceURL, uri) {
				http.Error(w, "not found", http.StatusNotFound)
				s.m.IncRejectURI()
				return
			}
		}
	}

	var buf bytes.Buffer

	err = s.Do(r.Context(), req, &buf)
	if err != nil {
		s.log.Error().Err(err).Str("url", req.SourceURL).Msg("unable to make proxy request")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Cache-Control", "public,max-age=259200")
	r.Header.Set("Cache-Type", string(req.CacheType))
	r.Header.Set("Cache-Url", req.SourceURL)

	if _, err := io.Copy(w, &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
