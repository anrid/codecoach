package user

import (
	"context"
	"time"

	"github.com/anrid/codecoach/internal/config"
	"github.com/anrid/codecoach/internal/domain"
	token_gen "github.com/anrid/codecoach/internal/pkg/token"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// UseCase ...
type UseCase struct {
	c *config.Config
	a domain.AccountDAO
	u domain.UserDAO
}

var _ domain.UserUseCase = &UseCase{}

// New ...
func New(c *config.Config, a domain.AccountDAO, u domain.UserDAO) *UseCase {
	return &UseCase{c, a, u}
}

// Signup ...
func (uc *UseCase) Signup(ctx context.Context, sa domain.SignupArgs) (*domain.SignupResult, error) {
	// Create new account.
	a, err := domain.NewAccount(sa.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "could not create account code")
	}

	// Create account admin.
	u := domain.NewUser(a.ID, sa.GivenName, sa.FamilyName, sa.Email, sa.Password, domain.AdminRole)

	// Add admin to account.
	a.AddMember(domain.Member{
		ID:      u.ID,
		Role:    domain.AdminRole,
		AddedAt: time.Now(),
	})
	a.OwnerID = u.ID

	// Save account.
	err = uc.a.Create(a)
	if err != nil {
		return nil, errors.Wrap(err, "could not create account")
	}

	// Generate access token.
	u.Token = token_gen.New()

	// Make sure token expires at some point in the future.
	expiresAt := time.Now().Add(uc.c.TokenExpires)
	u.TokenExpiresAt = &expiresAt

	err = uc.u.Create(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "could not create user")
	}

	// Log successful signup.
	zap.S().Infow(
		"signup successful",
		"account", a.ID,
		"user", u.ID,
		"email", u.Email,
		"token_expires", u.TokenExpiresAt,
	)

	return &domain.SignupResult{
		Account: a,
		User:    u,
		Token:   u.Token,
	}, nil
}

// Login ...
func (uc *UseCase) Login(ctx context.Context, accountCode, email, password string) (*domain.LoginResult, error) {
	// Get account by account code.
	accountCode = domain.CreateCode(accountCode)
	if len(accountCode) < 2 {
		return nil, errors.Errorf("invalid account code '%s'", accountCode)
	}

	// Get account by account code.
	a, err := uc.a.GetByCode(accountCode)
	if err != nil {
		return nil, errors.Wrap(err, "invalid account, email or password")
	}

	// Get user by email and account id.
	u, err := uc.u.GetByEmail(ctx, a.ID, email)
	if err != nil {
		return nil, errors.Wrap(err, "invalid account, email or password")
	}

	// Verify password.
	err = u.CheckPassword(password)
	if err != nil {
		return nil, errors.Wrap(err, "invalid account, email or password")
	}

	// Generate access token.
	token := token_gen.New()

	// Make sure token expires at some point in the future.
	tokenExpires := time.Now().Add(uc.c.TokenExpires)

	// Update user's token.
	_, err = uc.u.Update(ctx, u.AccountID, u.ID, []domain.Field{
		{Name: "token", Value: token},
		{Name: "token_expires_at", Value: tokenExpires},
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not update user's token")
	}

	// Log successful login.
	zap.S().Infow(
		"login successful",
		"account", u.AccountID,
		"user", u.ID,
		"email", u.Email,
		"token_expires", tokenExpires.String(),
	)

	return &domain.LoginResult{
		Account: a,
		User:    u,
		Token:   token,
	}, nil
}

// Create ...
func (uc *UseCase) Create(ctx context.Context, a domain.CreateUserArgs) (*domain.User, error) {
	se, err := domain.RequireSession(ctx)
	if err != nil {
		return nil, err
	}

	u := domain.NewUser(se.User.AccountID, a.GivenName, a.FamilyName, a.Email, a.Password, a.Role)

	err = uc.u.Create(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "could not create user")
	}

	return u, nil
}

// Update ...
func (uc *UseCase) Update(ctx context.Context, accountID, id domain.ID, a domain.UpdateArgs) (*domain.User, error) {
	se, err := domain.RequireSession(ctx)
	if err != nil {
		return nil, err
	}

	// Admins can update any user in their account.
	if se.User.Role != domain.AdminRole {
		if se.User.ID != id {
			return nil, errors.Errorf("current user %s.%s (role: %s) cannot update user %s.%s", se.User.AccountID, se.User.ID, se.User.Role, accountID, id)
		}
	}

	old, err := uc.u.Get(ctx, se.User.AccountID, id)
	if err != nil {
		return nil, errors.Wrapf(err, "could find user %s.%s", se.User.AccountID, id)
	}

	var updates []domain.Field

	if a.Email != "" {
		updates = append(updates, domain.Field{Name: "email", Value: a.Email})
	}
	if a.Password != "" {
		old.SetPassword(a.Password)
		updates = append(updates, domain.Field{Name: "password_hash", Value: old.PasswordHash})
	}

	up, err := uc.u.Update(ctx, se.User.AccountID, id, updates)
	if err != nil {
		return nil, errors.Wrapf(err, "could not update user %s.%s", se.User.AccountID, id)
	}

	return up, nil
}
