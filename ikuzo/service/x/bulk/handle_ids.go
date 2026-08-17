package bulk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/go-chi/chi/v5"
	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog/log"
)

// handleIndexIDs streams every indexed hubID for a dataset, one per line.
// The consumer (Narthex) diffs this list against its record registry to
// find records that were sent but silently never landed in the index —
// index_verify reports only counts, so which ids are missing can only be
// determined by the party that knows the expected set.
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

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	scroll := index.ESClient().Scroll(c.Config.ElasticSearch.GetIndexName(orgID)).
		Query(q).
		FetchSourceContext(elastic.NewFetchSourceContext(true).Include("meta.hubID")).
		Size(2500)
	defer scroll.Clear(r.Context())

	seen := 0

	for {
		res, err := scroll.Do(r.Context())
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Error().Err(err).Str("orgID", orgID).Str("datasetID", datasetID).
				Msg("index-ids scroll failed")

			if seen == 0 {
				http.Error(w, fmt.Sprintf("scroll failed: %v", err), http.StatusInternalServerError)
			}

			return
		}

		for _, hit := range res.Hits.Hits {
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
	}

	log.Info().Str("orgID", orgID).Str("datasetID", datasetID).Int("ids", seen).
		Msg("index-ids stream served")
}
