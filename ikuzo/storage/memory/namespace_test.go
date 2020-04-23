package memory

import (
	"reflect"
	"testing"

	"github.com/matryer/is"
	"github.com/delving/hub3/ikuzo/domain"
)

func TestNameSpaceStore(t *testing.T) {
	store := NewNameSpaceStore()
	if store.Len() != 0 {
		t.Errorf("memoryStore should be empty when initialized; got %d", store.Len())
	}

	dc := &domain.NameSpace{Base: "http://purl.org/dc/elements/1.1/", Prefix: "dc"}
	rdf := &domain.NameSpace{Base: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"}

	tests := []struct {
		name     string
		ns       *domain.NameSpace
		f        func(ns *domain.NameSpace) error
		nrStored int
		wantErr  bool
	}{
		{
			"add first",
			dc,
			store.Set,
			1,
			false,
		},
		{
			"empty namespace",
			nil,
			store.Set,
			1,
			true,
		},
		{
			"set duplicate",
			dc,
			store.Set,
			1,
			false,
		},
		{
			"add second",
			rdf,
			store.Set,
			2,
			false,
		},
		{
			"delete first",
			dc,
			store.Delete,
			1,
			false,
		},
		{
			"delete second",
			rdf,
			store.Delete,
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
func TestGetFromNameSpaceStore(t *testing.T) {
	is := is.New(t)

	store := NewNameSpaceStore()
	if store.Len() != 0 {
		t.Errorf("memoryStore should be empty when initialized; got %d", store.Len())
	}

	rdf := &domain.NameSpace{Base: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"}
	dc := &domain.NameSpace{Base: "http://purl.org/dc/elements/1.1/", Prefix: "dc"}
	unknown := &domain.NameSpace{Prefix: "unknown"}

	err := store.Set(dc)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}

	err = store.Set(rdf)
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
	is.Equal(err, domain.ErrNameSpaceNotFound)

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
		case domain.ErrNameSpaceNotFound:
		default:
			t.Errorf("Unexpected error: %#v", err)
		}
	}

	namespaces, err := store.List()
	is.NoErr(err)
	is.Equal(len(namespaces), 2)
}
