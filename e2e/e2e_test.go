package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/anrid/codecoach/internal/config"
	acc_c "github.com/anrid/codecoach/internal/controller/account"
	user_c "github.com/anrid/codecoach/internal/controller/user"
	"github.com/anrid/codecoach/internal/domain"
	"github.com/anrid/codecoach/internal/pg"
	account_dao "github.com/anrid/codecoach/internal/pg/dao/account"
	user_dao "github.com/anrid/codecoach/internal/pg/dao/user"
	"github.com/anrid/codecoach/internal/pkg/httpclient"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	acc_uc "github.com/anrid/codecoach/internal/usecase/account"
	user_uc "github.com/anrid/codecoach/internal/usecase/user"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// ts ...
type ts struct {
	suite.Suite
	db                  *sqlx.DB
	a                   *account_dao.DAO
	u                   *user_dao.DAO
	serv                *httpserver.HTTPServer
	logger              *zap.Logger
	APIURL              string
	TestTokenExpiration bool
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(ts))
}

func (su *ts) SetupSuite() {
	// Setup global zap logger.
	su.logger, _ = zap.NewDevelopment()
	zap.ReplaceGlobals(su.logger)

	zap.S().Infow("setup")

	c := config.New()

	// Override server host for testing.
	c.Host = ":10099"
	su.APIURL = "localhost:10099"

	// Override database schema for testing.
	c.DBName = "codecoach_test"

	// Override token expires at.
	c.TokenExpires = 1 * time.Second
	su.TestTokenExpiration = true

	su.db = pg.InitDB(c, true /* drop and recreate db every time */)

	// Setup DAOs.
	su.a = account_dao.New(su.db)
	su.u = user_dao.New(su.db)

	// Setup use cases.
	userUC := user_uc.New(c, su.a, su.u)
	accUC := acc_uc.New(su.a)

	// Setup HTTP server.
	su.serv = httpserver.New(su.u)

	// Setup controller.
	userC := user_c.New(userUC)
	accC := acc_c.New(accUC)

	// Setup routes.
	userC.SetupRoutes(su.serv)
	accC.SetupRoutes(su.serv)

	// Start server.
	go func() {
		_ = su.serv.Echo.Start(c.Host)
	}()

	zap.S().Infow("setup complete")
}

func (su *ts) TearDownSuite() {
	zap.S().Infow("tear down")

	su.db.Close()
	_ = su.serv.Echo.Shutdown(context.Background())
	_ = su.logger.Sync()

	zap.S().Infow("tear down complete")
}

type errorResp struct {
	Error string `json:"error"`
}

