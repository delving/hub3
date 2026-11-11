package bulk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/go-chi/chi/v5"
	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog/log"
)

const indexIDsKeepAlive = "1m"

// handleIndexIDs streams every indexed hubID for a dataset, one per line.
// The consumer (Narthex) diffs this list against its record registry to
// find records that were sent but silently never landed in the index —
// index_verify reports only counts, so which ids are missing can only be
// determined by the party that knows the expected set.
//
// It pages with a point in time and search_after rather than a scroll, so
// it does not count against search.max_open_scroll_context.
func (s *Service) handleIndexIDs(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	datasetID := chi.URLParam(r, "datasetID")

	if orgID == "" || datasetID == "" {
		http.Error(w, "orgID and datasetID are required", http.StatusBadRequest)
		return
	}

	q := elastic.NewBoolQuery().Must(
		elastic.NewMatchPhraseQuery(c.Config.ElasticSearch.SpecKey, datasetID),
		elastic.NewTermQuery("meta.docType", fragments.FragmentGraphDocType),
		elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, orgID),
	)

	es := c.Config.ElasticSearch

	pit, err := elasticsearch.NewPITClient(es.Urls, es.UserName, es.Password)
	if err != nil {
		log.Error().Err(err).Str("orgID", orgID).Str("datasetID", datasetID).
			Msg("index-ids unable to create point in time client")
		http.Error(w, fmt.Sprintf("search failed: %v", err), http.StatusInternalServerError)

		return
	}

	pitID, err := pit.CreatePIT(r.Context(), es.GetIndexName(orgID), indexIDsKeepAlive)
	if err != nil {
		log.Error().Err(err).Str("orgID", orgID).Str("datasetID", datasetID).
			Msg("index-ids unable to open point in time")
		http.Error(w, fmt.Sprintf("search failed: %v", err), http.StatusInternalServerError)

		return
	}

	// Close with a context that survives a client disconnect, so an aborted
	// stream does not leave the point in time open until it expires.
	defer func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 10*time.Second)
		defer cancel()

		if closeErr := pit.ClosePIT(ctx, pitID); closeErr != nil {
			log.Error().Err(closeErr).Str("orgID", orgID).Str("datasetID", datasetID).
				Msg("index-ids unable to close point in time")
		}
	}()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	seen, err := writeIndexIDs(w, func(searchAfter []interface{}) ([]*elastic.SearchHit, error) {
		search := index.ESClient().Search().
			Query(q).
			PointInTime(elastic.NewPointInTimeWithKeepAlive(pitID, indexIDsKeepAlive)).
			FetchSourceContext(elastic.NewFetchSourceContext(true).Include("meta.hubID")).
			Sort("meta.hubID", true).
			Size(2500)

		if len(searchAfter) > 0 {
			search = search.SearchAfter(searchAfter...)
		}

		res, err := search.Do(r.Context())
		if err != nil {
			return nil, err
		}

		if res.Hits == nil {
			return nil, nil
		}

		return res.Hits.Hits, nil
	})
	if err != nil {
		log.Error().Err(err).Str("orgID", orgID).Str("datasetID", datasetID).
			Msg("index-ids search failed")

		if seen == 0 {
			http.Error(w, fmt.Sprintf("search failed: %v", err), http.StatusInternalServerError)
		}

		return
	}

	log.Info().Str("orgID", orgID).Str("datasetID", datasetID).Int("ids", seen).
		Msg("index-ids stream served")
}

// writeIndexIDs pages through fetch with search_after until it returns an
// empty page, writing one hubID per line and flushing after every page.
// It returns the number of ids written, also when it fails part way.
func writeIndexIDs(w io.Writer, fetch func(searchAfter []interface{}) ([]*elastic.SearchHit, error)) (int, error) {
	var (
		seen        int
		searchAfter []interface{}
	)

	for {
		hits, err := fetch(searchAfter)
		if err != nil {
			return seen, err
		}

		if len(hits) == 0 {
			return seen, nil
		}

		for _, hit := range hits {
			var doc struct {
				Meta struct {
					HubID string `json:"hubID"`
				} `json:"meta"`
			}
			if err := json.Unmarshal(hit.Source, &doc); err != nil || doc.Meta.HubID == "" {
				continue
			}

			fmt.Fprintln(w, doc.Meta.HubID)
			seen++
		}

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		searchAfter = hits[len(hits)-1].Sort
	}
}
