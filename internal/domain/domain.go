package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

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

// Session ...
type Session struct {
	RequestID string
	User      *User
}

type contextKey string

const (
	contextKeySession contextKey = "session"
)

var requestIDCounter uint64

// ContextWithSession ...
func ContextWithSession(parent context.Context, u *User) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	id := atomic.AddUint64(&requestIDCounter, 1)
	requestID := fmt.Sprintf("%s-%d", time.Now().Format("2006-01-02"), id)

	s := &Session{
		RequestID: requestID,
		User:      u,
	}

	return context.WithValue(parent, contextKeySession, s)
}

// RequireSession ...
func RequireSession(ctx context.Context) (*Session, error) {
	v := ctx.Value(contextKeySession)
	if v == nil {
		return nil, errors.New("could not find session in context")
	}
	s, ok := v.(*Session)
	if !ok {
		return nil, errors.New("session is of invalid type")
	}
	return s, nil
}

var (
	notKebabCase = regexp.MustCompile(`[^a-z0-9\-]+`)
)

// CreateCode creates a kebab-case string.
func CreateCode(s string) string {
	code := notKebabCase.ReplaceAllString(strings.ToLower(s), "-")
	return strings.Trim(code, "- ")
}
