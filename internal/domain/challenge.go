package domain

import (
	"time"

	"github.com/pkg/errors"
)

// Challenge ...
type Challenge struct {
	AccountID       ID               `json:"account_id" db:"account_id"`
	ID              ID               `json:"id" db:"id"`
	ForUserID       ID               `json:"for_user_id" db:"for_user_id"`
	Status          Status           `json:"status" db:"status"`
	Details         ChallengeDetails `json:"details" db:"details"`
	CreatedByUserID ID               `json:"created_by_user_id" db:"created_by_user_id"`
	ExpiresAt       *time.Time       `json:"expires_at" db:"expires_at"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time       `json:"updated_at" db:"updated_at"`
}

type NewChallengeArgs struct {
	AccountID       ID
	ForUserID       ID
	CreatedByUserID ID
	Tasks           []*Task
}

// NewChallenge ...
func NewChallenge(a NewChallengeArgs) (*Challenge, error) {
	if a.AccountID == "" {
		return nil, errors.Errorf("missing account_id arg")
	}
	if a.ForUserID == "" {
		return nil, errors.Errorf("missing for_user_id arg")
	}
	if a.CreatedByUserID == "" {
		return nil, errors.Errorf("missing created_by_user_id arg")
	}
	if len(a.Tasks) == 0 {
		return nil, errors.Errorf("missing tasks arg")
	}

	c := &Challenge{
		AccountID:       a.AccountID,
		ID:              NewID(),
		ForUserID:       a.ForUserID,
		CreatedByUserID: a.CreatedByUserID,
		Details: ChallengeDetails{
			Tasks: a.Tasks,
		},
		CreatedAt: time.Now(),
	}

	return c, nil
}

// ChallengeDetails ...
type ChallengeDetails struct {
	Tasks []*Task `json:"tasks"`
}

// Status ...
type Status string

const (
	// StatusNew ...
	StatusNew Status = "new"
	// StatusScheduled ...
	StatusScheduled Status = "scheduled"
	// StatusActive ...
	StatusActive Status = "active"
	// StatusCompleted ...
	StatusCompleted Status = "completed"
	// StatusCanceled ...
	StatusCanceled Status = "canceled"
	// StatusExpired ...
	StatusExpired Status = "expired"
)

// ChallengeDAO ...
type ChallengeDAO interface {
	Create(u *Challenge) error
	Get(accountID, id ID) (*Challenge, error)
	GetAll(accountID ID, ids []ID) ([]*Challenge, error)
	Update(accountID ID, id ID, updates []Field) (*Challenge, error)
}
