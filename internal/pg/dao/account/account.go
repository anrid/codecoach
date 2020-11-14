package account

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

var _ domain.AccountDAO = &DAO{}

// New ...
func New(db *sqlx.DB) *DAO {
	return &DAO{db}
}

// Create ...
func (d *DAO) Create(ctx context.Context, a *domain.Account) error {
	stmt, err := d.db.PrepareNamed(`
	INSERT INTO accounts
		(id, name, code, profile, members, owner_id, created_at)
	VALUES
		(:id, :name, :code, :profile, :members, :owner_id, :created_at)
	RETURNING *
	`)
	if err != nil {
		return errors.Wrapf(err, "could not prepare statement")
	}

	err = stmt.Get(a, a)
	if err != nil {
		return errors.Wrapf(err, "could not create account")
	}

	return nil
}

// Get ...
func (d *DAO) Get(ctx context.Context, id domain.ID) (*domain.Account, error) {
	a := new(domain.Account)

	err := d.db.Get(a, "SELECT * FROM accounts WHERE id = $1", id)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get account %s", id)
	}

	return a, nil
}

// GetAll ...
func (d *DAO) GetAll(ctx context.Context, ids []domain.ID) ([]*domain.AccountInfo, error) {
	var as []*domain.AccountInfo

	err := d.db.Select(&as, "SELECT id, name, code, profile, created_at  FROM accounts WHERE id IN $1", ids)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get accounts with ids %v", ids)
	}

	return as, nil
}

// GetByCode ...
func (d *DAO) GetByCode(ctx context.Context, code string) (*domain.Account, error) {
	u := new(domain.Account)

	err := d.db.Get(u, "SELECT * FROM accounts WHERE code = $1", code)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get account by code %s", code)
	}

	return u, nil
}

// Update ...
func (d *DAO) Update(ctx context.Context, id domain.ID, updates []domain.Field) (*domain.Account, error) {
	q := psql.Update("accounts").Where("id = ?", id)

	for _, u := range updates {
		q = q.Set(u.Name, u.Value)
	}
	q = q.Set("updated_at", time.Now())

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "could not create update query")
	}

	println(sql)

	_, err = d.db.Exec(sql, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not update account %s", id)
	}

	return d.Get(ctx, id)
}

// CreateTable ...
func (d *DAO) CreateTable() int {
	d.db.MustExec(`
	CREATE TABLE accounts (
		id CHAR(20) NOT NULL,
		name VARCHAR(128),
		code VARCHAR(64),
		profile JSONB,
		members JSONB,
		owner_id CHAR(20) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NULL,
		PRIMARY KEY (id)
	)`)

	d.db.MustExec(`CREATE UNIQUE INDEX ON accounts (code)`)

	return 1
}
