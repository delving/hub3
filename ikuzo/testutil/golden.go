package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Shared update flag that can be imported by all test packages
var Update = flag.Bool("update", false, "update golden files")

// GoldenHelper provides utilities for golden file testing
type GoldenHelper struct {
	T         *testing.T
	PackageID string // Identifies which package the test belongs to
}

// NewGoldenHelper creates a new GoldenHelper
func NewGoldenHelper(t *testing.T, packageID string) *GoldenHelper {
	return &GoldenHelper{
		T:         t,
		PackageID: packageID,
	}
}

// Path returns the full path to a golden file
func (g *GoldenHelper) Path(name string) string {
	return filepath.Join("testdata", g.PackageID, name)
}

// Compare compares actual content with a golden file
func (g *GoldenHelper) Compare(name string, actual []byte) {
	g.T.Helper()
	golden := g.Path(name)

	if *Update {
		g.update(golden, actual)
		return
	}

	expected, err := os.ReadFile(golden)
	if err != nil {
		if os.IsNotExist(err) {
			g.update(golden, actual)
			return
		}
		g.T.Fatalf("failed to read golden file %s: %v", name, err)
	}

	if diff := cmp.Diff(string(expected), string(actual)); diff != "" {
		g.T.Errorf("golden file %s mismatch (-want +got):\n%s", name, diff)
	}
}

func (g *GoldenHelper) update(golden string, actual []byte) {
	g.T.Helper()
	if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
		g.T.Fatalf("failed to create golden file directory: %v", err)
	}
	if err := os.WriteFile(golden, actual, 0o644); err != nil {
		g.T.Fatalf("failed to update golden file: %v", err)
	}
}
