// Copyright 2015 go-gandalfclient authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gandalf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var GitTimeFormat = "Mon Jan _2 15:04:05 2006 -0700"

type Client struct {
	Endpoint string
	Client   *http.Client
}

// repository represents a git repository.
type repository struct {
	Name     string   `json:"name"`
	Users    []string `json:"users"`
	IsPublic bool     `json:"ispublic"`
	SSHURL   string   `json:"ssh_url,omitempty"`
	GitURL   string   `json:"git_url,omitempty"`
}

// repository represents a git user.
type user struct {
	Name string            `json:"name"`
	Keys map[string]string `json:"keys"`
}

type Author struct {
	Name  string
	Email string
	Date  GitTime
}

type Commit struct {
	Ref       string
	Author    Author
	Committer Author
	Subject   string
	CreatedAt GitTime
	Parent    []string
}

type GitTime time.Time

func (c *GitTime) UnmarshalJSON(raw []byte) error {
	strRaw := string(raw)
	if strRaw == `""` || strRaw == "null" {
		return nil
	}
	t, err := time.Parse(`"`+GitTimeFormat+`"`, strRaw)
	if err != nil {
		t, err = time.Parse(`"`+time.RFC3339+`"`, strRaw)
		if err != nil {
			return err
		}
	}
	*c = GitTime(t)
	return nil
}

type Log struct {
	Commits []Commit
	Next    string
}

type HTTPError struct {
	Code   int
	Reason string
}

func (e *HTTPError) Error() string {
	return e.Reason
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	endpoint := strings.TrimRight(c.Endpoint, "/")
	request, err := http.NewRequest(method, endpoint+path, body)
	if err != nil {
		return nil, errors.New("invalid Gandalf endpoint")
	}
	request = request.WithContext(ctx)
	request.Close = true
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Gandalf server (%s) - %s", c.Endpoint, err.Error())
	}
	return response, nil
}

func (c *Client) formatBody(b interface{}) (*bytes.Buffer, error) {
	if str, ok := b.(string); ok {
		return bytes.NewBufferString(str), nil
	}
	body := bytes.NewBufferString("null")
	if b != nil {
		j, err := json.Marshal(&b)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(j)
	}
	return body, nil
}

func (c *Client) post(ctx context.Context, b interface{}, path string) error {
	body, err := c.formatBody(b)
	if err != nil {
		return err
	}
	response, err := c.doRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		return &HTTPError{Code: response.StatusCode, Reason: string(b)}
	}
	return nil
}

func (c *Client) put(ctx context.Context, b, path string) error {
	response, err := c.doRequest(ctx, "PUT", path, strings.NewReader(b))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		return &HTTPError{Code: response.StatusCode, Reason: string(b)}
	}
	return nil
}

func (c *Client) delete(ctx context.Context, b interface{}, path string) error {
	body, err := c.formatBody(b)
	if err != nil {
		return err
	}
	response, err := c.doRequest(ctx, "DELETE", path, body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		return &HTTPError{Code: response.StatusCode, Reason: string(b)}
	}
	return err
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	response, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return []byte{}, &HTTPError{Code: 500, Reason: err.Error()}
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return []byte{}, &HTTPError{Code: response.StatusCode, Reason: string(b)}
	}
	return b, err
}

// NewRepository creates a new repository with a given name and,
// grants access to a list of users
// and defines whether the repository is public.
func (c *Client) NewRepository(ctx context.Context, name string, users []string, isPublic bool) (repository, error) {
	r := repository{Name: name, Users: users, IsPublic: isPublic}
	if err := c.post(ctx, r, "/repository"); err != nil {
		return repository{}, err
	}
	return r, nil
}

// GetRepository gets metadata from a repository in Gandalf server.
func (c *Client) GetRepository(ctx context.Context, name string) (repository, error) {
	url := fmt.Sprintf("/repository/%s?:name=%s", name, name)
	b, err := c.get(ctx, url)
	if err != nil {
		return repository{}, err
	}
	var r repository
	if err := json.Unmarshal(b, &r); err != nil {
		return repository{}, fmt.Errorf("Caught error decoding returned json: %s", err.Error())
	}
	return r, nil
}

