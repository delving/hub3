package ead

import (
	"github.com/delving/hub3/ikuzo/service/x/index"
)

type Option func(*Service) error

func SetDataDir(path string) Option {
	return func(s *Service) error {
		s.dataDir = path
		return nil
	}
}

func SetIndexService(is *index.Service) Option {
	return func(s *Service) error {
		s.index = is
		return nil
	}
}

func SetCreateTree(fn CreateTreeFn) Option {
	return func(s *Service) error {
		s.createTree = fn
		return nil
	}
}
