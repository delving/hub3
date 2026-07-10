package diwfragments

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/delving/hub3/ikuzo/domain"
)

// handleItem serves one record-detail fragment.
func (s *Service) handleItem(w http.ResponseWriter, r *http.Request) {
	s.serveFragment(w, r, KindItem, chi.URLParam(r, "id"))
}

// handleListing serves the collection's initial-listing fragment.
func (s *Service) handleListing(w http.ResponseWriter, r *http.Request) {
	s.serveFragment(w, r, KindListing, "")
}

// serveFragment implements the shared GET semantics: org from the request
// domain, lang defaulting to nl, 404 when absent, ETag revalidation, and
// the JSON envelope on 200.
func (s *Service) serveFragment(w http.ResponseWriter, r *http.Request, kind Kind, recordID string) {
	org, ok := domain.GetOrganization(r)
	if !ok {
		http.Error(w, `{"error":"unknown organization"}`, http.StatusNotFound)
		return
	}
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "nl"
	}
	fragment, err := s.store.Get(r.Context(), org.RawID(), chi.URLParam(r, "collection"), kind, recordID, lang)
	if err != nil {
		http.Error(w, `{"error":"fragment store unavailable"}`, http.StatusInternalServerError)
		return
	}
	if fragment == nil {
		http.Error(w, `{"error":"fragment not found"}`, http.StatusNotFound)
		return
	}
	etag := fragment.ETag()
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=300")
	_ = json.NewEncoder(w).Encode(Envelope{HeadTags: fragment.HeadTags, HTML: fragment.HTML, Meta: fragment.Meta})
}

// handleBulkPut is implemented in the bulk-write task.
func (s *Service) handleBulkPut(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}
