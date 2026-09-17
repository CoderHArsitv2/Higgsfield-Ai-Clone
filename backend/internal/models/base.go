// Package models holds one file per table. Each file carries the struct and the
// queries for that table together, so everything that can touch a table lives
// in one place.
//
// Callers depend on the interfaces in store.go rather than on these types'
// methods, which keeps SQL out of the services and controllers entirely.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *Base) BeforeCreate(*gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// JSON is a jsonb column.
//
// Defined here rather than imported from gorm.io/datatypes, which pulls the
// MySQL and MongoDB drivers into the binary for the sake of one column type.
type JSON json.RawMessage

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return string(j), nil
}

func (j *JSON) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*j = JSON("{}")
	case []byte:
		*j = JSON(append([]byte(nil), v...))
	case string:
		*j = JSON(v)
	default:
		return errors.New("models.JSON: unsupported scan type")
	}
	return nil
}

func (JSON) GormDataType() string { return "jsonb" }

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return j, nil
}

func (j *JSON) UnmarshalJSON(b []byte) error {
	*j = JSON(append([]byte(nil), b...))
	return nil
}

// tables is the migration order, and the order AutoMigrate would need too.
func tables() []any {
	return []any{&User{}, &UserAPIKey{}, &Generation{}, &Asset{}}
}
