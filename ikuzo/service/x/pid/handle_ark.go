package pid

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/delving/hub3/ikuzo/render"
)

func (s *Service) handleArk() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/ark:") {
			s.renderError(w, r, fmt.Errorf("unknown ark: %s", r.URL.Path), http.StatusNotFound)
			return
		}

		ark := validateARK(r.URL.Path)
		if !strings.HasSuffix(r.URL.Path, ark) {
			http.Redirect(w, r, "/"+ark, http.StatusFound)
			return
		}

		pid, err := s.store.Get(ark)
		if err != nil {
			switch {
			case errors.Is(err, ErrTombstone):
				s.renderError(w, r, err, http.StatusGone)
			case errors.Is(err, ErrPIDNotFound):
				s.renderError(w, r, err, http.StatusNotFound)
			default:
				s.renderError(w, r, err, http.StatusBadRequest)
			}
			return
		}

		if pid.IsShownAt != "" {
			http.Redirect(w, r, pid.IsShownAt, http.StatusFound)
			return
		}

		// return RDF in the future
		render.JSON(w, r, pid)
	}
}
