package common

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONField is a custom type for storing JSON data in SQLite
type JSONField[T any] struct {
	Data *T
}

// Scan implements the sql.Scanner interface
func (j *JSONField[T]) Scan(value interface{}) error {
	if value == nil {
		j.Data = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		bytes = []byte(value.(string))
	}
	return json.Unmarshal(bytes, &j.Data)
}

// Value implements the driver.Valuer interface
func (j JSONField[T]) Value() (driver.Value, error) {
	if j.Data == nil {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

func (j JSONField[T]) GormDBDataType() string {
	return "text"
}

// MarshalJSON implements json.Marshaler
func (j JSONField[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONField[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.Data)
}
