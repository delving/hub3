package lod

import (
	"fmt"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/delving/hub3/ikuzo/render"
)

type Request struct {
	URI    string
	Format string
}

func (s *Service) handleResolve(w http.ResponseWriter, r *http.Request) {
	store := r.URL.Query().Get("store")
	if store == "" {
		store = s.defaultStore
	}

	resolver, ok := s.stores[store]
	if !ok {
		render.Error(w, r, fmt.Errorf("unknown lod store: %q", store), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})

		return
	}

	uri := r.URL.Query().Get("uri")
	if uri == "" {
		render.Error(w, r, fmt.Errorf("uri param cannot be empty"), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})

		return
	}

	orgID := domain.GetOrganizationID(r)

	subj, err := rdf.NewIRI(uri)
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})

		return
	}

	g, err := resolver.Resolve(r.Context(), orgID, rdf.Subject(subj))
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})

		return
	}

	if err := ntriples.Serialize(g, w); err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusInternalServerError,
		})

		return
	}

	render.NTriples(w, r, "")
}
