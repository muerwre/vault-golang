package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Model struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

type SimpleJson struct{}
type StringArray []string

func (s *SimpleJson) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &s)
}

func (s SimpleJson) Value() (driver.Value, error) {
	val, err := json.Marshal(s)
	return string(val), err
}

func (s *StringArray) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &s)
}

func (s StringArray) Value() (driver.Value, error) {
	val, err := json.Marshal(s)
	return string(val), err
}
