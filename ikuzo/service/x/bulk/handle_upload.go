package bulk

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain"
)

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func (s *Service) Handle(w http.ResponseWriter, r *http.Request) {
	p := s.NewParser()

	if err := p.Parse(r.Context(), r.Body); err != nil {
		log.Error().Err(err).Msg("issue with bulk request")
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if len(s.postHooks) != 0 && len(p.postHooks) != 0 {
		applyHooks, ok := s.postHooks[p.stats.OrgID]
		if ok {
			go func() {
				for _, hook := range applyHooks {
					validHooks := []*domain.PostHookItem{}

					for _, ph := range p.postHooks {
						if hook.Valid(ph.DatasetID) {
							validHooks = append(validHooks, ph)
						}
					}

					if err := hook.Publish(validHooks...); err != nil {
						log.Error().Err(err).Msg("unable to submit posthooks")
					}

					log.Debug().Int("nr_hooks", len(validHooks)).Msg("submitted posthooks")
				}
			}()
		}
	}

	render.Status(r, http.StatusCreated)
	log.Info().Msgf("stats: %+v", p.stats)
	render.JSON(w, r, p.stats)
}

func (s *Service) NewParser() *Parser {
	p := &Parser{
		stats:         &Stats{},
		indexTypes:    s.indexTypes,
		bi:            s.index,
		sparqlUpdates: []fragments.SparqlUpdate{},
		s:             s,
		graphs:        map[string]*fragments.FragmentBuilder{},
	}

	if len(s.postHooks) != 0 {
		p.postHooks = []*domain.PostHookItem{}
	}

	return p
}

func (s *Service) GetGraph(hubID string, versionID string) error {
	parts := strings.Split(hubID, "_")
	path := fmt.Sprintf("%s/%s/fg/%s.json", parts[0], parts[1], parts[2])
	_, err := s.mc.GetObject(s.ctx, s.blobCfg.BucketName, path, minio.GetObjectOptions{VersionID: versionID})
	if err != nil {
		return err
	}

	return nil
}
