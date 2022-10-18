package namespace

import (
	"fmt"
	"sort"
	"sync"

	"github.com/delving/hub3/ikuzo/domain"
)

type ListOptions struct {
	URI    string
	Prefix string
	ID     string
	Limit  int
}

// Store provides functionality to query and persist namespaces.
type Store interface {
	// Put persists the NameSpace object.
	//
	// When the object already exists it is overwritten.
	Put(ns *domain.Namespace) error

	// Delete removes the NameSpace from the store.
	//
	// Delete matches by the Prefix of the Namespace.
	Delete(ID string) error

	// Len returns the number of stored namespaces
	Len() int

	// Get returns a namespace by its ID
	Get(id string) (ns *domain.Namespace, err error)

	// // GetWithPrefix returns the NameSpace for a given prefix.
	// // When the prefix is not found, an ErrNameSpaceNotFound error is returned.
	// GetWithPrefix(prefix string) (ns *domain.Namespace, err error)

	// // GetWithBase returns the NameSpace for a given base-URI.
	// // When the base-URI is not found, an ErrNameSpaceNotFound error is returned.
	// GetWithBase(base string) (ns *domain.Namespace, err error)

	// List returns a list of all the NameSpaces.
	// When an empty Filter{} is given all Namespaces are returned ordered by Prefix.
	// When the filter is not empty, NameSpaces are sorted by weight
	List(opts *ListOptions) ([]*domain.Namespace, error)
}

var _ Store = (*namespaceStore)(nil)

// namespaceStore is the default namespace.Store for namespace.Service.
//
// Note: mutations in this store are ephemeral.
type namespaceStore struct {
	sync.RWMutex

	// unique namespaces with prefix_uri as key
	namespaces map[string]*domain.Namespace

	// lookup indices
	prefix2base map[string][]*domain.Namespace
	base2prefix map[string][]*domain.Namespace
}

// NewNamespaceStore creates an in-memory namespace.Store.
func NewNamespaceStore() *namespaceStore {
	return &namespaceStore{
		prefix2base: make(map[string][]*domain.Namespace),
		base2prefix: make(map[string][]*domain.Namespace),
		namespaces:  make(map[string]*domain.Namespace),
	}
}

// Len returns the number of stored namespaces.
// Alternatives Base or Prefixes don't count towards the total.
func (ms *namespaceStore) Len() int {
	return len(ms.namespaces)
}

// Put stores the NameSpace in the Store
func (ms *namespaceStore) Put(ns *domain.Namespace) error {
	if ns == nil {
		return fmt.Errorf("cannot store empty namespace")
	}

	if ns.ID == "" {
		ns.ID = fmt.Sprintf("%s:%s", ns.Prefix, ns.URI)
	}

	stored, ok := ms.namespaces[ns.ID]
	if ok && stored.Weight == ns.Weight {
		// already in the store
		return nil
	}

	// if you lock earlier you get a deadlock
	ms.Lock()
	defer ms.Unlock()

	// add unique namespace to namespace Map
	ms.namespaces[ns.ID] = ns

	setWeight := ns.Weight == 0

	// add prefix to index
	prefixes, ok := ms.prefix2base[ns.Prefix]
	if !ok {
		prefixes = []*domain.Namespace{}
	}

	prefixes = append(prefixes, ns)
	if setWeight {
		ns.Weight = len(prefixes) + 1
	}
	sort.Slice(prefixes, func(i, j int) bool {
		return prefixes[i].Weight > prefixes[j].Weight
	})

	ms.prefix2base[ns.Prefix] = prefixes

	// add uris to index
	uris, ok := ms.base2prefix[ns.URI]
	if !ok {
		uris = []*domain.Namespace{}
	}

	uris = append(uris, ns)
	if setWeight {
		ns.Weight = len(uris) + 1
	}
	sort.Slice(uris, func(i, j int) bool {
		return uris[i].Weight > uris[j].Weight
	})

	ms.base2prefix[ns.URI] = uris

	return nil
}

