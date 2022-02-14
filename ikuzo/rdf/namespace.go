package rdf

import "github.com/delving/hub3/ikuzo/domain"

type NamespaceManager interface {
	GetWithPrefix(prefix string) (ns *domain.Namespace, err error)
	GetWithBase(base string) (ns *domain.Namespace, err error)
}
