package github

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const githubAPIV3 = "https://api.github.com"
const acceptGithubAPIJSON = "application/vnd.github.v3+json"

// User ...
type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Repo ...
type Repo struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	FullName   string `json:"full_name"`
	Private    bool   `json:"private"`
	Visibility string `json:"visibility"`
	Owner      struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	}
	BranchesURL      string `json:"branches_url"`
	CollaboratorsURL string `json:"collaborators_url"`
	PullsURL         string `json:"pulls_url"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	Fork             bool   `json:"fork"`
	HTMLURL          string `json:"html_url"`
}

// PullRequest ...
type PullRequest struct {
	ID     int64  `json:"id"`
	URL    string `json:"url"`
	State  string `json:"state"`
	Title  string `json:"title"`
	Number int    `json:"number"`
	User   struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	}
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Review ...
type Review struct {
	ID    int64  `json:"id"`
	State string `json:"state"`
	User  struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	}
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Branch ...
type Branch struct {
	Name   string `json:"name"`
	Commit struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
}

// API ...
type API struct {
	token string
}

// New ...
func New(token string) *API {
	return &API{token}
}

// SanityCheck that simply calls https://api.github.com/ and
// returns `current_user_url`.
// This ensures that we're able to reach Github's API and
// receive a valid response.
func (g *API) SanityCheck() bool {
	res := struct {
		CurrentUserURL string `json:"current_user_url"`
	}{}

	_, err := g.requestJSON(http.MethodGet, "/", nil, &res)
	if err != nil {
		panic(err)
	}

	return res.CurrentUserURL == githubAPIV3+"/user"
}

// CurrentUser returns the current Github user, i.e. the token owner.
func (g *API) CurrentUser() (*User, error) {
	u := new(User)

	_, err := g.requestJSON(http.MethodGet, "/user", nil, u)

	return u, err
}

// User returns a Github user.
func (g *API) User(username string) (*User, error) {
	u := new(User)
	_, err := g.requestJSON(http.MethodGet, "/users/"+username, nil, u)
	return u, err
}

// NewRepo creates a new Github repo for the given candidate.
func (g *API) NewRepo(candidate, name, desc string) (repo string, err error) {
	h := md5.New()
	_, _ = io.WriteString(h, name)
	_, _ = io.WriteString(h, time.Now().String())

	repo = fmt.Sprintf("%s-candidate-%s-%x", name, candidate, h.Sum(nil))

	req := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
		Visibility  string `json:"visibility"`
	}{
		Name:        repo,
		Description: desc,
		Private:     true,
		Visibility:  "private",
	}

	res := struct {
		ID         int64
		Name       string
		FullName   string
		Private    bool
		Visibility string
		Owner      struct {
			ID    int64
			Login string
		}
	}{}

	_, err = g.requestJSON(http.MethodPost, "/user/repos", req, &res)

	fmt.Printf("Created new repo: id=%d name=%s private=%t\n", res.ID, res.Name, res.Private)
	return repo, err
}

// ListRepos lists repos for the current user.
func (g *API) ListRepos(nextURL string) (repos []Repo, res *Response, err error) {
	url := "/user/repos?sort=updated&direction=desc&visibility=private"
	if nextURL != "" {
		url = nextURL
	}

	res, err = g.requestJSON(http.MethodGet, url, nil, &repos)
	return
}

// ListPulls lists pull requests for the given repo.
func (g *API) ListPulls(username, repo string) (prs []PullRequest, res *Response, err error) {
	res, err = g.requestJSON(http.MethodGet, "/repos/"+username+"/"+repo+"/pulls?state=all", nil, &prs)
	return
}

// ListReviews lists pull requests for the given repo.
func (g *API) ListReviews(username, repo string, pullNumber int) (rs []Review, res *Response, err error) {
	res, err = g.requestJSON(http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", username, repo, pullNumber), nil, &rs)
	return
}

// ListBranches lists branches for the given repo.
func (g *API) ListBranches(username, repo string) (bs []Branch, res *Response, err error) {
	res, err = g.requestJSON(http.MethodGet, fmt.Sprintf("/repos/%s/%s/branches", username, repo), nil, &bs)
	return
}

// DownloadRepo downloads a zipped branch of a repo.
func (g *API) DownloadRepo(username, repo, branch string) (*Response, error) {
	return g.request(http.MethodGet, fmt.Sprintf("/repos/%s/%s/zipball/%s", username, repo, branch), nil)
}

// DeleteRepo deletes a repo.
func (g *API) DeleteRepo(username, repo string) (res *Response, err error) {
	res, err = g.requestJSON(http.MethodDelete, fmt.Sprintf("/repos/%s/%s", username, repo), nil, nil)
	return
}

// Upload uploads files to a Github repo.
// Each file is committed separately as doing bulk uploads isn't supported.
// To upload in bulk see: https://developer.github.com/v3/git/trees/
func (g *API) Upload(owner, repo string, files []string, message string) error {
	// Create files in new repo.
	for _, f := range files {
		// Get basename.
		name := path.Base(f)

		// Create Github path.
		githubPath := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, name)

		// Read file contents as bytes.
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return errors.Wrapf(err, "could not read file %s", f)
		}

		req := struct {
			Message string `json:"message"`
			Content string `json:"content"`
		}{
			Message: message,
			Content: base64.StdEncoding.EncodeToString(content), // Base64 encode file contents.
		}
		res := struct {
			Content struct {
				Name string
				Path string
				Sha  string
			}
			Commit struct {
				Sha string
				URL string `json:"url"`
			}
		}{}

		_, err = g.requestJSON(http.MethodPut, githubPath, &req, &res)
		if err != nil {
			return err
		}

		fmt.Printf("Uploaded file: %s\n", res.Content.Name)
	}

	return nil
}

// AddCollaborator adds a collaborator to a repo.
func (g *API) AddCollaborator(owner, repo, candidate string) error {
	githubPath := fmt.Sprintf("/repos/%s/%s/collaborators/%s", owner, repo, candidate)

	req := struct {
		Permission string `json:"permission"`
	}{
		Permission: "push",
	}
	res := struct {
		ID         int64 `json:"id"`
		Repository struct {
			ID       int64 `json:"id"`
			Name     string
			FullName string
		}
	}{}

	_, err := g.requestJSON(http.MethodPut, githubPath, &req, &res)
	if err != nil {
		return err
	}

	fmt.Printf("Added collaborator: id=%s repo=%s\n", candidate, res.Repository.Name)

	return nil
}

// Create a Github API request and unmarshal response JSON into
// the given struct.
func (g *API) requestJSON(method, path string, in, out interface{}) (*Response, error) {
	var body io.Reader

	if in != nil {
		// Marshal request struct.
		data, err := json.Marshal(in)
		if err != nil {
			return nil, errors.Wrap(err, "could not marshal request into JSON")
		}
		// Wrap bytes in io.Reader.
		body = bytes.NewReader(data)
	}

	res, err := g.request(method, path, body)
	if err != nil {
		fmt.Printf("Request error: %s %s\n", method, path)
		return nil, err
	}

	if out != nil {
		err = json.Unmarshal(res.Body, out)
		if err != nil {
			println("Github response:\n" + string(res.Body))
			return nil, errors.Wrap(err, "could not unmarshal response from JSON")
		}
	}

	return res, nil
}

// Create a Github API request, optionally passing a JSON
// request body, and return the response.
func (g *API) request(method, url string, body io.Reader) (*Response, error) {
	const baseURL = githubAPIV3
	if !strings.Contains(url, "github.com") {
		url = baseURL + url
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 10 * time.Second,
		},
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}

	// Set token header.
	req.Header.Add("Authorization", "token "+g.token)
	req.Header.Add("Accept", acceptGithubAPIJSON)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not make http client call")
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Errorf("got Github API error: status_code=%d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read http client response")
	}

	r := &Response{
		StatusCode: resp.StatusCode,
		Body:       data,
	}

	link := resp.Header.Get("Link")
	if link != "" {
		parts := strings.Split(link, " ")
		for i := 0; i < len(parts); i += 2 {
			if parts[i+1][0:10] == `rel="next"` {
				r.Next = strings.Trim(parts[i], "<>;")
			}
			if parts[i+1][0:10] == `rel="last"` {
				r.Last = strings.Trim(parts[i], "<>;")
			}
		}
	}

	return r, nil
}

// Response represents a low-level Github API response.
type Response struct {
	Body       []byte
	Next       string
	Last       string
	StatusCode int
}

// HasNext ...
func (r *Response) HasNext() bool {
	return r.Next != ""
}

// ReadDir reads files in a directory (non-recursively).
func ReadDir(dir string, excludeFiles ...string) []string {
	dir, _ = filepath.Abs(dir)

	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(errors.Wrap(err, "could not read dir "+dir))
	}

	var files []string
outer:
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		for _, e := range excludeFiles {
			if f.Name() == e {
				continue outer
			}
		}
		files = append(files, path.Join(dir, f.Name()))
	}

	return files
}
