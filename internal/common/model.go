package common

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var _ driver.Valuer = &JSONField[any]{}
var _ sql.Scanner = &JSONField[any]{}
var _ schema.GormDataTypeInterface = &JSONField[any]{}
var _ migrator.GormDataTypeInterface = &JSONField[any]{}

// JSONField is a custom type for storing JSON data in SQLite
type JSONField[T any] struct {
	data *T
}

func NewJSONField[T any](data T) *JSONField[T] {
	return &JSONField[T]{data: &data}
}

func (j *JSONField[T]) GetValue() *T {
	return j.data
}

// Scan implements the sql.Scanner interface
func (j *JSONField[T]) Scan(value any) error {
	if value == nil {
		j.data = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		bytes = []byte(value.(string))
	}
	return json.Unmarshal(bytes, &j.data)
}

// Value implements the driver.Valuer interface
func (j JSONField[T]) Value() (driver.Value, error) {
	if j.data == nil {
		return nil, nil
	}
	return json.Marshal(j.data)
}

func (j JSONField[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "text"
}

func (j JSONField[T]) GormDataType() string {
	return "json"
}

// MarshalJSON implements json.Marshaler
func (j *JSONField[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.data)
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONField[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.data)
}
