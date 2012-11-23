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
	response, err := client.doRequest("POST", "/test", body)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, 200)
	c.Assert(string(h.body), Equals, `{"foo":"bar"}`)
	c.Assert(h.url, Equals, "/test")
}

func (s *S) TestDoRequestShouldNotSetContentTypeToJsonWhenBodyIsNil(c *C) {
	h := TestHandler{content: `some return message`}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	response, err := client.doRequest("DELETE", "/test", nil)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, 200)
	c.Assert(h.header.Get("Content-Type"), Not(Equals), "application/json")
}

func (s *S) TestPost(c *C) {
	h := TestHandler{content: `some return message`}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	r := repository{Name: "test", Users: []string{"samwan"}}
	err := client.post(r, "/repository")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/repository")
	c.Assert(h.method, Equals, "POST")
	c.Assert(string(h.body), Equals, `{"name":"test","users":["samwan"],"ispublic":false}`)
}

func (s *S) TestPostWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	r := repository{Name: "test", Users: []string{"samwan"}}
	err := client.post(r, "/repository")
	c.Assert(err, ErrorMatches, "^Got error while performing request. Code: 400 - Message: Error performing requested operation\n$")
}

func (s *S) TestDelete(c *C) {
	h := TestHandler{content: `some return message`}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	err := client.delete("/user/someuser")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/user/someuser")
	c.Assert(h.method, Equals, "DELETE")
}

func (s *S) TestDeleteWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	err := client.delete("/user/someuser")
	c.Assert(err, ErrorMatches, "^Got error while performing request. Code: 400 - Message: Error performing requested operation\n$")
}

func (s *S) TestNewRepository(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	_, err := client.NewRepository("proj1", []string{"someuser"}, false)
	c.Assert(err, IsNil)
	c.Assert(string(h.body), Equals, `{"name":"proj1","users":["someuser"],"ispublic":false}`)
	c.Assert(h.url, Equals, "/repository")
	c.Assert(h.method, Equals, "POST")
}

func (s *S) TestNewRepositoryWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	_, err := client.NewRepository("proj1", []string{"someuser"}, false)
	expected := "^Got error while performing request. Code: 400 - Message: Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestNewUser(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	_, err := client.NewUser("someuser", map[string]string{"testkey": "ssh-rsa somekey"})
	c.Assert(err, IsNil)
	c.Assert(string(h.body), Equals, `{"name":"someuser","keys":{"testkey":"ssh-rsa somekey"}}`)
	c.Assert(h.url, Equals, "/user")
	c.Assert(h.method, Equals, "POST")
}

func (s *S) TestNewUserWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	_, err := client.NewUser("someuser", map[string]string{"testkey": "ssh-rsa somekey"})
	expected := "^Got error while performing request. Code: 400 - Message: Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRemoveUser(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	err := client.RemoveUser("someuser")
	c.Assert(err, IsNil)
	c.Assert(string(h.body), Equals, "")
	c.Assert(h.url, Equals, "/user/someuser")
	c.Assert(h.method, Equals, "DELETE")
	c.Assert(h.header.Get("Content-Type"), Not(Equals), "application/json")
}

func (s *S) TestRemoveUserWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	err := client.RemoveUser("someuser")
	expected := "^Got error while performing request. Code: 400 - Message: Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRemoveRepository(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	err := client.RemoveRepository("project1")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/repository/project1")
	c.Assert(h.method, Equals, "DELETE")
	c.Assert(string(h.body), Equals, "")
	c.Assert(h.header.Get("Content-Type"), Not(Equals), "application/json")
}

func (s *S) TestRemoveRepositoryWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	err := client.RemoveRepository("proj2")
	expected := "^Got error while performing request. Code: 400 - Message: Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}
