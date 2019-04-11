package namespace

import (
	"reflect"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	store := newMemoryStore()
	if store.Len() != 0 {
		t.Errorf("memoryStore should be empty when initialised; got %d", store.Len())
	}

	tests := []struct {
		name     string
		ns       *NameSpace
		f        func(ns *NameSpace) error
		nrStored int
		wantErr  bool
	}{
		{
			"add first",
			&NameSpace{Base: "http://purl.org/dc/elements/1.1/", Prefix: "dc"},
			store.Set,
			1,
			false,
		},
		{
			"add duplicate",
			&NameSpace{Base: "http://purl.org/dc/elements/1.1/", Prefix: "dc"},
			store.Set,
			1,
			false,
		},
		//{
		//"replace prefix",
		//&NameSpace{Base: URI("http://purl.org/dc/elements/1.1/"), Prefix: "dce"},
		//store.Set,
		//1,
		//false,
		//},
		{
			"add second",
			&NameSpace{Base: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"},
			store.Set,
			2,
			false,
		},
		{
			"delete first",
			&NameSpace{Base: "http://purl.org/dc/elements/1.1/", Prefix: "dc"},
			store.Delete,
			1,
			false,
		},
		{
			"delete second",
			&NameSpace{Base: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"},
			store.Delete,
			0,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f(tt.ns)
			if err != nil && tt.wantErr == false {
				t.Errorf("did not expect error: %#v", err)
			}
			if store.Len() != tt.nrStored {
				t.Errorf("expected %d stored: got %d", tt.nrStored, store.Len())
			}
		})
	}
}

func TestGetFromMemoryStore(t *testing.T) {
	store := newMemoryStore()
	if store.Len() != 0 {
		t.Errorf("memoryStore should be empty when initialised; got %d", store.Len())
	}
	rdf := &NameSpace{Base: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"}
	dc := &NameSpace{Base: "http://purl.org/dc/elements/1.1/", Prefix: "dc"}
	unknown := &NameSpace{Prefix: "unknown"}

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
		case ErrNameSpaceNotFound:
		default:
			t.Errorf("Unexpected error: %#v", err)
		}
	}

}
