package sitemap

import (
	"reflect"
	"testing"
)

func Test_getMaxPages(t *testing.T) {
	type args struct {
		count int64
	}
	tests := []struct {
		name      string
		args      args
		wantPages []int
	}{
		{
			name:      "single page",
			args:      args{count: 100},
			wantPages: []int{1},
		},
		{
			name:      "two page",
			args:      args{count: 65000},
			wantPages: []int{1, 2},
		},
		{
			name:      "multiple pages",
			args:      args{count: 565000},
			wantPages: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPages := getMaxPages(tt.args.count); !reflect.DeepEqual(gotPages, tt.wantPages) {
				t.Errorf("getMaxPages() = %v, want %v", gotPages, tt.wantPages)
			}
		})
	}
}
