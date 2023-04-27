package bulk

import (
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/index"
)

type Option func(*Service) error

func SetDBPath(path string) Option {
	return func(s *Service) error {
		if path == "" {
			return nil
		}

		s.dbPath = path
		return nil
	}
}

type BlobConfig struct {
	Endpoint        string `json:"endpoint,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	UseSSL          bool   `json:"useSSL,omitempty"`
	BucketName      string `json:"bucketName,omitempty"`
}

func SetBlobConfig(cfg BlobConfig) Option {
	return func(s *Service) error {
		s.blobCfg = cfg
		return nil
	}
}

func SetIndexService(is *index.Service) Option {
	return func(s *Service) error {
		s.index = is
		return nil
	}
}

func SetIndexTypes(indexTypes ...string) Option {
	return func(s *Service) error {
		s.indexTypes = indexTypes
		return nil
	}
}

func SetLogRequests(enable bool) Option {
	return func(s *Service) error {
		s.logRequests = enable
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