// TestAll ...
func (su *ts) TestAll() {
	r := require.New(su.T())

	apiURL := func(path string, args ...interface{}) string {
		if strings.Contains(path, `%`) && len(args) > 0 {
			path = fmt.Sprintf(path, args...)
		}
		return "http://" + su.APIURL + path
	}

	now := time.Now().UnixNano()

	// Signup
	var a1 *domain.Account
	var admin1 *domain.User
	var admin1Token string
	{
		req := user_c.SignupRequest{
			AccountName: fmt.Sprintf("Massa Mun GAI! %d", now),
			GivenName:   "Massa",
			FamilyName:  "Mun",
			Email:       fmt.Sprintf("admin-%d@example.com", now),
			Password:    "massa123",
		}
		res := user_c.SignupResponse{}

		_, _ = httpclient.Call("POST", apiURL("/api/v1/signup"), &req, &res)

		r.NotEmpty(res.Account.ID)
		r.NotEmpty(res.User.ID)
		r.Nil(res.User.UpdatedAt)
		r.Nil(res.User.UpdatedAt)
		r.False(res.Account.CreatedAt.IsZero())
		r.False(res.User.CreatedAt.IsZero())
		r.Empty(res.User.PasswordHash)
		r.Empty(res.User.Token)

		a1 = res.Account
		admin1 = res.User
		admin1Token = res.Token
	}

	// POST /users
	var hm1 *domain.User
	{
		req := user_c.PostUserRequest{
			GivenName:  "Ace",
			FamilyName: "Base",
			Email:      fmt.Sprintf("hm-%d@example.com", now),
			Password:   "massa123",
			Role:       domain.RoleHiringManager,
		}
		res := domain.User{}

		_, _ = httpclient.CallWithToken("POST", apiURL("/api/v1/accounts/%s/users", a1.ID), admin1Token, &req, &res)

		r.NotEmpty(res.ID)
		r.Nil(res.UpdatedAt)
		r.False(res.CreatedAt.IsZero())
		r.Empty(res.PasswordHash)

		hm1 = &res
	}

	// POST /login
	var hm1Token string
	{
		req := user_c.LoginRequest{
			AccountCode: a1.Code,
			Email:       hm1.Email,
			Password:    "massa123",
		}
		res := user_c.LoginResponse{}

		_, _ = httpclient.Call("POST", apiURL("/api/v1/login"), &req, &res)

		r.Equal(hm1.ID, res.User.ID)
		r.Equal(hm1.Email, res.User.Email)
		r.NotEmpty(res.Token)
		r.Equal(60, len(res.Token))

		hm1Token = res.Token
	}

	secretURL := apiURL("/api/v1/accounts/%s/secret", hm1.AccountID)

	// GET /secret
	{
		res := user_c.GetSecretResponse{}

		_, _ = httpclient.CallWithToken("GET", secretURL, hm1Token, nil, &res)

		r.Contains(res.Secret, "All your base")
		r.Equal(hm1.AccountID, res.AccountID)
		r.Equal(hm1.ID, res.ID)
	}

	// FAIL: Missing token
	{
		res := errorResp{}

		_, _ = httpclient.CallWithToken("GET", secretURL, "", nil, &res)

		r.True(strings.Contains(res.Error, "token invalid") || strings.Contains(res.Error, "missing token"))
	}

	// FAIL: Invalid token
	{
		res := errorResp{}

		_, _ = httpclient.CallWithToken("GET", secretURL, "xxx", nil, &res)

		r.Contains(res.Error, "token invalid")
	}

	// PATCH /accounts/{account_id}/users/{id} - update password
	{
		time.Sleep(100 * time.Millisecond)

		req := user_c.PatchUserRequest{
			Password: "123massa",
		}
		res := domain.User{}

		_, _ = httpclient.CallWithToken("PATCH", apiURL("/api/v1/accounts/%s/users/%s", admin1.AccountID, admin1.ID), admin1Token, &req, &res)

		r.Equal(admin1.ID, res.ID)
		r.NotNil(res.UpdatedAt)
		r.False((*res.UpdatedAt).IsZero())
		r.Equal(admin1.Email, res.Email)
	}

	// POST /login (using newly updated password)
	{
		req := user_c.LoginRequest{
			AccountCode: a1.Code,
			Email:       admin1.Email,
			Password:    "123massa",
		}
		res := user_c.LoginResponse{}

		_, _ = httpclient.Call("POST", apiURL("/api/v1/login"), &req, &res)

		r.NotEmpty(res.Token)
		r.Equal(res.User.ID, admin1.ID)

		admin1Token = res.Token
	}

	// PATCH /users/{id} - update email
	{
		time.Sleep(100 * time.Millisecond)

		req := user_c.PatchUserRequest{
			Email: "updated-" + hm1.Email,
		}
		res := domain.User{}

		_, _ = httpclient.CallWithToken("PATCH", apiURL("/api/v1/accounts/%s/users/%s", hm1.AccountID, hm1.ID), hm1Token, &req, &res)

		r.Equal(hm1.ID, res.ID)
		r.NotNil(res.UpdatedAt)
		r.False((*res.UpdatedAt).IsZero())
		r.Equal(req.Email, res.Email)
	}

	// FAIL: Access denied when hiring manager tries to update admin.
	{
		req := user_c.PatchUserRequest{
			Email: "updated-" + hm1.Email,
		}
		res := errorResp{}

		_, _ = httpclient.CallWithToken("PATCH", apiURL("/api/v1/accounts/%s/users/%s", admin1.AccountID, admin1.ID), hm1Token, &req, &res)

		r.Contains(res.Error, "could not update user")
	}

	{
		// FAIL: Email validation.
		{
			req := user_c.PatchUserRequest{
				Email: "not@valid",
			}
			res := errorResp{}

			_, _ = httpclient.CallWithToken("PATCH", apiURL("/api/v1/accounts/%s/users/%s", admin1.AccountID, admin1.ID), admin1Token, &req, &res)

			r.Contains(res.Error, "validation error: email")
		}

		// FAIL: Password validation.
		{
			req := user_c.PatchUserRequest{
				Password: "toobad!",
			}
			res := errorResp{}

			_, _ = httpclient.CallWithToken("PATCH", apiURL("/api/v1/accounts/%s/users/%s", admin1.AccountID, admin1.ID), admin1Token, &req, &res)

			r.Contains(res.Error, "validation error: password")
		}
	}

	// Update account name and logo.
	{
		req := acc_c.PatchAccountRequest{
			Name: "updated-" + a1.Name,
			Logo: "https://static.wikia.nocookie.net/streetfighter/images/1/1b/AngryChargeTigerUppercut.jpg/revision/latest/scale-to-width-down/1000",
		}
		res := domain.Account{}

		_, _ = httpclient.CallWithToken("PATCH", apiURL("/api/v1/accounts/%s", a1.ID), admin1Token, &req, &res)

		r.Equal(req.Name, res.Name)
		r.Equal(req.Logo, res.Profile.Logo)
	}

	// FAIL: Hiring manager cannot update account.
	{
		req := acc_c.PatchAccountRequest{
			Name: "updated-" + a1.Name,
		}
		res := errorResp{}

		_, _ = httpclient.CallWithToken("PATCH", apiURL("/api/v1/accounts/%s", a1.ID), hm1Token, &req, &res)

		r.Contains(res.Error, "could not update account")
	}

	// List all users as admin.
	{
		res := user_c.GetListResponse{}

		_, _ = httpclient.CallWithToken("GET", apiURL("/api/v1/accounts/%s/users", a1.ID), admin1Token, nil, &res)

		r.True(len(res.Users) == 2)
	}

	// List all users as hiring manager.
	{
		res := user_c.GetListResponse{}

		_, _ = httpclient.CallWithToken("GET", apiURL("/api/v1/accounts/%s/users", a1.ID), hm1Token, nil, &res)

		r.True(len(res.Users) == 1)
		r.Equal(hm1.ID, res.Users[0].ID)
	}

	if su.TestTokenExpiration {
		msSinceTestStarted := (time.Now().UnixNano() - now) / 1000000
		r.True(msSinceTestStarted < 1000, "tests should only have been running for a maximum of 1,000 ms at this point")

		// Wait until token has expired!
		time.Sleep(1000 * time.Millisecond)

		res := errorResp{}

		_, _ = httpclient.CallWithToken("GET", secretURL, admin1Token, nil, &res)

		r.Equal("token expired", res.Error)
	}
}
