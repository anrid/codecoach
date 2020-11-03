package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	ctrl "github.com/anrid/codecoach/internal/controller/user"
	"github.com/anrid/codecoach/internal/domain"
	"github.com/stretchr/testify/require"
)

// Options ...
type Options struct {
	Host                string
	TestTokenExpiration bool
}

// AllTests ...
func AllTests(r *require.Assertions, o Options) {
	type errorResp struct {
		Error string `json:"error"`
	}

	apiURL := func(path string, args ...interface{}) string {
		if strings.Contains(path, `%`) && len(args) > 0 {
			path = fmt.Sprintf(path, args...)
		}
		return "http://" + o.Host + path
	}

	now := time.Now().UnixNano()

	// Signup
	var a1 *domain.Account
	var u1 *domain.User
	var u1Token string
	{
		req := ctrl.SignupRequest{
			AccountName: fmt.Sprintf("Massa Mun GAI! %d", now),
			Email:       fmt.Sprintf("admin-%d@example.com", now),
			Password:    "massa123",
		}
		res := ctrl.SignupResponse{}

		call("POST", apiURL("/api/v1/signup"), &req, &res)

		r.NotEmpty(res.Account.ID)
		r.NotEmpty(res.User.ID)
		r.Nil(res.User.UpdatedAt)
		r.Nil(res.User.UpdatedAt)
		r.False(res.Account.CreatedAt.IsZero())
		r.False(res.User.CreatedAt.IsZero())
		r.Empty(res.User.PasswordHash)
		r.Empty(res.User.Token)

		a1 = res.Account
		u1 = res.User
		u1Token = res.Token
	}

	// domain.Dump(a1)
	// domain.Dump(u1)
	// domain.Dump(u1Token)

	// POST /users
	var u2 *domain.User
	{
		req := ctrl.PostUserRequest{
			Email:    fmt.Sprintf("hm-%d@example.com", now),
			Password: "massa123",
			Role:     domain.HiringManagerRole,
		}
		res := domain.User{}

		_call("POST", apiURL("/api/v1/accounts/%s/users", a1.ID), u1Token, &req, &res)

		r.NotEmpty(res.ID)
		r.Nil(res.UpdatedAt)
		r.False(res.CreatedAt.IsZero())
		r.Empty(res.PasswordHash)

		u2 = &res
	}

	// POST /login
	var u2Token string
	{
		req := ctrl.LoginRequest{
			AccountCode: a1.Code,
			Email:       u2.Email,
			Password:    "massa123",
		}
		res := ctrl.LoginResponse{}

		call("POST", apiURL("/api/v1/login"), &req, &res)

		r.Equal(u2.ID, res.User.ID)
		r.Equal(u2.Email, res.User.Email)
		r.NotEmpty(res.Token)
		r.Equal(60, len(res.Token))

		u2Token = res.Token
	}

	secretURL := apiURL("/api/v1/accounts/%s/secret", u2.AccountID)

	// GET /secret
	{
		res := ctrl.SecretResponse{}

		_call("GET", secretURL, u2Token, nil, &res)

		r.Contains(res.Secret, "All your base")
		r.Equal(u2.AccountID, res.AccountID)
		r.Equal(u2.ID, res.ID)
	}

	// Missing token
	{
		res := errorResp{}

		_call("GET", secretURL, "", nil, &res)

		r.True(strings.Contains(res.Error, "token invalid") || strings.Contains(res.Error, "missing token"))
	}

	// Invalid token
	{
		res := errorResp{}

		_call("GET", secretURL, "xxx", nil, &res)

		r.Contains(res.Error, "token invalid")
	}

	// PATCH /accounts/{account_id}/users/{id} - update password
	{
		time.Sleep(100 * time.Millisecond)

		req := ctrl.PatchUserRequest{
			Password: "123massa",
		}
		res := domain.User{}

		_call("PATCH", apiURL("/api/v1/accounts/%s/users/%s", u1.AccountID, u1.ID), u1Token, &req, &res)

		r.Equal(u1.ID, res.ID)
		r.NotNil(res.UpdatedAt)
		r.False((*res.UpdatedAt).IsZero())
		r.Equal(u1.Email, res.Email)
	}

	// POST /login (using newly updated password)
	{
		req := ctrl.LoginRequest{
			AccountCode: a1.Code,
			Email:       u1.Email,
			Password:    "123massa",
		}
		res := ctrl.LoginResponse{}

		call("POST", apiURL("/api/v1/login"), &req, &res)

		r.NotEmpty(res.Token)
		r.Equal(res.User.ID, u1.ID)

		u1Token = res.Token
	}

	// PATCH /users/{id} - update email
	{
		time.Sleep(100 * time.Millisecond)

		req := ctrl.PatchUserRequest{
			Email: "updated-" + u2.Email,
		}
		res := domain.User{}

		_call("PATCH", apiURL("/api/v1/accounts/%s/users/%s", u2.AccountID, u2.ID), u2Token, &req, &res)

		r.Equal(u2.ID, res.ID)
		r.NotNil(res.UpdatedAt)
		r.False((*res.UpdatedAt).IsZero())
		r.Equal(req.Email, res.Email)
	}

	// Access denied when hiring manager tries to update admin.
	{
		req := ctrl.PatchUserRequest{
			Email: "updated-" + u2.Email,
		}
		res := errorResp{}

		_call("PATCH", apiURL("/api/v1/accounts/%s/users/%s", u1.AccountID, u1.ID), u2Token, &req, &res)

		r.Contains(res.Error, "access denied")
	}

	{
		// Email validation fails.
		{
			req := ctrl.PatchUserRequest{
				Email: "not@valid",
			}
			res := errorResp{}

			_call("PATCH", apiURL("/api/v1/accounts/%s/users/%s", u1.AccountID, u1.ID), u1Token, &req, &res)

			r.Contains(res.Error, "validation error: email")
		}

		// Password validation fails.
		{
			req := ctrl.PatchUserRequest{
				Password: "toobad!",
			}
			res := errorResp{}

			_call("PATCH", apiURL("/api/v1/accounts/%s/users/%s", u1.AccountID, u1.ID), u1Token, &req, &res)

			r.Contains(res.Error, "validation error: password")
		}
	}

	if o.TestTokenExpiration {
		// Token expired!
		time.Sleep(3000 * time.Millisecond)

		res := errorResp{}

		_call("GET", secretURL, u1Token, nil, &res)

		r.Equal("token expired", res.Error)
	}
}

func call(method, url string, in, out interface{}) string {
	return _call(method, url, "", in, out)
}

func _call(method, url, token string, in, out interface{}) string {
	log.Printf("calling url: %s %s  --  token: %s", strings.ToUpper(method), url, token)

	client := &http.Client{}

	var payload io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			panic(err)
		}
		log.Printf("payload: %s", string(b))
		payload = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, payload)

	// Always expect a JSON response.
	req.Header.Add("Content-Type", "application/json")

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	dataPreview := data
	if len(data) > 1000 {
		dataPreview = data[0:1000]
	}

	log.Printf("got response: %s\n", string(dataPreview))

	err = json.Unmarshal(data, out)
	if err != nil {
		println("Error: " + err.Error())
		println("Could not unmarshal response body:")
		println(string(dataPreview))
	}

	return string(dataPreview)
}
