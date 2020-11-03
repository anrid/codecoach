package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User ...
type User struct {
	AccountID      ID          `json:"account_id" db:"account_id"`
	ID             ID          `json:"id" db:"id"`
	Email          string      `json:"email" db:"email"`
	PasswordHash   []byte      `json:"-" db:"password_hash"`
	Token          string      `json:"-" db:"token"`
	TokenExpiresAt *time.Time  `json:"-" db:"token_expires_at"`
	Profile        UserProfile `json:"profile" db:"profile"`
	Role           Role        `json:"role" db:"role"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time  `json:"updated_at" db:"updated_at"`
}

// NewUser ...
func NewUser(accountID ID, firstName, lastName, email, password string, role Role) *User {
	u := &User{
		AccountID: accountID,
		ID:        NewID(),
		Email:     email,
		Profile: UserProfile{
			FirstName: firstName,
			LastName:  lastName,
		},
		Role:      role,
		CreatedAt: time.Now(),
	}
	u.SetPassword(password)

	return u
}

// SetPassword ...
func (u *User) SetPassword(plaintextPassword string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	u.PasswordHash = hash
}

// CheckPassword ...
func (u *User) CheckPassword(plaintextPassword string) error {
	return bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(plaintextPassword))
}

// UserProfile ...
type UserProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Photo     string `json:"photo"`
}

// UserDAO ...
type UserDAO interface {
	Create(u *User) error
	Get(accountID, id ID) (*User, error)
	GetByEmail(accountID ID, email string) (*User, error)
	GetByToken(token string) (*User, error)
	Update(accountID, id ID, updates []Field) (*User, error)
}
