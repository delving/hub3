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
		s.CreateTreeFn = fn
		return nil
	}
}
