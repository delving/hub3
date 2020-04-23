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

	// Add either the Base and Prefix alternatives depending on which one
	// is found first. When neither is found a new NameSpace is created.
	Add(prefix, base string) (*NameSpace, error)

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
	namespaces  map[string]*NameSpace
}

// newMemoryStore creates an in-memory namespace.Store.
func newMemoryStore() Store {
	return &memoryStore{
		prefix2base: make(map[string]*NameSpace),
		base2prefix: make(map[string]*NameSpace),
		namespaces:  make(map[string]*NameSpace),
	}
}

// Len returns the number of stored namespaces.
// Alternatives Base or Prefixes don't count towards the total.
func (ms *memoryStore) Len() int {
	return len(ms.namespaces)
}

// Set stores the NameSpace in the Store
func (ms *memoryStore) Set(ns *NameSpace) error {
	err := ms.Delete(ns)
	ms.Lock()
	defer ms.Unlock()
	if err != nil {
		return err
	}

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

// Add adds the prefix and base to a NameSpace.
// If nor prefix, base are previously stored a new NameSpace is created.
func (ms *memoryStore) Add(prefix, base string) (*NameSpace, error) {
	ns, err := ms.GetWithPrefix(prefix)
	if err != nil {
		if err != ErrNameSpaceNotFound {
			return nil, err
		}
	}
	if ns != nil {
		err = ns.AddBase(base)
		if err != nil {
			return nil, err
		}
		return ns, nil
	}

	ns, err = ms.GetWithBase(base)
	if err != nil {
		if err != ErrNameSpaceNotFound {
			return nil, err
		}
	}
	if ns != nil {
		err = ns.AddPrefix(prefix)
		if err != nil {
			return nil, err
		}
		return ns, nil

	}

	ns = &NameSpace{
		Prefix: prefix,
		Base:   base,
	}
	err = ms.Set(ns)
	if err != nil {
		return nil, err
	}

	return ns, nil
}

// Delete removes a NameSpace from the store
func (ms *memoryStore) Delete(ns *NameSpace) error {
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
