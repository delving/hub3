package sparql

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
)

func (s *Service) sparqlProxy(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)
	repo, err := s.GetRepo(orgID)
	if err != nil {
		if errors.Is(err, domain.ErrServiceNotEnabled) {
			render.Error(w, r, fmt.Errorf("sparql: %w", err), &render.ErrorConfig{
				StatusCode: http.StatusNotAcceptable,
			})

			return
		}
		render.Error(w, r, fmt.Errorf("sparql: %w", err), &render.ErrorConfig{
			StatusCode: http.StatusInternalServerError,
		})

		return
	}

	var query string
	switch r.Method {
	case http.MethodGet:
		query = r.URL.Query().Get("query")
	case http.MethodPost:
		query = r.FormValue("query")
	}

	if query == "" {
		render.Error(w, r, fmt.Errorf("sparql query cannot be empty"), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})
		return
	}
	if !strings.Contains(strings.ToLower(query), "limit ") {
		query = fmt.Sprintf("%s LIMIT %d", query, s.queryLimit)
	}

	resp, statusCode, contentType, err := s.runSparqlQuery(repo, query)
	if err != nil {
		render.Error(w, r, fmt.Errorf("error with sparql query"), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
			Message:    string(resp),
		})
		return
	}

	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(resp)
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusInternalServerError,
			Message:    string(resp),
		})
		return
	}
	render.Status(r, statusCode)
}

// runSparqlQuery sends a SPARQL query to the SPARQL-endpoint specified in the configuration
func (s *Service) runSparqlQuery(repo *Repo, query string) (body []byte, statusCode int, contentType string, err error) {
	resp, err := repo.queryRaw(query, http.MethodGet, "")
	if err != nil {
		s.log.Error().Err(err).Msg("Error in sparql query")
		if strings.Contains(err.Error(), "connection refused") {
			body = []byte("triple store unavailable")
		}
		statusCode = http.StatusBadRequest
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		s.log.Error().Err(err).Msg("Unable to read the response body with error: %s")
		return
	}

	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")
	return
}
