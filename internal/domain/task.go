package domain

import (
	"time"

	"github.com/pkg/errors"
)

// Task ...
type Task struct {
	AccountID ID          `json:"account_id" db:"account_id"`
	ID        ID          `json:"id" db:"id"`
	Name      string      `json:"name" db:"name"`
	Type      TaskType    `json:"type" db:"type"`
	Details   TaskDetails `json:"details" db:"details"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time  `json:"updated_at" db:"updated_at"`
}

type NewTaskArgs struct {
	AccountID ID
	Name      string
	Type      TaskType
	Details   TaskDetails
}

// NewTask ...
func NewTask(a NewTaskArgs) (*Task, error) {
	if a.AccountID == "" {
		return nil, errors.Errorf("missing account_id arg")
	}
	if a.Name == "" {
		return nil, errors.Errorf("missing name arg")
	}
	if a.Type == "" {
		return nil, errors.Errorf("missing type arg")
	}

	c := &Task{
		AccountID: a.AccountID,
		ID:        NewID(),
		Name:      a.Name,
		Type:      a.Type,
		Details:   a.Details,
		CreatedAt: time.Now(),
	}

	return c, nil
}

// TaskDetails ...
type TaskDetails struct {
	GithubRepoName string        `json:"github_repo_name" db:"github_repo_name"`
	TimeLimit      time.Duration `json:"time_limit" db:"time_limit"`
}

// TaskType ...
type TaskType string

const (
	// TaskTypeRefactor ...
	TaskTypeRefactor TaskType = "refactor"
	// TaskTypeCodeReview ...
	TaskTypeCodeReview TaskType = "code_review"
	// TaskTypeCoding ...
	TaskTypeCoding TaskType = "coding"
)

// TaskDAO ...
type TaskDAO interface {
	Create(u *Task) error
	Get(accountID, id ID) (*Task, error)
	GetAll(accountID ID, ids []ID) ([]*Task, error)
	Update(accountID ID, id ID, updates []Field) (*Task, error)
}