func (ms *namespaceStore) delete(ns *domain.Namespace) error {
	return ms.Delete(ns.ID)
}

// Delete removes a NameSpace from the store.
//
// When the Namespace is not found it returns an domain.ErrNameSpaceNotFound error.
func (ms *namespaceStore) Delete(id string) error {
	ms.Lock()
	defer ms.Unlock()

	if id == "" {
		return nil
	}

	ns, ok := ms.namespaces[id]
	if !ok {
		return domain.ErrNamespaceNotFound
	}

	delete(ms.namespaces, id)

	// remove instance of Namespace from the Namespace slice
	prefixes, ok := ms.prefix2base[ns.Prefix]
	if ok {
		var filtered []*domain.Namespace

		for _, targetNS := range prefixes {
			if targetNS.ID != ns.ID {
				filtered = append(filtered, targetNS)
			}
		}

		ms.prefix2base[ns.Prefix] = filtered
		if len(filtered) == 0 {
			delete(ms.prefix2base, ns.Prefix)
		}
	}

	// remove instance of Namespace from the Namespace slice
	uris, ok := ms.base2prefix[ns.URI]
	if ok {
		var filtered []*domain.Namespace

		for _, targetURI := range uris {
			if targetURI.ID != ns.ID {
				filtered = append(filtered, targetURI)
			}
		}

		ms.base2prefix[ns.URI] = filtered
		if len(filtered) == 0 {
			delete(ms.base2prefix, ns.URI)
		}
	}

	return nil
}

func (ms *namespaceStore) Get(id string) (*domain.Namespace, error) {
	ns, ok := ms.namespaces[id]
	if !ok {
		return nil, domain.ErrNamespaceNotFound
	}

	return ns, nil
}

// GetWithPrefix returns a NameSpace from the store if the prefix is found.
func (ms *namespaceStore) GetWithPrefix(prefix string) (*domain.Namespace, error) {
	ms.RLock()
	defer ms.RUnlock()

	ns, ok := ms.prefix2base[prefix]
	if !ok || len(ns) == 0 {
		return nil, domain.ErrNamespaceNotFound
	}

	return ns[0], nil
}

// GetWithBase returns a NameSpace from the store if the base URI is found.
func (ms *namespaceStore) GetWithBase(base string) (*domain.Namespace, error) {
	ms.RLock()
	defer ms.RUnlock()

	ns, ok := ms.base2prefix[base]
	if !ok || len(ns) == 0 {
		return nil, domain.ErrNamespaceNotFound
	}

	return ns[0], nil
}

// List returns a list of all the stored NameSpace objects.
// An error is only returned when the underlying datastructure is unavailable.
func (ms *namespaceStore) List(opts *ListOptions) ([]*domain.Namespace, error) {
	namespaces := []*domain.Namespace{}

	if opts == nil {
		for _, ns := range ms.namespaces {
			if ns != nil {
				namespaces = append(namespaces, ns)
			}
		}

		return namespaces, nil
	}

	if opts.Prefix != "" && opts.URI != "" {
		opts.ID = fmt.Sprintf("%s:%s", opts.Prefix, opts.URI)
	}

	if opts.ID != "" {
		ns, err := ms.Get(opts.ID)
		if err != nil {
			return namespaces, err
		}

		namespaces = append(namespaces, ns)

		return namespaces, nil
	}

	if opts.Prefix != "" {
		ns, ok := ms.prefix2base[opts.Prefix]
		if !ok {
			return namespaces, nil
		}

		namespaces = append(namespaces, ns...)
	}

	if opts.URI != "" {
		ns, ok := ms.base2prefix[opts.URI]
		if !ok {
			return namespaces, nil
		}

		namespaces = append(namespaces, ns...)
	}

	return namespaces, nil
}
