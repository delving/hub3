package imageproxy

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
)

// 1x1 white JPEG, enough bytes to pass the non-empty cache checks.
var tinyJPEG = []byte{
	0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01,
	0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43,
	0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
	0x09, 0x08, 0x0a, 0x0c, 0x14, 0x0d, 0x0c, 0x0b, 0x0b, 0x0c, 0x19, 0x12,
	0x13, 0x0f, 0x14, 0x1d, 0x1a, 0x1f, 0x1e, 0x1d, 0x1a, 0x1c, 0x1c, 0x20,
	0x24, 0x2e, 0x27, 0x20, 0x22, 0x2c, 0x23, 0x1c, 0x1c, 0x28, 0x37, 0x29,
	0x2c, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1f, 0x27, 0x39, 0x3d, 0x38, 0x32,
	0x3c, 0x2e, 0x33, 0x34, 0x32, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
	0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4, 0x00, 0x1f, 0x00, 0x00,
	0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	0x09, 0x0a, 0x0b, 0xff, 0xc4, 0x00, 0xb5, 0x10, 0x00, 0x02, 0x01, 0x03,
	0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7d,
	0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
	0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xa1, 0x08,
	0x23, 0x42, 0xb1, 0xc1, 0x15, 0x52, 0xd1, 0xf0, 0x24, 0x33, 0x62, 0x72,
	0x82, 0xff, 0xda, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3f, 0x00, 0xfb,
	0xd0, 0xff, 0xd9,
}

// captureServer records the Cache-Control header of every inbound request.
func captureServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()

	var (
		mu   sync.Mutex
		seen []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Header.Get("Cache-Control"))
		mu.Unlock()

		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(tinyJPEG)
	}))
	t.Cleanup(srv.Close)

	return srv, &seen
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	s, err := NewService(SetCacheDir(t.TempDir()), SetLruCacheSize(16))
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	return s
}

// TestStoreSource_ForceRefreshReachesUpstream verifies that the ForceRefresh
// flag set on a recache request results in no-cache headers on the actual
// outgoing HTTP request, and that a normal fetch sends none.
func TestStoreSource_ForceRefreshReachesUpstream(t *testing.T) {
	srv, seen := captureServer(t)
	s := newTestService(t)

	normal, err := NewRequest(srv.URL+"/img.jpg", SetService(s))
	if err != nil {
		t.Fatalf("NewRequest normal: %v", err)
	}
	if err := s.storeSource(normal); err != nil {
		t.Fatalf("storeSource normal: %v", err)
	}

	forced, err := NewRequest(srv.URL+"/img.jpg", SetService(s))
	if err != nil {
		t.Fatalf("NewRequest forced: %v", err)
	}
	forced.ForceRefresh = true
	if err := s.storeSource(forced); err != nil {
		t.Fatalf("storeSource forced: %v", err)
	}

	if len(*seen) != 2 {
		t.Fatalf("expected 2 upstream requests, got %d: %v", len(*seen), *seen)
	}
	if got := (*seen)[0]; got != "" {
		t.Errorf("normal fetch should not send Cache-Control, got %q", got)
	}
	if got := (*seen)[1]; got != "no-cache" {
		t.Errorf("force-refresh fetch Cache-Control = %q, want %q", got, "no-cache")
	}
}

// TestRecacheHandler_SetsForceRefresh drives the full proxy route and confirms
// that the second (recache) call sends no-cache upstream while the priming raw
// call does not.
func TestRecacheHandler_SetsForceRefresh(t *testing.T) {
	srv, seen := captureServer(t)
	s := newTestService(t)
	s.enableResize = false // avoid vips dependency in CI

	router := chi.NewRouter()
	s.Routes("", router)

	proxy := httptest.NewServer(router)
	t.Cleanup(proxy.Close)

	target := srv.URL + "/img.jpg" // e.g. http://127.0.0.1:PORT/img.jpg

	// Prime the cache with a raw fetch.
	doGet(t, proxy.URL+"/imageproxy/raw/"+target)
	// Recache: should purge and refetch with no-cache headers.
	doGet(t, proxy.URL+"/imageproxy/recache/"+target)

	if len(*seen) < 2 {
		t.Fatalf("expected at least 2 upstream requests, got %d: %v", len(*seen), *seen)
	}
	if first := (*seen)[0]; first != "" {
		t.Errorf("raw fetch should not send Cache-Control, got %q", first)
	}
	last := (*seen)[len(*seen)-1]
	if last != "no-cache" {
		t.Errorf("recache fetch Cache-Control = %q, want %q", last, "no-cache")
	}
}

func doGet(t *testing.T, url string) {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
}
