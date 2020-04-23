package search

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewService(t *testing.T) {
	type args struct {
		options []OptionFunc
	}
	tests := []struct {
		name    string
		args    args
		want    *Service
		wantErr bool
	}{
		{
			"with no options",
			args{},
			&Service{
				responseSize:    16,
				maxResponseSize: 500,
				facetSize:       50,
			},
			false,
		},
		{
			"set response size",
			args{
				options: []OptionFunc{ResponseSize(20)},
			},
			&Service{
				responseSize:    20,
				maxResponseSize: 500,
				facetSize:       50,
			},
			false,
		},
		{
			"set response size cannot exceed max size",
			args{
				options: []OptionFunc{ResponseSize(600)},
			},
			&Service{
				responseSize:    500,
				maxResponseSize: 500,
				facetSize:       50,
			},
			false,
		},
		{
			"set error opt",
			args{
				options: []OptionFunc{errorOpt()},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			opt := cmp.AllowUnexported(Service{})
			if diff := cmp.Diff(tt.want, got, opt); diff != "" {
				t.Errorf("NewService() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// errorOpt always returns an error for testing
func errorOpt() OptionFunc {
	return func(s *Service) error {
		return fmt.Errorf("we expect this error")
	}
}
