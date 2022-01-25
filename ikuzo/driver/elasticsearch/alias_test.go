package elasticsearch

import (
	"io"
	"strings"
	"testing"
)

func Test_getIndexNameFromAlias(t *testing.T) {
	type args struct {
		r io.Reader
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"sample",
			args{strings.NewReader(
				`{
					"logs_20302801" : {
						"aliases" : {
						"2030" : {}
					}
				}`,
			)},
			"logs_20302801",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := getIndexNameFromAlias(tt.args.r); got != tt.want {
				t.Errorf("getIndexNameFromAlias() = %q, want %v", got, tt.want)
			}
		})
	}
}
