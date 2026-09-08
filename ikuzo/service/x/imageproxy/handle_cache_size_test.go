package imageproxy

import (
	"os"
	"path/filepath"
	"testing"
)

// nolint:gocritic
func TestRemoveCacheFiles(t *testing.T) {
	cacheDir := t.TempDir()
	base := filepath.Join(cacheDir, "d45", "e8f", "f71", "aHR0cHM6Ly9leGFtcGxlLm9yZy9pbWcudGlm")

	mkFile := func(p string) {
		t.Helper()

		if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	source := base
	dzi := base + ".dzi"
	tile1 := base + "_files/10/2_1.jpg"
	tile2 := base + "_files/11/0_0.jpg"
	thumb := base + "_500,smartcrop_tn.jpg"

	for _, p := range []string{source, dzi, tile1, tile2, thumb} {
		mkFile(p)
	}

	s := &Service{cacheDir: cacheDir}

	// evicting a single tile must take the whole pyramid and the .dzi
	// marker with it, but leave the source and thumbnail alone
	if err := s.removeCacheFiles([]string{"2026-09-01+00:00:00 " + tile1}); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{tile1, tile2, dzi, base + "_files"} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed", p)
		}
	}

	for _, p := range []string{source, thumb} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to survive; %s", p, err)
		}
	}

	// evicting a non-tile file removes only that file
	if err := s.removeCacheFiles([]string{"2026-09-01+00:00:00 " + thumb}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(thumb); !os.IsNotExist(err) {
		t.Error("expected thumbnail to be removed")
	}

	if _, err := os.Stat(source); err != nil {
		t.Errorf("expected source to survive; %s", err)
	}

	// already-removed paths and malformed lines must not error, so the
	// eviction loop always makes progress
	if err := s.removeCacheFiles([]string{"2026-09-01+00:00:00 " + tile2, "malformed-line-without-space", ""}); err != nil {
		t.Fatal(err)
	}
}
