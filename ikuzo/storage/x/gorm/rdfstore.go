package gorm

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/delving/hub3/ikuzo/rdf"
)

func main() {
	dsn := "host=localhost user=postgres password=pw123 dbname=rdf port=5433 sslmode=disable TimeZone=Europe/Amsterdam"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to connect to the database: %s", err)
		return
	}

	db.AutoMigrate(
		&NamedGraph{},
		&Triple{},
		&Resource{},
		&Predicate{},
		&Namespace{},
		&Literal{},
		&DataType{},
		&Language{},
	)
}

type RDFStore struct {
	db *gorm.DB
}

func NewRDFStore(db *gorm.DB) (*RDFStore, error) {
	return &RDFStore{db: db}, nil
}

func (s *RDFStore) Save(g *rdf.Graph) error {
	return nil
}
