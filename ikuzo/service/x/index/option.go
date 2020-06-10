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

package index

import "github.com/elastic/go-elasticsearch/v8/esutil"

type Option func(*Service) error

func SetBulkIndexer(bi esutil.BulkIndexer, direct bool) Option {
	return func(s *Service) error {
		s.bi = bi
		s.direct = direct

		return nil
	}
}

func SetNatsConfiguration(ncfg *NatsConfig) Option {
	return func(s *Service) error {
		s.stan = ncfg
		s.stan.setDefaults()
		return nil
	}
}

func WithDefaultMessageHandle() Option {
	return func(s *Service) error {
		s.MsgHandler = s.submitBulkMsg
		return nil
	}
}