// NewUser creates a new user with her/his given keys.
func (c *Client) NewUser(ctx context.Context, name string, keys map[string]string) (user, error) {
	u := user{Name: name, Keys: keys}
	if err := c.post(ctx, u, "/user"); err != nil {
		return user{}, err
	}
	return u, nil
}

// RemoveUser removes a user.
func (c *Client) RemoveUser(ctx context.Context, name string) error {
	return c.delete(ctx, nil, "/user/"+name)
}

// RemoveRepository removes a repository.
func (c *Client) RemoveRepository(ctx context.Context, name string) error {
	return c.delete(ctx, nil, "/repository/"+name)
}

// GrantAccess grants access to N users into N repositories.
func (c *Client) GrantAccess(ctx context.Context, rNames, uNames []string) error {
	b := map[string][]string{"repositories": rNames, "users": uNames}
	return c.post(ctx, b, "/repository/grant")
}

// RevokeAccess revokes access from N users from N repositories.
func (c *Client) RevokeAccess(ctx context.Context, rNames, uNames []string) error {
	b := map[string][]string{"repositories": rNames, "users": uNames}
	return c.delete(ctx, b, "/repository/revoke")
}

// AddKey adds keys to the user.
func (c *Client) AddKey(ctx context.Context, uName string, key map[string]string) error {
	url := fmt.Sprintf("/user/%s/key", uName)
	return c.post(ctx, key, url)
}

func (c *Client) UpdateKey(ctx context.Context, uName, kName, kBody string) error {
	url := fmt.Sprintf("/user/%s/key/%s", uName, kName)
	return c.put(ctx, kBody, url)
}

// RemoveKey removes the key from the user.
func (c *Client) RemoveKey(ctx context.Context, uName, kName string) error {
	url := fmt.Sprintf("/user/%s/key/%s", uName, kName)
	return c.delete(ctx, nil, url)
}

// ListKeys retrieves all keys a given user has
func (c *Client) ListKeys(ctx context.Context, uName string) (map[string]string, error) {
	url := fmt.Sprintf("/user/%s/keys", uName)
	resp, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	keys := map[string]string{}
	err = json.Unmarshal(resp, &keys)
	return keys, err
}

//GetDiff gets diff output between commits from a repository in Gandalf server.
func (c *Client) GetDiff(ctx context.Context, repo, previousCommit, lastCommit string) (string, error) {
	url := fmt.Sprintf("/repository/%s/diff/commits?:name=%s&previous_commit=%s&last_commit=%s", repo, repo, previousCommit, lastCommit)
	diffOutput, err := c.get(ctx, url)
	if err != nil {
		return "", fmt.Errorf("Caught error getting repository metadata: %s", err.Error())
	}
	return string(diffOutput), nil
}

func (c *Client) GetLog(ctx context.Context, repo, ref, path string, total int) (Log, error) {
	v := url.Values{}
	v.Set("ref", ref)
	if path != "" {
		v.Set("path", path)
	}
	if total > 0 {
		v.Set("total", strconv.Itoa(total))
	}
	u := fmt.Sprintf("/repository/%s/logs?%s", repo, v.Encode())
	var ret Log
	output, err := c.get(ctx, u)
	if err != nil {
		return ret, fmt.Errorf("Caught error getting repository log: %s", err.Error())
	}
	err = json.Unmarshal(output, &ret)
	return ret, err
}

//GetHealthCheck gets healthcheck request output in Gandalf server.
func (c *Client) GetHealthCheck(ctx context.Context) ([]byte, error) {
	result, err := c.get(ctx, "/healthcheck")
	if err != nil {
		return []byte{}, &HTTPError{Code: 500, Reason: err.Error()}
	}
	return result, nil
}
