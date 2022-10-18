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

const defaultListSize = 2195

func TestService_SearchLabel(t *testing.T) {
	dc := &domain.Namespace{
		URI:    "http://purl.org/dc/elements/1.1/",
		Prefix: "dc",
		Weight: 100,
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
			_, err = s.Put("dce", dc.URI, 0)
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
		weight int
	}

	tests := []struct {
		name    string
		args    args
		stored  int
		wantErr bool
	}{
		{
			"empty base",
			args{prefix: "dc", base: "", weight: 0},
			0,
			true,
		},
		{
			"empty prefix",
			args{prefix: "", base: "http://purl.org/dc/elements/1.1/", weight: 0},
			0,
			true,
		},
		{
			"adding a new pair",
			args{prefix: "dc", base: "http://purl.org/dc/elements/1.1/", weight: 100},
			1,
			false,
		},
		{
			"adding the same pair again",
			args{prefix: "dc", base: "http://purl.org/dc/elements/1.1/", weight: 100},
			1,
			false,
		},
		{
			"adding the alternative prefix",
			args{prefix: "dce", base: "http://purl.org/dc/elements/1.1/", weight: 0},
			2,
			false,
		},
		{
			"adding new namespace pair",
			args{prefix: "skos", base: "http://www.w3.org/2004/02/skos/core#", weight: 0},
			3,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ns, err := svc.Put(tt.args.prefix, tt.args.base, tt.args.weight)
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
		})
	}
}

// nolint:gocritic
func TestListDelete(t *testing.T) {
	is := is.New(t)

	svc, err := NewService(WithDefaults())
	is.NoErr(err)

	namespaces, err := svc.List(nil)
	is.NoErr(err)

	is.Equal(len(namespaces), defaultListSize)

	first := namespaces[0]

	err = svc.Delete(first.ID)
	is.NoErr(err)

	namespaces, err = svc.List(nil)
	is.NoErr(err)

	is.Equal(len(namespaces), defaultListSize-1)
}

func TestDefaults(t *testing.T) {
	is := is.New(t)

	svc, err := NewService(WithDefaults())
	is.NoErr(err)
	is.Equal(svc.Len(), defaultListSize)

	nsList, err := svc.List(&ListOptions{Prefix: "schema"})
	is.NoErr(err)
	is.Equal(len(nsList), 1)
	t.Logf("ns list schema: %#v", nsList)

	nsList, err = svc.List(&ListOptions{URI: "http://schema.org/"})
	is.NoErr(err)
	is.Equal(len(nsList), 3)
	t.Logf("ns list schema from base: %#v", nsList)
	is.Equal(nsList[len(nsList)-1].Weight, 3)
	is.Equal(nsList[0].Weight, 100)

	ns, err := svc.GetWithBase("http://schema.org/")
	is.NoErr(err)
	t.Logf("ns: %#v", ns)
	is.Equal(ns.Prefix, "schema")

	ns, err = svc.GetWithPrefix("sdo")
	is.NoErr(err)
	is.Equal(ns.Prefix, "sdo")

	ns, err = svc.GetWithPrefix("schema")
	is.NoErr(err)
	is.Equal(ns.Prefix, "schema")
}
