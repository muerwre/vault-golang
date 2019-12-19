package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
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
type CommaUintArray []uint

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

func (s *CommaUintArray) Scan(src interface{}) error {
	strings := strings.Split(string(src.([]byte)), ",")
	var numbers []uint

	for _, k := range strings {
		val, err := strconv.ParseUint(k, 10, 32)

		if err == nil {
			numbers = append(numbers, uint(val))
		}
	}

	*s = numbers
	return nil
}

func (s CommaUintArray) Value() (driver.Value, error) {
	return strings.Trim(strings.Replace(fmt.Sprint(s), " ", ",", -1), "[]"), nil
}

func (s CommaUintArray) Contains(e uint) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (s *CommaStringArray) Scan(src interface{}) error {
	*s = strings.Split(string(src.([]byte)), ",")
	return nil
}

func (s CommaStringArray) Value() (driver.Value, error) {
	return strings.Join(s, ","), nil
}
