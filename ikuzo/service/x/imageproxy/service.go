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
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

type Option func(*Service) error

type Service struct {
	client      http.Client
	cacheDir    string // The path to the imageCache
	timeOut     int    // timelimit for request served by this proxy. 0 is for no timeout
	proxyPrefix string // The prefix where we mount the imageproxy. default: imageproxy. default: imageproxy.
	memoryCache string
	referrers   []string
	blacklist   []string
	// deepzoom    bool     // Enable deepzoom of remote images.
}

func SetCacheDir(path string) Option {
	return func(s *Service) error {
		s.cacheDir = path
		return nil
	}
}

func SetTimeout(duration int) Option {
	return func(s *Service) error {
		s.timeOut = duration
		return nil
	}
}

func SetProxyReferrer(referrer []string) Option {
	return func(s *Service) error {
		s.referrers = referrer
		return nil
	}
}

func SetBlackList(blacklist []string) Option {
	return func(s *Service) error {
		s.blacklist = blacklist
		return nil
	}
}

func SetProxyPrefix(prefix string) Option {
	return func(s *Service) error {
		s.proxyPrefix = prefix
		return nil
	}
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		cacheDir:    "/tmp/imageproxy",
		timeOut:     10,
		proxyPrefix: "imageproxy",
		memoryCache: "memory:500:1h",
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

func (s *Service) Routes() chi.Router {
	router := chi.NewRouter()

	proxyPrefix := fmt.Sprintf("/%s/{options}/*", s.proxyPrefix)
	router.Get(proxyPrefix, s.proxyImage)

	return router
}

func (s *Service) Do(ctx context.Context, req *Request, w io.Writer) error {
	_ = ctx
	cachePath := filepath.Join(s.cacheDir, req.sourcePath())
	// check cache
	r, err := req.Read(cachePath)
	if err != nil && !errors.Is(err, ErrCacheKeyNotFound) {
		log.Error().Err(err).Str("cmp", "imageproxy").Msg("unexpected error reading from cache")
		return err
	}

	if !errors.Is(err, ErrCacheKeyNotFound) {
		defer r.Close()

		_, err = io.Copy(w, r)
		if err != nil {
			log.Error().Err(err).Str("cmp", "imageproxy").Msg("error copying image from cache")
			return err
		}

		return nil
	}

	// make request
	proxyRequest, err := req.GET()
	if err != nil {
		log.Error().Err(err).Str("cmp", "imageproxy").Str("url", req.sourceURL).Msg("unable to create GET request")
		return err
	}

	resp, err := s.client.Do(proxyRequest)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Error().Err(err).Str("cmp", "imageproxy").Str("url", req.sourceURL).Msg("unable to make remote request")
		return err
	}

	defer resp.Body.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	// copy to response writer
	_, err = io.Copy(w, tee)
	if err != nil {
		log.Error().Err(err).Str("cmp", "imageproxy").Msg("error copying remote image")

		return err
	}

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/xml") && bytes.Contains(buf.Bytes(), []byte("adlibXML")) {
		// don't cache adlib error messages
		log.Warn().Str("cmp", "imageproxy").Str("url", req.sourceURL).Msg("adlib error retrieving image")
		return fmt.Errorf("unable to retrieve adlib result")
	}
	// check for adlib error when content type is xml

	err = req.Write(cachePath, &buf)
	if err != nil {
		// do not return error here or cache write error
		log.Error().Err(err).Str("cmp", "imageproxy").Msg("unable to write remote file to cache")
	}

	return nil
}

// create handler fuction to serve the proxied images
func (s *Service) proxyImage(w http.ResponseWriter, r *http.Request) {
	url := chi.URLParam(r, "*")

	// add referer
	if len(s.referrers) != 0 {
		var allowed bool

		for _, referrer := range s.referrers {
			if strings.Contains(r.Referer(), referrer) {
				allowed = true
				break
			}
		}

		if !allowed {
			http.Error(w, "not found", http.StatusNotFound)

			return
		}
	}

	req, err := NewRequest(
		url,
		SetRawQueryString(r.URL.RawQuery),
	)
	if err != nil {
		log.Error().Err(err).Str("cmp", "imageproxy").Str("url", url).Msg("unable to create proxy request")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if len(s.blacklist) != 0 {
		for _, uri := range s.blacklist {
			if strings.Contains(req.sourceURL, uri) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}
	}

	err = s.Do(r.Context(), req, w)
	if err != nil {
		log.Error().Err(err).Str("cmp", "imageproxy").Str("url", req.sourceURL).Msg("unable to make proxy request")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// needed to implement ikuzo service interface
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}
