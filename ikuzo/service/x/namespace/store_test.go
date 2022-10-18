package namespace

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

	dc := &domain.Namespace{URI: "http://purl.org/dc/elements/1.1/", Prefix: "dc"}
	rdf := &domain.Namespace{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"}

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

	rdf := &domain.Namespace{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Prefix: "rdf"}
	dc := &domain.Namespace{URI: "http://purl.org/dc/elements/1.1/", Prefix: "dc"}
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

	ns2, err := store.GetWithBase(rdf.URI)
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

	namespaces, err := store.List(nil)
	is.NoErr(err)
	is.Equal(len(namespaces), 2)
}

func TestDuplicateURIS(t *testing.T) {
	is := is.New(t)

	svc, err := NewService(WithDefaults())
	is.NoErr(err)
	is.Equal(svc.Len(), defaultListSize)

	nsStore := svc.store.(*namespaceStore)
	suspect := []string{}

	for base, nsList := range nsStore.base2prefix {
		if len(nsList) > 1 {
			var isSet bool
			for _, ns := range nsList {
				if ns.Weight > 20 {
					isSet = true
					break
				}
			}
			if !isSet {
				suspect = append(suspect, base)
			}
		}
	}

	// TODO(kiivihal): finish this in default
	t.Logf("suspect uris (%d): %#v", len(suspect), suspect)
	is.Equal(len(suspect), 107)
}
