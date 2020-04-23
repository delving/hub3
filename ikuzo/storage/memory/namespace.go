package memory

import (
	"fmt"
	"sync"

	"github.com/delving/hub3/ikuzo/domain"
)

// NameSpaceStore is the default namespace.Store for namespace.Service.
//
// Note: mutations in this store are ephemeral.
type NameSpaceStore struct {
	sync.RWMutex
	prefix2base map[string]*domain.NameSpace
	base2prefix map[string]*domain.NameSpace
	namespaces  map[string]*domain.NameSpace
}

// NewNameSpaceStore creates an in-memory namespace.Store.
func NewNameSpaceStore() *NameSpaceStore {
	return &NameSpaceStore{
		prefix2base: make(map[string]*domain.NameSpace),
		base2prefix: make(map[string]*domain.NameSpace),
		namespaces:  make(map[string]*domain.NameSpace),
	}
}

// Len returns the number of stored namespaces.
// Alternatives Base or Prefixes don't count towards the total.
func (ms *NameSpaceStore) Len() int {
	return len(ms.namespaces)
}

// Set stores the NameSpace in the Store
func (ms *NameSpaceStore) Set(ns *domain.NameSpace) error {
	if ns == nil {
		return fmt.Errorf("cannot store empty namespace")
	}

	// this implementation of Delete can never return an error
	_ = ms.Delete(ns)

	// if you lock earlier you get a deadlock
	ms.Lock()
	defer ms.Unlock()

	for _, prefix := range ns.Prefixes() {
		ms.prefix2base[prefix] = ns
	}

	for _, base := range ns.BaseURIs() {
		ms.base2prefix[base] = ns
	}

	id := ns.GetID()
	ms.namespaces[id] = ns

	return nil
}

// Delete removes a NameSpace from the store
func (ms *NameSpaceStore) Delete(ns *domain.NameSpace) error {
	ms.Lock()
	defer ms.Unlock()

	id := ns.GetID()

	_, ok := ms.namespaces[id]
	if ok {
		delete(ms.namespaces, id)
	}
	// drop all prefixes
	for _, p := range ns.Prefixes() {
		_, ok := ms.prefix2base[p]
		if ok {
			delete(ms.prefix2base, p)
		}
	}

	// drop all base-URIs
	for _, b := range ns.BaseURIs() {
		_, ok := ms.base2prefix[b]
		if ok {
			delete(ms.base2prefix, b)
		}
	}

	return nil
}

// GetWithPrefix returns a NameSpace from the store if the prefix is found.
func (ms *NameSpaceStore) GetWithPrefix(prefix string) (*domain.NameSpace, error) {
	ms.RLock()
	defer ms.RUnlock()

	ns, ok := ms.prefix2base[prefix]
	if !ok {
		return nil, domain.ErrNameSpaceNotFound
	}

	return ns, nil
}

// GetWithBase returns a NameSpace from the store if the base URI is found.
func (ms *NameSpaceStore) GetWithBase(base string) (*domain.NameSpace, error) {
	ms.RLock()
	defer ms.RUnlock()

	ns, ok := ms.base2prefix[base]
	if !ok {
		return nil, domain.ErrNameSpaceNotFound
	}

	return ns, nil
}

// List returns a list of all the stored NameSpace objects.
// An error is only returned when the underlying datastructure is unavailable.
func (ms *NameSpaceStore) List() ([]*domain.NameSpace, error) {
	namespaces := []*domain.NameSpace{}
	for _, ns := range ms.namespaces {
		if ns != nil {
			namespaces = append(namespaces, ns)
		}
	}

	return namespaces, nil
}
