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

import (
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

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

func SetOrphanWait(wait int) Option {
	return func(s *Service) error {
		s.orphanWait = wait
		return nil
	}
}

func SetDisableMetrics(disable bool) Option {
	return func(s *Service) error {
		s.disableMetrics = disable
		return nil
	}
}

func SetOrganisationService(org *organization.Service) Option {
	return func(s *Service) error {
		s.org = org
		return nil
	}
}

func SetPostHookService(hooks ...domain.PostHookService) Option {
	return func(s *Service) error {
		for _, hook := range hooks {
			s.postHooks[hook.OrgID()] = append(s.postHooks[hook.OrgID()], hook)
		}

		return nil
	}
}
