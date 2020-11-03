package domain

import (
	"time"
)

// Account ...
type Account struct {
	ID        ID             `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	Code      string         `json:"code" db:"code"`
	Profile   AccountProfile `json:"profile" db:"profile"`
	Members   Members        `json:"members" db:"members"`
	OwnerID   ID             `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at" db:"updated_at"`
}

// NewAccount ...
func NewAccount(name, code string) *Account {
	u := &Account{
		ID:        NewID(),
		Name:      name,
		Code:      code,
		CreatedAt: time.Now(),
	}

	return u
}

// AddMember adds a new member to account, or updates a member
// if the id already exists.
func (a *Account) AddMember(nm Member) {
	for i, m := range a.Members {
		if m.ID == nm.ID {
			a.Members[i] = nm
			return
		}
	}
	a.Members = append(a.Members, nm)
}

// AccountProfile ...
type AccountProfile struct {
	Logo string `json:"logo"`
}

// Role ...
type Role string

const (
	// AdminRole ...
	AdminRole Role = "admin"
	// HiringManagerRole ...
	HiringManagerRole Role = "hiring_manager"
	// CandidateRole ...
	CandidateRole Role = "candidate"
)

// Member ...
type Member struct {
	ID      ID        `json:"id"`
	Role    Role      `json:"role"`
	AddedAt time.Time `json:"added_at"`
}

// Members ...
type Members []Member

// AccountDAO ...
type AccountDAO interface {
	Create(u *Account) error
	Get(id ID) (*Account, error)
	GetByCode(code string) (*Account, error)
	Update(id ID, updates []Field) (*Account, error)
}
