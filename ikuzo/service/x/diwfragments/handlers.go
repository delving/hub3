package diwfragments

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/delving/hub3/ikuzo/domain"
)

// maxBulkBodyBytes caps the size of a bulk-ingest request body.
// handleBulkPut sits behind no authentication (per the platform's current
// posture for this route) and buffers every scanned line into an in-memory
// fragments slice before validating or storing any of it — with no cap, an
// unauthenticated caller could stream an arbitrarily large body and exhaust
// server memory. 64 MiB is generous for a legitimate fragment batch from the
// render worker but bounds the damage a hostile or misbehaving client can do.
const maxBulkBodyBytes = 64 << 20

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

// handleBulkPut ingests NDJSON fragments from the render worker. The
// whole batch is validated before anything is stored so a partial write
// cannot leave a collection half-refreshed; orgID and collection always
// come from the request, never from the payload.
func (s *Service) handleBulkPut(w http.ResponseWriter, r *http.Request) {
	org, ok := domain.GetOrganization(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown organization")
		return
	}
	collection := chi.URLParam(r, "collection")
	r.Body = http.MaxBytesReader(w, r.Body, maxBulkBodyBytes)
	var fragments []Fragment
	scanner := bufio.NewScanner(r.Body)
	scanner.Buffer(make([]byte, 0, 1024*1024), 8*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}
		var f Fragment
		if err := json.Unmarshal(raw, &f); err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("line %d: %s", line, err))
			return
		}
		f.OrgID = org.RawID()
		f.Collection = collection
		if err := f.Validate(); err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("line %d: %s", line, err))
			return
		}
		fragments = append(fragments, f)
	}
	if err := scanner.Err(); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("read body: %s", err))
		return
	}
	if err := s.store.Put(r.Context(), fragments); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "fragment store unavailable")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"stored": %d}`, len(fragments))
}
