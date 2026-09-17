package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

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
