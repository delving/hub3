package pid

import "gorm.io/gorm"

var _ Store = (*gormStore)(nil)

type gormStore struct {
	orm *gorm.DB
}

func NewGormStore(orm *gorm.DB) (*gormStore, error) {
	if migrateErr := orm.AutoMigrate(&PID{}); migrateErr != nil {
		return nil, migrateErr
	}

	return &gormStore{orm: orm}, nil
}

// Get implements Store.
func (g *gormStore) Get(id string, subject string) (resp StoreResponse, err error) {
	var pids []*PID
	err = g.orm.Where("external_id = ? OR id = ?", id, subject).Find(&pids).Error
	if err != nil {
		return StoreResponse{}, err
	}
	return NewStoreResponse(pids), nil
}

// Save implements Store.
func (g *gormStore) Save(pids ...*PID) (err error) {
	for _, pid := range pids {
		if err := g.orm.Save(pid).Error; err != nil {
			return err
		}
	}
	return nil
}
