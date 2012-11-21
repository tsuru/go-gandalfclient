package gandalfclient

import (
	"bytes"
	. "launchpad.net/gocheck"
	"net/http/httptest"
)

func (s *S) TestDoRequest(c *C) {
	h := TestHandler{content: `some return message`}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	body := bytes.NewBufferString(`{"foo":"bar"}`)
	response, err := client.doRequest("POST", body)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, 200)
	c.Assert(string(h.body), Equals, `{"foo":"bar"}`)
}

func (s *S) TestNewRepository(c *C) {
	h := TestHandler{content: `Repository "proj1" created successfuly`}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	_, err := client.NewRepository("proj1", []string{"someuser"}, false)
	c.Assert(err, IsNil)
	c.Assert(string(h.body), Equals, `{"name":"proj1","users":["someuser"],"ispublic":false}`)
}

func (s *S) TestNewRepositoryWithError(c *C) {
	h := ErrorHandler{content: `Error creating repository`}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	_, err := client.NewRepository("proj1", []string{"someuser"}, false)
	expected := "^Got error while performing request. Code: 400 - Message: Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}
