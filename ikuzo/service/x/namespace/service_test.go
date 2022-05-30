// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package namespace

import (
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/matryer/is"
)

const defaultListSize = 2021

func TestService_SearchLabel(t *testing.T) {
	dc := &domain.Namespace{
		Base:   "http://purl.org/dc/elements/1.1/",
		Prefix: "dc",
	}

	type args struct {
		uri string
	}

	tests := []struct {
		name    string
		ns      *domain.Namespace
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
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			s := &Service{}
			err := s.Set(tt.ns)
			if err != nil {
				t.Errorf("Service.SearchLabel() unexpected error = %v", err)
				return
			}

			// add alternative
			_, err = s.Put("dce", dc.Base)
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
		options []ServiceOptionFunc
	}

	tests := []struct {
		name     string
		args     args
		loadedNS int
		wantErr  bool
	}{
		{
			"loaded without defaults",
			args{[]ServiceOptionFunc{}},
			0,
			false,
		},
		{
			"loaded with defaults",
			args{
				[]ServiceOptionFunc{
					WithDefaults(),
				},
			},
			defaultListSize,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Len() != tt.loadedNS {
				t.Errorf("NewService() %s = %v, want %v", tt.name, got.Len(), tt.loadedNS)
			}
		})
	}
}

func TestService_Add(t *testing.T) {
	svc, err := NewService()
	if err != nil {
		t.Errorf("Unable to start namespace Service; %#v", err)
	}

	type args struct {
		prefix string
		base   string
	}

	tests := []struct {
		name      string
		args      args
		stored    int
		prefixes  int
		temporary bool
		wantErr   bool
	}{
		{
			"empty base",
			args{prefix: "dc", base: ""},
			0,
			0,
			false,
			true,
		},
		{
			"empty prefix",
			args{prefix: "", base: "http://purl.org/dc/elements/1.1/"},
			1,
			1,
			true,
			false,
		},
		{
			"setting default over temporary",
			args{prefix: "dc", base: "http://purl.org/dc/elements/1.1/"},
			1,
			1,
			false,
			false,
		},
		{
			"adding the same pair again",
			args{prefix: "dc", base: "http://purl.org/dc/elements/1.1/"},
			1,
			1,
			false,
			false,
		},
		{
			"adding the alternative prefix",
			args{prefix: "dce", base: "http://purl.org/dc/elements/1.1/"},
			1,
			2,
			false,
			false,
		},
		{
			"adding new namespace pair",
			args{prefix: "skos", base: "http://www.w3.org/2004/02/skos/core#"},
			2,
			1,
			false,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ns, err := svc.Put(tt.args.prefix, tt.args.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if ns == nil {
				t.Errorf("Service.Add() namespace should not be nil; %#v", ns)
			}

			if svc.Len() != tt.stored {
				t.Errorf("Service.Add() stored %d, want %d", svc.Len(), tt.stored)
			}

			if ns.Temporary != tt.temporary {
				t.Errorf("Service.Add() temporary %v, want %v", ns.Temporary, tt.temporary)
			}

			if len(ns.Prefixes()) != tt.prefixes {
				t.Errorf("Service.Add() number of prefixes %v, want %v", len(ns.Prefixes()), tt.prefixes)
			}
		})
	}
}

// nolint:gocritic
func TestListDelete(t *testing.T) {
	is := is.New(t)

	svc, err := NewService(WithDefaults())
	is.NoErr(err)

	namespaces, err := svc.List()
	is.NoErr(err)

	is.Equal(len(namespaces), defaultListSize)

	first := namespaces[0]

	err = svc.Delete(first.GetID())
	is.NoErr(err)

	namespaces, err = svc.List()
	is.NoErr(err)

	is.Equal(len(namespaces), defaultListSize-1)
}

func TestDefaults(t *testing.T) {
	is := is.New(t)

	svc, err := NewService(WithDefaults())
	is.NoErr(err)
	is.Equal(svc.Len(), defaultListSize)

	ns, err := svc.GetWithBase("http://schema.org/")
	is.NoErr(err)
	t.Logf("ns: %#v", ns)
	is.Equal(ns.Prefix, "schema")

	// ns, err = svc.GetWithPrefix("sdo")
	// is.NoErr(err)
	// is.Equal(ns.Prefix, "schema")
}
