package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/go-chi/chi"
	"github.com/mailgun/groupcache"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

type esCtxKey int

var esKey esCtxKey

type Proxy struct {
	es    *elasticsearch.Client
	group *groupcache.Group
}

func NewProxy(es *elasticsearch.Client) (*Proxy, error) {
	p := &Proxy{
		es: es,
	}

	p.group = groupcache.NewGroup(
		"esRemote",
		50*1024*1024,
		groupcache.GetterFunc(p.retrieveFromElasticSearch),
	)

	return p, nil
}

func requestKey(r *http.Request) string {
	index := chi.URLParam(r, "index")

	var buf bytes.Buffer

	hash := xxhash.New64()
	_, _ = hash.WriteString(index)

	_, err := io.Copy(&buf, io.TeeReader(r.Body, hash))
	if err != nil {
		log.Warn().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("unable to copy request body")

		return ""
	}

	r.Body = ioutil.NopCloser(&buf)

	return fmt.Sprintf("%016x", hash.Sum64())
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := requestKey(r)

	log.Info().Str("requestKey", key).Msg("")

	var data []byte

	ctx := context.WithValue(r.Context(), esKey, r)

	err := p.group.Get(ctx, key, groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		if ctx.Done() != nil {
			log.Debug().Err(err).Msg("request was canceled")
			http.Error(w, err.Error(), http.StatusAccepted)

			return
		}

		getErr := fmt.Errorf("error groupcache response: %s", err)
		log.Warn().Err(getErr).Msg("")

		http.Error(w, getErr.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(data)
	if err != nil {
		getErr := fmt.Errorf("unable to write elastic response to writer; %w", err)
		log.Warn().Err(getErr).Msg("")

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
		log.Warn().Err(err).Msg("unable to get elasticsearch response")
		return err
	}

	defer res.Body.Close()
	defer r.Body.Close()

	var buf bytes.Buffer

	size, err := io.Copy(&buf, res.Body)
	if err != nil {
		return err
	}

	requestID, _ := hlog.IDFromRequest(r)

	var query bytes.Buffer
	if _, err := query.ReadFrom(&queryBody); err != nil {
		log.Warn().Err(err).Msg("unable to read query body from request")
	}

	log.Info().
		Int("status", res.StatusCode).
		Int64("size", size).
		Str("req_id", requestID.String()).
		Str("query", query.String()).
		Dur("duration", queryEnd.Sub(queryStart)).
		Msg("elastic ead cluster search request")

	return dest.SetBytes(buf.Bytes(), time.Now().Add(20*time.Second))
}
