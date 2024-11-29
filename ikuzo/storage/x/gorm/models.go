package gorm

import (
	"time"

	"github.com/google/uuid"
)

/*
* The goal of this format is to only store each set of information once. Via the deterministic way of convering the Triple
* components into their UUID parts we can determine if a triple is already stored without having to do lookups.
*
* For retrieval it can query retrieve all triples for a name graph or check if triple or resource exists
*
* Only the Namegraph is not immutable. Everytime this is modified a trigger is run that creates a copy of the previous version
* together with its start and end date. This allows for historical query to see what whas the view in any point in time.
 */

type Namespace struct {
	Prefix     string
	URI        string
	RDFSSchema string
	ID         uuid.UUID `gorm:"primaryKey"`
}

type Language struct {
	Value string    `gorm:"unique,not null"`
	ID    uuid.UUID `gorm:"primaryKey"`
}

type DataType struct {
	Value string    `gorm:"unique,not null"`
	ID    uuid.UUID `gorm:"primaryKey"`
}

type Literal struct {
	Value    string    `gorm:"unique,not null"`
	Language Language  `gorm:"foreignKey:ID"`
	DataType DataType  `gorm:"foreignKey:ID"`
	ID       uuid.UUID `gorm:"primaryKey"`
}

type Predicate struct {
	Value     string    `gorm:"unique,not null"`
	Namespace Namespace `gorm:"foreignKey:ID"`
	ID        uuid.UUID `gorm:"primaryKey"`
}

type Resource struct {
	Value string    `gorm:"unique,not null"`
	ID    uuid.UUID `gorm:"primaryKey"`
}

type Triple struct {
	Resource      Resource  `gorm:"foreignKey:ID,not null"`
	ObjectPointer Resource  `gorm:"foreignKey:ID"`
	Predicate     Predicate `gorm:"foreignKey:ID,not null"`
	ObjectLiteral Literal   `gorm:"foreignKey:ID"`
	ID            uuid.UUID `gorm:"primaryKey"`
}

type NamedGraph struct {
	ModifiedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;index"`
	Meta       Meta      `gorm:"type:jsonb"`
	URI        string    `gorm:"unique"`
	Triples    []Triple  `gorm:"many2many:graph_triples"`
	ID         uuid.UUID `gorm:"primaryKey"`
}

type Meta struct {
	Dataset string
	OrgID   string
}
