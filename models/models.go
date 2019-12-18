package models

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
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
type CommaStringArray []string

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

func (s *CommaStringArray) Scan(src interface{}) error {
	*s = strings.Split(string(src.([]byte)), ",")
	return nil
}

func (s CommaStringArray) Value() (driver.Value, error) {
	return strings.Join(s, ","), nil
}
