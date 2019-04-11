package namespace_test

import (
	"testing"

	"github.com/delving/hub3/pkg/namespace"
)

func TestService_SearchLabel(t *testing.T) {

	dc := &namespace.NameSpace{
		Base:   "http://purl.org/dc/elements/1.1/",
		Prefix: "dc",
	}
	type args struct {
		uri string
	}
	tests := []struct {
		name    string
		ns      *namespace.NameSpace
		args    args
		want    string
		wantErr bool
	}{
		{
			"simple add",
			dc,
			args{uri: "http://purl.org/dc/elements/1.1/title"},
			"dc_title",
			false,
		},
		{
			"unknown namespace",
			dc,
			args{uri: "http://purl.org/unknown/elements/1.1/title"},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &namespace.Service{}
			err := s.Set(tt.ns)
			if err != nil {
				t.Errorf("Service.SearchLabel() unexpected error = %v", err)
				return
			}
			got, err := s.SearchLabel(tt.args.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.SearchLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Service.SearchLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	type args struct {
		options []namespace.ServiceOptionFunc
	}
	tests := []struct {
		name     string
		args     args
		loadedNS int
		wantErr  bool
	}{
		{
			"loaded without defaults",
			args{[]namespace.ServiceOptionFunc{}},
			0,
			false,
		},
		{
			"loaded with defaults",
			args{
				[]namespace.ServiceOptionFunc{
					namespace.WithDefaults(),
				},
			},
			2188,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := namespace.NewService(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Len() != tt.loadedNS {
				t.Errorf("NewService() = %v, want %v", got.Len(), tt.loadedNS)
			}
		})
	}
}
