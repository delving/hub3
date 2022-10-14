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

package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-chi/chi"
	"github.com/mailgun/groupcache"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type esCtxKey int

var esKey esCtxKey

type Proxy struct {
	es    *elasticsearch.Client
	group *groupcache.Group
	log   zerolog.Logger
	cfg   *Config
}

func NewProxy(es *Client) (*Proxy, error) {
	p := &Proxy{
		es:  es.index,
		cfg: es.cfg,
	}

	p.SetLogger(es.cfg.Logger)

	p.group = groupcache.NewGroup(
		"esRemote",
		50*1024*1024,
		groupcache.GetterFunc(p.retrieveFromElasticSearch),
	)

	return p, nil
}

func (p *Proxy) SetLogger(log *zerolog.Logger) {
	p.log = log.With().Str("svc", "esproxy").Logger()
}

func (p *Proxy) requestKey(r *http.Request) string {
	index := chi.URLParam(r, "index")

	var buf bytes.Buffer

	hash := xxhash.New64()
	_, _ = hash.WriteString(index)

	_, err := io.Copy(&buf, io.TeeReader(r.Body, hash))
	if err != nil {
		p.log.Warn().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("unable to copy request body")

		return ""
	}

	r.Body = io.NopCloser(&buf)

	return fmt.Sprintf("%016x", hash.Sum64())
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := p.requestKey(r)

	p.log.Info().Str("requestKey", key).Msg("")

	var data []byte

	ctx := context.WithValue(r.Context(), esKey, r)

	err := p.group.Get(ctx, key, groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		if ctx.Done() != nil {
			p.log.Debug().Err(err).Msg("request was canceled")
			http.Error(w, err.Error(), http.StatusAccepted)

			return
		}

		getErr := fmt.Errorf("error groupcache response: %s", err)
		p.log.Warn().Err(getErr).Msg("")

		http.Error(w, getErr.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(data)
	if err != nil {
		getErr := fmt.Errorf("unable to write elastic response to writer; %w", err)
		p.log.Warn().Err(getErr).Msg("")

		http.Error(w, getErr.Error(), http.StatusInternalServerError)
	}
}

func (p *Proxy) retrieveFromElasticSearch(gctx groupcache.Context, id string, dest groupcache.Sink) error {
	ctx := gctx.(context.Context)
	r := ctx.Value(esKey).(*http.Request)

	var body bytes.Buffer

	var queryBody bytes.Buffer

	_, _ = io.Copy(&body, io.TeeReader(r.Body, &queryBody))

	index := chi.URLParam(r, "index")

	queryStart := time.Now()

	res, err := p.es.Search(
		p.es.Search.WithContext(ctx),
		p.es.Search.WithIndex(index),
		p.es.Search.WithBody(&body),
		p.es.Search.WithTrackTotalHits(true),
	)

	queryEnd := time.Now()

	if err != nil {
		p.log.Warn().Err(err).Msg("unable to get elasticsearch response")
		return err
	}

	defer res.Body.Close()
	defer r.Body.Close()

	var buf bytes.Buffer

	if res.IsError() {
		msg, _ := io.ReadAll(res.Body)
		p.log.Warn().RawJSON("error", msg).Msg("elasticsearch error message")
	}

	size, err := io.Copy(&buf, res.Body)
	if err != nil {
		return err
	}

	requestID, _ := hlog.IDFromRequest(r)

	var query bytes.Buffer
	if _, err := query.ReadFrom(&queryBody); err != nil {
		p.log.Warn().Err(err).Msg("unable to read query body from request")
	}

	p.log.Debug().
		Int("status", res.StatusCode).
		Int64("size", size).
		Str("req_id", requestID.String()).
		Str("query", query.String()).
		Dur("duration", queryEnd.Sub(queryStart)).
		Msg("elastic ead cluster search request")

	return dest.SetBytes(buf.Bytes(), time.Now().Add(20*time.Second))
}

func (p *Proxy) SafeHTTP(w http.ResponseWriter, r *http.Request) {
	// parse the url. Always take the first url for now
	esURL, _ := url.Parse(p.cfg.Urls[0])

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(esURL)

	// strip prefix from path
	r.URL.Path = strings.TrimPrefix(r.URL.EscapedPath(), "/api/es")

	switch {
	case strings.HasSuffix(r.URL.EscapedPath(), "/_analyze") && r.Method == "POST":
		// allow post requests on analyze
	case r.Method != "GET":
		http.Error(w, fmt.Sprintf("method %s is not allowed on esProxy", r.Method), http.StatusBadRequest)
		return
	case r.URL.Path == "/":
		// root is allowed to provide version
	case strings.Contains(r.URL.EscapedPath(), "v2") || strings.Contains(r.URL.EscapedPath(), "v1"):
		// direct access on get is allowed via the proxy on v2 indices
	case !strings.HasPrefix(r.URL.EscapedPath(), "/_cat"):
		http.Error(
			w,
			fmt.Sprintf(
				"path %q is not allowed on esProxy",
				domain.LogUserInput(r.URL.EscapedPath())),
			http.StatusBadRequest,
		)
		return
	}

	if p.cfg.UserName != "" && p.cfg.Password != "" {
		r.SetBasicAuth(p.cfg.UserName, p.cfg.Password)
	}

	// Update the headers to allow for SSL redirection
	r.URL.Host = esURL.Host
	r.URL.Scheme = esURL.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = esURL.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w, r)
}
