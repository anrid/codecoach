package domain

import (
	"encoding/json"
	"fmt"

	"github.com/rs/xid"
)

// ID is a primary key.
type ID string

// NewID creates a new primary key.
func NewID() ID {
	return ID(xid.New().String())
}

// Field ...
type Field struct {
	Name  string
	Value interface{}
}

// Dump ...
func Dump(o interface{}) {
	b, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(b))
}
