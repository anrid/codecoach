package user

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/anrid/codecoach/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// DAO ...
type DAO struct {
	db *sqlx.DB
}

var _ domain.UserDAO = &DAO{}

// New ...
func New(db *sqlx.DB) *DAO {
	return &DAO{db}
}

// Create ...
func (d *DAO) Create(ctx context.Context, u *domain.User) error {
	stmt, err := d.db.PrepareNamed(`
	INSERT INTO users
		(account_id, id, github_id, email, password_hash, token, token_expires_at, profile, role, created_at)
	VALUES
		(:account_id, :id, :github_id, :email, :password_hash, :token, :token_expires_at, :profile, :role, :created_at)
	RETURNING *
	`)
	if err != nil {
		return errors.Wrapf(err, "could not prepare statement")
	}

	err = stmt.Get(u, u)
	if err != nil {
		return errors.Wrapf(err, "could not create user")
	}

	return nil
}

// Get ...
func (d *DAO) Get(ctx context.Context, accountID, id domain.ID) (*domain.User, error) {
	u := new(domain.User)

	err := d.db.Get(u, "SELECT * FROM users WHERE account_id = $1 AND id = $2", accountID, id)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get user %s in account %s", id, accountID)
	}

	return u, nil
}

// GetByEmail ...
func (d *DAO) GetByEmail(ctx context.Context, accountID domain.ID, email string) (*domain.User, error) {
	u := new(domain.User)

	err := d.db.Get(u, "SELECT * FROM users WHERE account_id = $1 AND email = $2", accountID, email)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get user by email %s in account %s", email, accountID)
	}

	return u, nil
}

// GetByGithubID ...
func (d *DAO) GetByGithubID(ctx context.Context, accountID domain.ID, githubID int64) (*domain.User, error) {
	u := new(domain.User)

	err := d.db.Get(u, "SELECT * FROM users WHERE account_id = $1 AND github_id = $2", accountID, githubID)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get user by github id %d in account %s", githubID, accountID)
	}

	return u, nil
}

// GetAllByGithubID ...
func (d *DAO) GetAllByGithubID(ctx context.Context, githubID int64) ([]*domain.User, error) {
	var us []*domain.User

	err := d.db.Select(&us, "SELECT * FROM users WHERE github_id = $1", githubID)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get users with github id %d", githubID)
	}

	return us, nil
}

// GetAll ...
func (d *DAO) GetAll(ctx context.Context, accountID, id domain.ID) ([]*domain.User, int, error) {
	fields := "account_id, id, github_id, email, profile, role"

	sqlC := "SELECT COUNT(*) FROM users WHERE account_id = $1 AND id = $2"
	sqlN := "SELECT " + fields + " FROM users WHERE account_id = $1 AND id = $2"
	args := []interface{}{accountID, id}

	if id == "" {
		sqlC = "SELECT COUNT(*) FROM users WHERE account_id = $1"
		sqlN = "SELECT " + fields + " FROM users WHERE account_id = $1"
		args = []interface{}{accountID}
	}

	var total int
	err := d.db.Get(&total, sqlC, args...)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "could not list users (count)")
	}

	var us []*domain.User
	err = d.db.Select(&us, sqlN, args...)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "could not list users")
	}

	return us, total, nil
}

// GetByToken ...
func (d *DAO) GetByToken(ctx context.Context, token string) (*domain.User, error) {
	u := new(domain.User)

	err := d.db.Get(u, "SELECT * FROM users WHERE token = $1", token)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get user by token %s", token)
	}

	return u, nil
}

// Update ...
func (d *DAO) Update(ctx context.Context, accountID, id domain.ID, updates []domain.Field) (*domain.User, error) {
	q := psql.Update("users").Where("account_id = ? AND id = ?", accountID, id)

	for _, u := range updates {
		q = q.Set(u.Name, u.Value)
	}
	q = q.Set("updated_at", time.Now())

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "could not create update query")
	}

	_, err = d.db.Exec(sql, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not update user %s in account %s", id, accountID)
	}

	return d.Get(ctx, accountID, id)
}

// CreateTable ...
func (d *DAO) CreateTable() int {
	d.db.MustExec(`
	CREATE TABLE users (
		account_id CHAR(20) NOT NULL,
		id CHAR(20) NOT NULL,
		github_id BIGINT,
		email VARCHAR(255),
		password_hash VARCHAR(60),
		token CHAR(60) NOT NULL,
		token_expires_at TIMESTAMPTZ NULL,
		profile JSONB,
		role VARCHAR(32),
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NULL,
		PRIMARY KEY (account_id, id)
	)`)

	d.db.MustExec(`CREATE UNIQUE INDEX ON users (account_id, email)`)
	d.db.MustExec(`CREATE INDEX ON users (token)`)
	d.db.MustExec(`CREATE INDEX ON users (github_id)`)

	return 1
}
