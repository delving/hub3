package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewScrollPager(t *testing.T) {
	tests := []struct {
		name string
		want ScrollPager
	}{
		{
			"new pager with defaults",
			ScrollPager{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewScrollPager()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewScrollPager() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
