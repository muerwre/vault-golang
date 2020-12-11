package utils

import (
	"database/sql/driver"
	"strings"
)

type EnumStringArray []string

// Contains determines if element is in a string array
func (f EnumStringArray) Contains(t string) bool {
	for _, a := range f {
		if a == t {
			return true
		}
	}

	return false
}

// Value used to pass this array as sql filter
func (f EnumStringArray) Value() (driver.Value, error) {
	val := strings.Join(f, ",")
	return val, nil
}
