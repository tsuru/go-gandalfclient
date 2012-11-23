package gandalfclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Endpoint string
}

type repository struct {
	Name     string   `json:"name"`
	Users    []string `json:"users"`
	IsPublic bool     `json:"ispublic"`
}

type user struct {
	Name string            `json:"name"`
	Keys map[string]string `json:"keys"`
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
		return response, err
	}
	return response, nil
}

func (c *Client) post(i interface{}, path string) error {
	j, err := json.Marshal(&i)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(j)
	response, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		err := fmt.Errorf("Got error while performing request. Code: %d - Message: %s", response.StatusCode, b)
		return err
	}
	return nil
}

func (c *Client) delete(path string) error {
	response, err := c.doRequest("DELETE", path, nil)
	if response.StatusCode != 200 {
		b, _ := ioutil.ReadAll(response.Body)
		err := fmt.Errorf("Got error while performing request. Code: %d - Message: %s", response.StatusCode, b)
		return err
	}
	return err
}

func (c *Client) NewRepository(name string, users []string, isPublic bool) (repository, error) {
	r := repository{Name: name, Users: users, IsPublic: isPublic}
	if err := c.post(r, "/repository"); err != nil {
		return repository{}, err
	}
	return r, nil
}

func (c *Client) NewUser(name string, keys map[string]string) (user, error) {
	u := user{Name: name, Keys: keys}
	if err := c.post(u, "/user"); err != nil {
		return user{}, err
	}
	return u, nil
}

func (c *Client) RemoveUser(name string) error {
	return c.delete("/user/" + name)
}
