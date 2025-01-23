package testutil

import (
	"flag"
	"log/slog"
	"os"
	"sync"
	"testing"
)

var (
	debugLog  = flag.Bool("debug", false, "enable debug logging")
	setupOnce sync.Once
)

// TestMainHelper provides a reusable TestMain implementation
func TestMainHelper(m *testing.M) {
	setupOnce.Do(func() {
		flag.Parse()
		if *debugLog {
			opts := &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}
			handler := slog.NewTextHandler(os.Stdout, opts)
			logger := slog.New(handler)
			slog.SetDefault(logger)
		}
	})
	os.Exit(m.Run())
}
