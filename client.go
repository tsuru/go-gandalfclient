package gandalf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Endpoint string
}

// repository represents a git repository.
type repository struct {
	Name     string   `json:"name"`
	Users    []string `json:"users"`
	IsPublic bool     `json:"ispublic"`
}

// repository represents a git user.
type user struct {
	Name string            `json:"name"`
	Keys map[string]string `json:"keys"`
}

type httpError struct {
	code   int
	reason string
}

func (e *httpError) Error() string {
	return e.reason
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, c.Endpoint+path, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return nil, errors.New("Failed to connect to Gandalf server, it's probably down.")
	}
	return response, nil
}

func (c *Client) formatBody(b interface{}) (*bytes.Buffer, error) {
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

func (c *Client) post(b interface{}, path string) error {
	body, err := c.formatBody(b)
	if err != nil {
		return err
	}
	response, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		return &httpError{code: response.StatusCode, reason: string(b)}
	}
	return nil
}

func (c *Client) delete(b interface{}, path string) error {
	body, err := c.formatBody(b)
	if err != nil {
		return err
	}
	response, err := c.doRequest("DELETE", path, body)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		return &httpError{code: response.StatusCode, reason: string(b)}
	}
	return err
}

func (c *Client) get(path string) error {
	response, err := c.doRequest("GET", path, nil)
	if response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		return &httpError{code: response.StatusCode, reason: string(b)}
	}
	return err
}

// NewRepository creates a new repository with a given name and,
// grants access to a list of users
// and defines whether the repository is public.
func (c *Client) NewRepository(name string, users []string, isPublic bool) (repository, error) {
	r := repository{Name: name, Users: users, IsPublic: isPublic}
	if err := c.post(r, "/repository"); err != nil {
		return repository{}, err
	}
	return r, nil
}

// NewUser creates a new user with her/his given keys.
func (c *Client) NewUser(name string, keys map[string]string) (user, error) {
	u := user{Name: name, Keys: keys}
	if err := c.post(u, "/user"); err != nil {
		return user{}, err
	}
	return u, nil
}

// RemoveUser removes a user.
func (c *Client) RemoveUser(name string) error {
	return c.delete(nil, "/user/"+name)
}

// RemoveRepository removes a repository.
func (c *Client) RemoveRepository(name string) error {
	return c.delete(nil, "/repository/"+name)
}

// GrantAccess grants access to N users into N repositories.
func (c *Client) GrantAccess(rNames, uNames []string) error {
	b := map[string][]string{"repositories": rNames, "users": uNames}
	return c.post(b, "/repository/grant")
}

// RevokeAccess revokes access from N users from N repositories.
func (c *Client) RevokeAccess(rNames, uNames []string) error {
	b := map[string][]string{"repositories": rNames, "users": uNames}
	return c.delete(b, "/repository/revoke")
}

// AddKey adds keys to the user.
func (c *Client) AddKey(uName string, key map[string]string) error {
	url := fmt.Sprintf("/user/%s/key", uName)
	return c.post(key, url)
}

// RemoveKey removes the key from the user.
func (c *Client) RemoveKey(uName, kName string) error {
	url := fmt.Sprintf("/user/%s/key/%s", uName, kName)
	return c.delete(nil, url)
}
