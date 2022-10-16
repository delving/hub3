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

package memory

import (
	"reflect"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/matryer/is"
)

func TestNameSpaceStore(t *testing.T) {
	store := NewNamespaceStore()
	if store.Len() != 0 {
		t.Errorf("memoryStore should be empty when initialized; got %d", store.Len())
	}

	dc := &domain.Namespace{Base: "http://purl.org/dc/elements/1.1/", Prefix: "dc"}
	rdf := &domain.Namespace{Base: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"}

	tests := []struct {
		name     string
		ns       *domain.Namespace
		f        func(ns *domain.Namespace) error
		nrStored int
		wantErr  bool
	}{
		{
			"add first",
			dc,
			store.Put,
			1,
			false,
		},
		{
			"empty namespace",
			nil,
			store.Put,
			1,
			true,
		},
		{
			"set duplicate",
			dc,
			store.Put,
			1,
			false,
		},
		{
			"add second",
			rdf,
			store.Put,
			2,
			false,
		},
		{
			"delete first",
			dc,
			store.delete,
			1,
			false,
		},
		{
			"delete second",
			rdf,
			store.delete,
			0,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			err := tt.f(tt.ns)
			if err != nil && tt.wantErr == false {
				t.Errorf("did not expect error: %#v", err)
			}

			if store.Len() != tt.nrStored {
				t.Errorf("%s = expected %d stored: got %d", tt.name, tt.nrStored, store.Len())
			}
		})
	}
}

// nolint:gocritic
func TestGetFromNamespaceStore(t *testing.T) {
	is := is.New(t)

	store := NewNamespaceStore()
	if store.Len() != 0 {
		t.Errorf("memoryStore should be empty when initialized; got %d", store.Len())
	}

	rdf := &domain.Namespace{Base: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"}
	dc := &domain.Namespace{Base: "http://purl.org/dc/elements/1.1/", Prefix: "dc"}
	unknown := &domain.Namespace{Prefix: "unknown"}

	err := store.Put(dc)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}

	err = store.Put(rdf)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}

	if store.Len() != 2 {
		t.Errorf("memoryStore should have 2 namespaces; got %d", store.Len())
	}

	ns1, err := store.GetWithPrefix(dc.Prefix)
	if err != nil {
		t.Errorf("Unexpected error retrieving namespace: %#v", err)
	}

	if !reflect.DeepEqual(dc, ns1) {
		t.Errorf("GetWithPrefix expected %#v; got %#v", dc, ns1)
	}

	nsErr, err := store.GetWithBase("http://unknown.com/base")
	is.Equal(nsErr, nil)
	is.Equal(err, domain.ErrNamespaceNotFound)

	ns2, err := store.GetWithBase(rdf.Base)
	if err != nil {
		t.Errorf("Unexpected error retrieving namespace: %#v", err)
	}

	if !reflect.DeepEqual(rdf, ns2) {
		t.Errorf("GetWithPrefix expected %#v; got %#v", rdf, ns2)
	}

	_, err = store.GetWithPrefix(unknown.Prefix)
	if err != nil {
		switch err {
		case domain.ErrNamespaceNotFound:
		default:
			t.Errorf("Unexpected error: %#v", err)
		}
	}

	namespaces, err := store.List()
	is.NoErr(err)
	is.Equal(len(namespaces), 2)
}
