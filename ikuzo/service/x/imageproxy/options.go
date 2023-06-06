package imageproxy

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rs/zerolog"
)

type Option func(*Service) error

func SetCacheDir(path string) Option {
	return func(s *Service) error {
		s.cacheDir = path
		return nil
	}
}

func SetMaxSizeCacheDir(size int) Option {
	return func(s *Service) error {
		s.maxSizeCacheDir = size * 1024
		return nil
	}
}

func SetLruCacheSize(size int) Option {
	return func(s *Service) error {
		lruCache, err := lru.NewARC[string, []byte](size)
		if err != nil {
			return fmt.Errorf("unable to create lru: %w", err)
		}

		s.lruCache = lruCache

		return nil
	}
}

func SetTimeout(duration int) Option {
	return func(s *Service) error {
		s.timeOut = duration
		return nil
	}
}

func SetEnableResize(enabled bool) Option {
	return func(s *Service) error {
		s.enableResize = enabled
		return nil
	}
}

func SetProxyReferrer(referrer []string) Option {
	return func(s *Service) error {
		s.referrers = referrer
		return nil
	}
}

func SetRefuseList(refuseList []string) Option {
	return func(s *Service) error {
		s.refuselist = refuseList
		return nil
	}
}

func SetAllowList(allowList []string) Option {
	return func(s *Service) error {
		s.allowList = allowList
		return nil
	}
}

func SetAllowPorts(ports []string) Option {
	return func(s *Service) error {
		s.allowPorts = ports
		return nil
	}
}

func SetAllowedMimeTypes(allowedMimeTypes []string) Option {
	return func(s *Service) error {
		s.allowedMimeTypes = allowedMimeTypes
		return nil
	}
}

func SetProxyPrefix(prefix string) Option {
	return func(s *Service) error {
		s.proxyPrefix = prefix
		return nil
	}
}

func SetLogger(logger *zerolog.Logger) Option {
	return func(s *Service) error {
		s.log = logger.With().Str("svc", "imageproxy").Logger()
		return nil
	}
}

func SetDefaultImagePath(path string) Option {
	return func(s *Service) error {
		s.defaultImagePath = path
		return nil
	}
}
