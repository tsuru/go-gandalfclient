package gandalfclient

import (
	"bytes"
    "encoding/json"
    "fmt"
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

func (c *Client) NewRepository(name string, users []string, isPublic bool) (repository, error) {
    r := repository{Name: name, Users: users, IsPublic: isPublic}
    j, err := json.Marshal(&r)
    if err != nil {
        return repository{}, err
    }
	body := bytes.NewBuffer(j)
	request, err := http.NewRequest("POST", c.Endpoint, body)
    if err != nil {
        return repository{}, err
    }
    request.Header.Set("Content-Type", "application/json")
    response, err := (&http.Client{}).Do(request)
    if err != nil {
        return repository{}, err
    }
    if response.StatusCode != 200 {
        b, _ := ioutil.ReadAll(response.Body)
        msg := fmt.Errorf("Got error while performing request. Code: %s - Message: %s", response.StatusCode, b)
        return repository{}, msg
    }
    return r, nil
}
