package semantic

import (
	"testing"
	"time"
)

func TestNewQueryContext(t *testing.T) {
	opts := &QueryOptions{
		Query: &TextQuery{Value: "amsterdam"},
	}

	ctx := NewQueryContext(opts, 12847)

	if ctx.ID == "" {
		t.Error("expected non-empty context ID")
	}

	if len(ctx.ID) > 12 {
		t.Errorf("context ID should be short, got %d chars: %s", len(ctx.ID), ctx.ID)
	}

	if ctx.TotalResults != 12847 {
		t.Errorf("TotalResults = %d, want 12847", ctx.TotalResults)
	}

	if ctx.Query == nil {
		t.Error("expected query to be preserved in context")
	}

	if ctx.IsExpired() {
		t.Error("freshly created context should not be expired")
	}
}

func TestNewQueryContext_UniqueIDs(t *testing.T) {
	opts := &QueryOptions{
		Query: &TextQuery{Value: "test"},
	}

	ctx1 := NewQueryContext(opts, 100)
	ctx2 := NewQueryContext(opts, 100)

	if ctx1.ID == ctx2.ID {
		t.Error("expected unique context IDs")
	}
}

func TestNewQueryContext_IDFormat(t *testing.T) {
	opts := &QueryOptions{
		Query: &TextQuery{Value: "test"},
	}

	ctx := NewQueryContext(opts, 100)

	if ctx.ID[:4] != "ctx_" {
		t.Errorf("expected ID to start with 'ctx_', got %s", ctx.ID)
	}

	// Token should equal ID
	if ctx.Token != ctx.ID {
		t.Errorf("Token = %s, want %s (same as ID)", ctx.Token, ctx.ID)
	}
}

func TestSearchContext_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		ctx := &SearchContext{
			ExpiresAt: time.Now().Add(10 * time.Minute).Format(time.RFC3339),
		}
		if ctx.IsExpired() {
			t.Error("context should not be expired")
		}
	})

	t.Run("expired", func(t *testing.T) {
		ctx := &SearchContext{
			ExpiresAt: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		}
		if !ctx.IsExpired() {
			t.Error("context should be expired")
		}
	})

	t.Run("empty expiry", func(t *testing.T) {
		ctx := &SearchContext{}
		if ctx.IsExpired() {
			t.Error("context with no expiry should not be expired")
		}
	})
}

func TestSearchContext_ExtendTTL(t *testing.T) {
	ctx := &SearchContext{
		ExpiresAt: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
	}

	if !ctx.IsExpired() {
		t.Fatal("context should be expired before extending")
	}

	ctx.ExtendTTL()

	if ctx.IsExpired() {
		t.Error("context should not be expired after extending TTL")
	}
}
