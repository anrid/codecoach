package domain

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User ...
type User struct {
	AccountID      ID          `json:"account_id" db:"account_id"`
	ID             ID          `json:"id" db:"id"`
	GithubID       int64       `json:"github_id" db:"github_id"`
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
func NewUser(accountID ID, givenName, familyName, email, password string, role Role) *User {
	u := &User{
		AccountID: accountID,
		ID:        NewID(),
		Email:     email,
		Profile: UserProfile{
			GivenName:  givenName,
			FamilyName: familyName,
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
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
	PhotoURL    string `json:"photo_url"`
	GithubLogin string `json:"github_login"`
	Location    string `json:"location"`
	IsSuspended bool   `json:"is_suspended"`
}

// UserDAO ...
type UserDAO interface {
	Create(ctx context.Context, u *User) error
	Get(ctx context.Context, accountID, id ID) (*User, error)
	GetByEmail(ctx context.Context, accountID ID, email string) (*User, error)
	GetByGithubID(ctx context.Context, accountID ID, githubID int64) (*User, error)
	GetAllByGithubID(ctx context.Context, githubID int64) ([]*User, error)
	GetByToken(ctx context.Context, token string) (*User, error)
	Update(ctx context.Context, accountID, id ID, updates []Field) (*User, error)
}

// UserUseCases ...
type UserUseCases interface {
	Signup(ctx context.Context, a SignupArgs) (*SignupResult, error)
	Login(ctx context.Context, accountCode, email, password string) (*LoginResult, error)
	GithubLogin(ctx context.Context, accountCode string, githubID int64) (*LoginResult, error)
	GithubGetAvailableAccounts(ctx context.Context, githubID int64) ([]*Account, error)
	Create(ctx context.Context, a CreateUserArgs) (*User, error)
	Update(ctx context.Context, accountID, id ID, a UpdateArgs) (*User, error)
}

// SignupArgs ...
type SignupArgs struct {
	AccountName string
	GivenName   string
	FamilyName  string
	Email       string
	Password    string
	GithubID    int64
	GithubLogin string
	PhotoURL    string
	Location    string
}

// SignupResult ...
type SignupResult struct {
	Account *Account
	User    *User
	Token   string
}

// LoginResult ...
type LoginResult struct {
	Account *Account
	User    *User
	Token   string
}

// CreateUserArgs ...
type CreateUserArgs struct {
	GivenName  string
	FamilyName string
	Email      string
	Password   string
	Role       Role
}

// UpdateArgs ...
type UpdateArgs struct {
	GivenName  string
	FamilyName string
	Email      string
	Password   string
}
