package pid

import "gorm.io/gorm"

type Option func(*Service) error

func SetOrm(orm *gorm.DB) Option {
	return func(s *Service) error {
		s.orm = orm
		return nil
	}
}
