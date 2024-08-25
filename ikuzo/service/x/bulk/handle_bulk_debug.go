package bulk

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/delving/hub3/ikuzo/service/x/sparql"
)

func (s *Service) handleBulkDebug(w http.ResponseWriter, r *http.Request) {
	cfgID := r.URL.Query().Get("cfg")
	subject := r.URL.Query().Get("uri")
	format := r.URL.Query().Get("format")
	errorConfig := render.DefaultConfig

	if cfgID == "" || subject == "" {
		errorConfig.StatusCode = http.StatusBadRequest
		render.Error(w, r, fmt.Errorf("cfg and uri params must be non-empty"), &errorConfig)
		return
	}

	configPath := filepath.Join(strings.TrimSuffix(s.harvestConfigPath, "/"), cfgID+".toml")
	cfg, err := decodeConfig(configPath)
	if err != nil {
		render.Error(w, r, err, nil)
		return
	}

	g, harvestErr := sparql.HarvestGraph(context.Background(), cfg, subject)
	if harvestErr != nil {
		slog.Info("unable to harvest sparql graphs", "datasetID", cfg.Spec, "error", harvestErr)
		render.Error(w, r, harvestErr, nil)
		return
	}

	orgID := domain.GetOrganizationID(r)

	if format != "" {
		fg := fragments.NewFragmentGraph()
		fg.Meta.EntryURI = subject
		fg.Meta.OrgID = orgID.String()
		fg.Meta.NamedGraphURI = subject + "/graph"

		fb := fragments.NewFragmentBuilder(fg)
		var err error
		fb.Graph, err = g.AsLegacyGraph()
		if err != nil {
			render.Error(w, r, err, &errorConfig)
			return
		}

		if format == "legacyGraph" {
			render.PlainText(w, r, fb.Graph.String())
			return
		}
		rm, err := fb.ResourceMap()
		if err != nil {
			render.Error(w, r, err, &errorConfig)
			return
		}


		if format == "resourceMap" {
			render.JSON(w, r, rm.Resources())
			return
		}

		f, err := fb.Doc()
		if err != nil {
			render.Error(w, r, err, &errorConfig)
			return
		}

		render.JSON(w, r, f)
		return
	}

	if err := ntriples.Serialize(g, w); err != nil {
		render.Error(w, r, err, &errorConfig)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}
