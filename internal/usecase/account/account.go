package account

import (
	"context"

	"github.com/anrid/codecoach/internal/domain"
	"github.com/pkg/errors"
)

// UseCase ...
type UseCase struct {
	a domain.AccountDAO
}

var _ domain.AccountUseCases = &UseCase{}

// New ...
func New(a domain.AccountDAO) *UseCase {
	return &UseCase{a}
}

// Update ...
func (uc *UseCase) Update(ctx context.Context, id domain.ID, a domain.UpdateAccountArgs) (*domain.Account, error) {
	se, err := domain.RequireSession(ctx)
	if err != nil {
		return nil, err
	}

	// Only admins can update the account.
	if se.User.Role != domain.RoleAdmin {
		return nil, errors.Errorf("current user %s.%s (role: %s) cannot update account %s", se.User.AccountID, se.User.ID, se.User.Role, id)
	}

	old, err := uc.a.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "could not find account %s.%s", se.User.AccountID, id)
	}

	var updates []domain.Field

	if a.Name != "" {
		updates = append(updates, domain.Field{Name: "name", Value: a.Name})
	}
	if a.Logo != "" {
		old.Profile.Logo = a.Logo
		updates = append(updates, domain.Field{Name: "profile", Value: old.Profile})
	}

	up, err := uc.a.Update(ctx, id, updates)
	if err != nil {
		return nil, errors.Wrapf(err, "could not update account %s", id)
	}

	return up, nil
}
