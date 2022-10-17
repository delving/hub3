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
	"fmt"
	"sync"

	"github.com/delving/hub3/ikuzo/domain"
)

// NamespaceStore is the default namespace.Store for namespace.Service.
//
// Note: mutations in this store are ephemeral.
type NamespaceStore struct {
	sync.RWMutex
	prefix2base map[string]*domain.Namespace
	base2prefix map[string]*domain.Namespace
	namespaces  map[string]*domain.Namespace
}

// NewNamespaceStore creates an in-memory namespace.Store.
func NewNamespaceStore() *NamespaceStore {
	return &NamespaceStore{
		prefix2base: make(map[string]*domain.Namespace),
		base2prefix: make(map[string]*domain.Namespace),
		namespaces:  make(map[string]*domain.Namespace),
	}
}

// Len returns the number of stored namespaces.
// Alternatives Base or Prefixes don't count towards the total.
func (ms *NamespaceStore) Len() int {
	return len(ms.namespaces)
}

// Put stores the NameSpace in the Store
func (ms *NamespaceStore) Put(ns *domain.Namespace) error {
	if ns == nil {
		return fmt.Errorf("cannot store empty namespace")
	}

	// this implementation of Delete can never return an error
	_ = ms.Delete(ns.GetID())

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

func (ms *NamespaceStore) delete(ns *domain.Namespace) error {
	return ms.Delete(ns.GetID())
}

// Delete removes a NameSpace from the store.
//
// When the Namespace is not found it returns an domain.ErrNameSpaceNotFound error.
func (ms *NamespaceStore) Delete(id string) error {
	ms.Lock()
	defer ms.Unlock()

	ns, ok := ms.namespaces[id]
	if !ok {
		return domain.ErrNamespaceNotFound
	}

	delete(ms.namespaces, id)

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

func (ms *NamespaceStore) Get(id string) (*domain.Namespace, error) {
	ns, ok := ms.namespaces[id]
	if !ok {
		return nil, domain.ErrNamespaceNotFound
	}

	return ns, nil
}

// GetWithPrefix returns a NameSpace from the store if the prefix is found.
func (ms *NamespaceStore) GetWithPrefix(prefix string) (*domain.Namespace, error) {
	ms.RLock()
	defer ms.RUnlock()

	ns, ok := ms.prefix2base[prefix]
	if !ok {
		return nil, domain.ErrNamespaceNotFound
	}

	return ns, nil
}

// GetWithBase returns a NameSpace from the store if the base URI is found.
func (ms *NamespaceStore) GetWithBase(base string) (*domain.Namespace, error) {
	ms.RLock()
	defer ms.RUnlock()

	ns, ok := ms.base2prefix[base]
	if !ok {
		return nil, domain.ErrNamespaceNotFound
	}

	return ns, nil
}

// List returns a list of all the stored NameSpace objects.
// An error is only returned when the underlying datastructure is unavailable.
func (ms *NamespaceStore) List() ([]*domain.Namespace, error) {
	namespaces := []*domain.Namespace{}
	for _, ns := range ms.namespaces {
		if ns != nil {
			namespaces = append(namespaces, ns)
		}
	}

	return namespaces, nil
}
