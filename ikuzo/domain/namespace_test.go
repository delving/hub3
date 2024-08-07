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

package domain_test

import (
	"fmt"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

const (
	dcTitle = "http://purl.org/dc/elements/1.1/title"
	dcNS    = "http://purl.org/dc/elements/1.1/"
	dcAltNS = "http://purl.org/dc/elements/1.2/"
)

func TestSplitURI(t *testing.T) {
	type args struct {
		uri string
	}

	tests := []struct {
		name     string
		args     args
		wantBase string
		wantName string
	}{
		{
			"split by /",
			args{dcTitle},
			dcNS,
			"title",
		},
		{
			"split by #",
			args{"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			"http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"type",
		},
		{
			"unable to split URI",
			args{"urn:123"},
			"",
			"urn:123",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotBase, gotName := domain.SplitURI(tt.args.uri)
			if gotBase != tt.wantBase {
				t.Errorf("SplitURI() gotBase = %v, want %v", gotBase, tt.wantBase)
			}
			if gotName != tt.wantName {
				t.Errorf("SplitURI() gotName = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func ExampleSplitURI() {
	fmt.Println(domain.SplitURI("http://purl.org/dc/elements/1.1/subject"))
	// output: http://purl.org/dc/elements/1.1/ subject
}

// nolint:gocritic
func TestURI_String(t *testing.T) {
	is := is.New(t)

	is.Equal(
		domain.URI(dcTitle).String(),
		dcTitle,
	)
}

func TestNameSpace_AddPrefix(t *testing.T) {
	type fields struct {
		UUID      string
		Base      string
		Prefix    string
		BaseAlt   []string
		PrefixAlt []string
		Schema    string
		Temporary bool
	}

	type args struct {
		prefix string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *domain.Namespace
		wantErr bool
	}{
		{
			"add same prefix",
			fields{Prefix: "dc", PrefixAlt: []string{"dc"}},
			args{"dc"},
			&domain.Namespace{Prefix: "dc", PrefixAlt: []string{"dc"}},
			false,
		},
		{
			"add alt prefix",
			fields{Prefix: "dc", PrefixAlt: []string{"dc"}},
			args{"dct"},
			&domain.Namespace{Prefix: "dc", PrefixAlt: []string{"dc", "dct"}},
			false,
		},
		{
			"correct temporary prefix",
			fields{Prefix: "x123", Temporary: true},
			args{"dc"},
			&domain.Namespace{Prefix: "dc"},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ns := &domain.Namespace{
				ID:      tt.fields.UUID,
				Base:      tt.fields.Base,
				Prefix:    tt.fields.Prefix,
				BaseAlt:   tt.fields.BaseAlt,
				PrefixAlt: tt.fields.PrefixAlt,
				Schema:    tt.fields.Schema,
				Temporary: tt.fields.Temporary,
			}
			if err := ns.AddPrefix(tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("NameSpace.AddPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, ns); diff != "" {
				t.Errorf("NameSpace.AddPrefix() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestNameSpace_AddBase(t *testing.T) {
	type fields struct {
		Base      string
		Prefix    string
		BaseAlt   []string
		PrefixAlt []string
	}

	type args struct {
		base string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *domain.Namespace
		wantErr bool
	}{
		{
			"add same base",
			fields{
				Prefix:  "dc",
				Base:    dcNS,
				BaseAlt: []string{dcNS},
			},
			args{dcNS},
			&domain.Namespace{
				Prefix:  "dc",
				Base:    dcNS,
				BaseAlt: []string{dcNS},
			},
			false,
		},
		{
			"add alt base",
			fields{
				Prefix:  "dc",
				Base:    dcNS,
				BaseAlt: []string{dcNS},
			},
			args{dcAltNS},
			&domain.Namespace{
				Prefix: "dc",
				Base:   dcNS,
				BaseAlt: []string{
					dcNS,
					dcAltNS,
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ns := &domain.Namespace{
				Base:      tt.fields.Base,
				Prefix:    tt.fields.Prefix,
				BaseAlt:   tt.fields.BaseAlt,
				PrefixAlt: tt.fields.PrefixAlt,
			}

			if err := ns.AddBase(tt.args.base); (err != nil) != tt.wantErr {
				t.Errorf("NameSpace.AddBase() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, ns); diff != "" {
				t.Errorf("NameSpace.AddBase() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestNameSpace_GetID(t *testing.T) {
	type fields struct {
		UUID      string
		Base      string
		Prefix    string
		BaseAlt   []string
		PrefixAlt []string
		Schema    string
	}

	tests := []struct {
		name   string
		fields fields
	}{
		{
			"known uuid",
			fields{UUID: "123", Prefix: "dc"},
		},
		{
			"unknown uuid",
			fields{Prefix: "dc"},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ns := &domain.Namespace{
				ID:      tt.fields.UUID,
				Base:      tt.fields.Base,
				Prefix:    tt.fields.Prefix,
				BaseAlt:   tt.fields.BaseAlt,
				PrefixAlt: tt.fields.PrefixAlt,
				Schema:    tt.fields.Schema,
			}
			if got := ns.GetID(); got == "" {
				t.Errorf("NameSpace.GetID() = %v, it should not be empty", got)
			}
		})
	}
}

func TestNameSpace_Merge(t *testing.T) {
	type fields struct {
		Base      string
		Prefix    string
		BaseAlt   []string
		PrefixAlt []string
	}

	type args struct {
		other *domain.Namespace
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		prefixAlt []string
		baseAlt   []string
		wantErr   bool
	}{
		{
			"merge without overlap",
			fields{dcNS, "dc", []string{}, []string{}},
			args{&domain.Namespace{
				Base:      dcAltNS,
				Prefix:    "dce",
				BaseAlt:   []string{},
				PrefixAlt: []string{},
			}},
			[]string{"dc", "dce"},
			[]string{dcNS, dcAltNS},
			false,
		},
		{
			"merge with prefix overlap",
			fields{dcNS, "dc", []string{}, []string{}},
			args{&domain.Namespace{
				Base:      dcAltNS,
				Prefix:    "dc",
				BaseAlt:   []string{},
				PrefixAlt: []string{},
			}},
			[]string{"dc"},
			[]string{dcNS, dcAltNS},
			false,
		},
		{
			"merge with base overlap",
			fields{dcNS, "dc", []string{}, []string{}},
			args{&domain.Namespace{
				Base:      dcNS,
				Prefix:    "dce",
				BaseAlt:   []string{},
				PrefixAlt: []string{},
			}},
			[]string{"dc", "dce"},
			[]string{dcNS},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ns := &domain.Namespace{
				Base:      tt.fields.Base,
				Prefix:    tt.fields.Prefix,
				BaseAlt:   tt.fields.BaseAlt,
				PrefixAlt: tt.fields.PrefixAlt,
			}
			if err := ns.Merge(tt.args.other); (err != nil) != tt.wantErr {
				t.Errorf("NameSpace.Merge() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(tt.prefixAlt, ns.Prefixes()) {
				t.Errorf("NameSpace.Merge() got %v; want %v", ns.Prefixes(), tt.prefixAlt)
			}

			if !cmp.Equal(tt.baseAlt, ns.BaseURIs()) {
				t.Errorf("NameSpace.Merge() got %v; want %v", ns.BaseURIs(), tt.baseAlt)
			}

			if diff := cmp.Diff(fmt.Sprintf("%s: %s", tt.fields.Prefix, tt.fields.Base), ns.String()); diff != "" {
				t.Errorf("NameSpace.Merg() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
