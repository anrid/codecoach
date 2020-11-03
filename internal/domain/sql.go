package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Value ...
func (s AccountProfile) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan ...
func (s *AccountProfile) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

// Value ...
func (s Members) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan ...
func (s *Members) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

// Value ...
func (s UserProfile) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan ...
func (s *UserProfile) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}
