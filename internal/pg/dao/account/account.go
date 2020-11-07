package account

import (
	"fmt"
	"strings"
	"time"

	"github.com/anrid/codecoach/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

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
func (d *DAO) Create(a *domain.Account) error {
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
func (d *DAO) Get(id domain.ID) (*domain.Account, error) {
	u := new(domain.Account)

	err := d.db.Get(u, "SELECT * FROM accounts WHERE id = $1", id)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get account %d", id)
	}

	return u, nil
}

// GetByCode ...
func (d *DAO) GetByCode(code string) (*domain.Account, error) {
	u := new(domain.Account)

	err := d.db.Get(u, "SELECT * FROM accounts WHERE code = $1", code)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get account by code %s", code)
	}

	return u, nil
}

// Update ...
func (d *DAO) Update(id domain.ID, updates []domain.Field) (*domain.Account, error) {
	var fields []string
	var values []interface{}

	updates = append(updates, domain.Field{Name: "updated_at", Value: time.Now()})

	for i, u := range updates {
		fields = append(fields, fmt.Sprintf("%s = $%d", u.Name, i+1))
		values = append(values, u.Value)
	}

	q := fmt.Sprintf(`UPDATE accounts SET %s WHERE id = %s`, strings.Join(fields, ", "), id)

	_, err := d.db.Exec(q, values...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not update account %d", id)
	}

	return d.Get(id)
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
