package domain

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// Job ...
type Job struct {
	ID            ID         `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	Type          JobType    `json:"type" db:"type"`
	Status        Status     `json:"status" db:"status"`
	Details       JobDetails `json:"details" db:"details"`
	ScheduledDate *time.Time `json:"scheduled_at" db:"scheduled_at"`
	RunAt         time.Time  `json:"run_at" db:"run_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	FailedAt      *time.Time `json:"failed_at" db:"failed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at"`
}

// NewJob ...
func NewJob(name string, t JobType, d JobDetails) (*Job, error) {
	if name == "" {
		return nil, errors.Errorf("missing name arg")
	}
	if t == "" {
		return nil, errors.Errorf("missing name job type")
	}
	if d.Something == "" {
		return nil, errors.Errorf("missing name job details")
	}

	j := &Job{
		ID:        NewID(),
		Name:      name,
		Type:      t,
		Status:    StatusNew,
		Details:   d,
		CreatedAt: time.Now(),
	}

	return j, nil
}

// JobDetails ...
type JobDetails struct {
	Something string `json:"something"`
}

// JobType ...
type JobType string

const (
	// JobTypeGithub ...
	JobTypeGithub JobType = "github"
)

// JobDAO ...
type JobDAO interface {
	Create(ctx context.Context, u *Job) error
	Get(ctx context.Context, id ID) (*Job, error)
	GetAll(ctx context.Context, ids []ID) ([]*Job, error)
	Update(ctx context.Context, id ID, updates []Field) (*Job, error)
}

// JobUseCases ...
type JobUseCases interface {
	Enqueue(ctx context.Context, id ID, a EnqueueJobArgs) (*Job, error)
	Update(ctx context.Context, id ID, a UpdateJobArgs) (*Job, error)
}

// EnqueueJobArgs ...
type EnqueueJobArgs struct {
	Something string
}

// UpdateJobArgs ...
type UpdateJobArgs struct {
	Name string
	Logo string
}
