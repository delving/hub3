package diwfragments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
)

// fakeStore serves canned fragments keyed by DocID.
type fakeStore struct{ byID map[string]*Fragment }

func (s *fakeStore) Put(ctx context.Context, fragments []Fragment) error {
	for i := range fragments {
		f := fragments[i]
		s.byID[f.DocID()] = &f
	}
	return nil
}

func (s *fakeStore) Get(ctx context.Context, orgID, collection string, kind Kind, recordID, lang string) (*Fragment, error) {
	f := Fragment{OrgID: orgID, Collection: collection, Kind: kind, RecordID: recordID, Lang: lang}
	return s.byID[f.DocID()], nil
}

func newTestService(t *testing.T, seed ...Fragment) *Service {
	t.Helper()
	store := &fakeStore{byID: map[string]*Fragment{}}
	if err := store.Put(context.Background(), seed); err != nil {
		t.Fatal(err)
	}
	svc, err := NewService(SetStore(store))
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

// doRequest routes a request through the service with an org in context,
// mirroring hub3's per-domain org middleware.
//
// The org is attached via domain.SetOrganizationInContext, the helper
// domain/organization.go documents as "primarily useful for testing where
// you need to create a context with an organization without an
// http.Request" — this is the real mechanism domain.GetOrganization reads
// (via the unexported orgIDKey{} context key), not a guess.
func doRequest(svc *Service, method, url, ifNoneMatch string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, nil)
	org, _ := domain.NewOrganizationID("demo")
	req = req.WithContext(domain.SetOrganizationInContext(req.Context(), domain.Organization{ID: org}))
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}
	w := httptest.NewRecorder()
	svc.ServeHTTP(w, req)
	return w
}

func seededItem() Fragment {
	return Fragment{
		OrgID: "demo", Collection: "coll1", Kind: KindItem, RecordID: "demo_spec_158",
		Lang: "nl", HeadTags: "<link>", HTML: "<div>item</div>",
		Meta: Meta{RenderedAt: "2026-07-10T12:00:00Z", DiwVersion: "1.0.0", ConfigVersion: "abc"},
	}
}

func TestGetItemServesEnvelopeWithETag(t *testing.T) {
	svc := newTestService(t, seededItem())
	w := doRequest(svc, http.MethodGet, "/api/ui/v1/coll1/item/demo_spec_158?lang=nl", "")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("ETag") == "" {
		t.Fatal("ETag header must be set")
	}
	if cc := w.Header().Get("Cache-Control"); cc != "public, max-age=300" {
		t.Fatalf("unexpected Cache-Control %q", cc)
	}
	var env Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.HTML != "<div>item</div>" || env.Meta.DiwVersion != "1.0.0" {
		t.Fatalf("unexpected envelope %+v", env)
	}
}

// TestGetItemHonorsIfNoneMatch asserts the ETag round-trip: whatever the
// server emits in the ETag header (the RFC 7232 quoted form) is what it
// honors in If-None-Match — the tag is read from a first response rather
// than hardcoded from Fragment.ETag(), which stays a bare hex string.
func TestGetItemHonorsIfNoneMatch(t *testing.T) {
	svc := newTestService(t, seededItem())
	first := doRequest(svc, http.MethodGet, "/api/ui/v1/coll1/item/demo_spec_158", "")
	if first.Code != http.StatusOK {
		t.Fatalf("want 200 on first request, got %d: %s", first.Code, first.Body.String())
	}
	etag := first.Header().Get("ETag")
	if len(etag) < 2 || etag[0] != '"' || etag[len(etag)-1] != '"' {
		t.Fatalf("ETag must be a quoted string per RFC 7232, got %q", etag)
	}
	w := doRequest(svc, http.MethodGet, "/api/ui/v1/coll1/item/demo_spec_158", etag)
	if w.Code != http.StatusNotModified {
		t.Fatalf("want 304, got %d", w.Code)
	}
	if got := w.Header().Get("ETag"); got != etag {
		t.Fatalf("304 must resend ETag %q, got %q", etag, got)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "public, max-age=300" {
		t.Fatalf("304 must resend Cache-Control, got %q", cc)
	}
}

func TestGetListingDefaultsLangNL(t *testing.T) {
	listing := seededItem()
	listing.Kind = KindListing
	listing.RecordID = ""
	listing.HTML = "<div>listing</div>"
	svc := newTestService(t, listing)
	w := doRequest(svc, http.MethodGet, "/api/ui/v1/coll1/listing", "")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetMissingFragmentIs404(t *testing.T) {
	svc := newTestService(t)
	w := doRequest(svc, http.MethodGet, "/api/ui/v1/coll1/item/nope", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("error responses must be JSON, got Content-Type %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("error body must be valid JSON: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("error body must carry an error message, got %q", w.Body.String())
	}
}
