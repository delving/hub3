package namespace

import (
	"sync"
)

// Store provides functionality to query and persist namespaces.
type Store interface {

	// Set persists the NameSpace object.
	//
	// When the object already exists it is overwritten.
	Set(ns *NameSpace) error

	// Delete removes the NameSpace from the store.
	//
	// Delete matches by the Prefix of the Namespace.
	Delete(ns *NameSpace) error

	// Len returns the number of stored namespaces
	Len() int

	// GetWithPrefix returns the NameSpace for a given prefix.
	// When the prefix is not found, an ErrNameSpaceNotFound error is returned.
	GetWithPrefix(prefix string) (ns *NameSpace, err error)

	// GetWithBase returns the NameSpace for a given base-URI.
	// When the base-URI is not found, an ErrNameSpaceNotFound error is returned.
	GetWithBase(base string) (ns *NameSpace, err error)
}

// memoryStore is the default namespace.Store for namespace.Service.
//
// Note: mutations in this store are ephemeral.
type memoryStore struct {
	sync.RWMutex
	prefix2base map[string]*NameSpace
	base2prefix map[string]*NameSpace
}

// newMemoryStore creates an in-memory namespace.Store.
func newMemoryStore() Store {
	return &memoryStore{
		prefix2base: make(map[string]*NameSpace),
		base2prefix: make(map[string]*NameSpace),
	}
}

// Len returns the number of stored namespaces
func (ms *memoryStore) Len() int {
	return len(ms.prefix2base)
}

// Set stores the NameSpace in the Store
func (ms *memoryStore) Set(ns *NameSpace) error {
	ms.Lock()
	defer ms.Unlock()
	ms.prefix2base[ns.Prefix] = ns
	ms.base2prefix[string(ns.Base)] = ns
	return nil
}

// Delete removes a NameSpace from the store
func (ms *memoryStore) Delete(ns *NameSpace) error {
	ms.Lock()
	defer ms.Unlock()
	_, ok := ms.prefix2base[ns.Prefix]
	if ok {
		delete(ms.prefix2base, ns.Prefix)
	}

	_, ok = ms.base2prefix[ns.Base.String()]
	if ok {
		delete(ms.base2prefix, ns.Base.String())
	}
	return nil
}

// GetWithPrefix returns a NameSpace from the store if the prefix is found.
func (ms *memoryStore) GetWithPrefix(prefix string) (*NameSpace, error) {
	ms.RLock()
	defer ms.RUnlock()
	ns, ok := ms.prefix2base[prefix]
	if !ok {
		return nil, ErrNameSpaceNotFound
	}
	return ns, nil
}

// GetWithBase returns a NameSpace from the store if the base URI is found.
func (ms *memoryStore) GetWithBase(base string) (*NameSpace, error) {
	ms.RLock()
	defer ms.RUnlock()
	ns, ok := ms.base2prefix[base]
	if !ok {
		return nil, ErrNameSpaceNotFound
	}
	return ns, nil
}
