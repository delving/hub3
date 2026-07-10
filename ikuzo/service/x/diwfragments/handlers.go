package diwfragments

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/delving/hub3/ikuzo/domain"
)

// writeJSONError writes an error response as a JSON body. http.Error would
// force Content-Type: text/plain, breaking API consumers that parse every
// /api/ui/v1 response as JSON — so all error paths go through here instead.
func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

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
		writeJSONError(w, http.StatusNotFound, "unknown organization")
		return
	}
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "nl"
	}
	fragment, err := s.store.Get(r.Context(), org.RawID(), chi.URLParam(r, "collection"), kind, recordID, lang)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "fragment store unavailable")
		return
	}
	if fragment == nil {
		writeJSONError(w, http.StatusNotFound, "fragment not found")
		return
	}
	// RFC 7232 requires entity-tags to be quoted strings on the wire;
	// Fragment.ETag() stays a bare hex string because quoting is an
	// HTTP-layer concern, applied here.
	etag := strconv.Quote(fragment.ETag())
	if r.Header.Get("If-None-Match") == etag {
		// RFC 7232 section 4.1: a 304 must resend the headers a cache
		// would need to update its stored response.
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "public, max-age=300")
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
	writeJSONError(w, http.StatusNotImplemented, "not implemented")
}
