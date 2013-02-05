package gandalf

import (
	"bytes"
	"encoding/json"
	. "launchpad.net/gocheck"
	"net/http/httptest"
)

func (s *S) TestDoRequest(c *C) {
	h := TestHandler{content: `some return message`}
	ts := httptest.NewServer(&h)
	defer ts.Close()
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
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	response, err := client.doRequest("DELETE", "/test", nil)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, 200)
	c.Assert(h.header.Get("Content-Type"), Not(Equals), "application/json")
}

func (s *S) TestDoRequestConnectionError(c *C) {
	client := Client{Endpoint: "http://127.0.0.1:747399"}
	response, err := client.doRequest("GET", "/", nil)
	c.Assert(response, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Failed to connect to Gandalf server, it's probably down.")
}

func (s *S) TestPost(c *C) {
	h := TestHandler{content: `some return message`}
	ts := httptest.NewServer(&h)
	defer ts.Close()
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
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	r := repository{Name: "test", Users: []string{"samwan"}}
	err := client.post(r, "/repository")
	c.Assert(err, ErrorMatches, "^Error performing requested operation\n$")
}

func (s *S) TestPostConnectionFailure(c *C) {
	client := Client{Endpoint: "http://127.0.0.1:747399"}
	err := client.post(nil, "/")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Failed to connect to Gandalf server, it's probably down.")
}

func (s *S) TestDelete(c *C) {
	h := TestHandler{content: `some return message`}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.delete(nil, "/user/someuser")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/user/someuser")
	c.Assert(h.method, Equals, "DELETE")
	c.Assert(string(h.body), Equals, "null")
}

func (s *S) TestDeleteWithConnectionError(c *C) {
	client := Client{Endpoint: "http://127.0.0.1:747399"}
	err := client.delete(nil, "/users/something")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Failed to connect to Gandalf server, it's probably down.")
}

func (s *S) TestDeleteWithResponseError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.delete(nil, "/user/someuser")
	c.Assert(err, ErrorMatches, "^Error performing requested operation\n$")
	c.Assert(string(h.body), Equals, "null")
}

func (s *S) TestDeleteWithBody(c *C) {
	h := TestHandler{content: `some return message`}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.delete(map[string]string{"test": "foo"}, "/user/someuser")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/user/someuser")
	c.Assert(h.method, Equals, "DELETE")
	c.Assert(string(h.body), Equals, `{"test":"foo"}`)
}

func (s *S) TestGet(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.get("/user/someuser")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/user/someuser")
	c.Assert(h.method, Equals, "GET")
}

func (s *S) TestGetWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.get("/user/someuser")
	c.Assert(err, ErrorMatches, "^Error performing requested operation\n$")
}

func (s *S) TestFormatBody(c *C) {
	b, err := (&Client{}).formatBody(map[string]string{"test": "foo"})
	c.Assert(err, IsNil)
	c.Assert(b.String(), Equals, `{"test":"foo"}`)
}

func (s *S) TestFormatBodyReturnJsonNullWithNilBody(c *C) {
	b, err := (&Client{}).formatBody(nil)
	c.Assert(err, IsNil)
	c.Assert(b.String(), Equals, "null")
}

func (s *S) TestNewRepository(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
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
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.NewRepository("proj1", []string{"someuser"}, false)
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestNewUser(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
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
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.NewUser("someuser", map[string]string{"testkey": "ssh-rsa somekey"})
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRemoveUser(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveUser("someuser")
	c.Assert(err, IsNil)
	c.Assert(string(h.body), Equals, "null")
	c.Assert(h.url, Equals, "/user/someuser")
	c.Assert(h.method, Equals, "DELETE")
}

func (s *S) TestRemoveUserWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveUser("someuser")
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRemoveRepository(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveRepository("project1")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/repository/project1")
	c.Assert(h.method, Equals, "DELETE")
	c.Assert(string(h.body), Equals, "null")
}

func (s *S) TestRemoveRepositoryWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveRepository("proj2")
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestAddKey(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	key := map[string]string{"pubkey": "ssh-rsa somekey me@myhost"}
	err := client.AddKey("username", key)
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/user/username/key")
	c.Assert(h.method, Equals, "POST")
	c.Assert(string(h.body), Equals, `{"pubkey":"ssh-rsa somekey me@myhost"}`)
	c.Assert(h.header.Get("Content-Type"), Equals, "application/json")
}

func (s *S) TestAddKeyWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.AddKey("proj2", map[string]string{"key": "ssh-rsa keycontent user@host"})
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRemoveKey(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveKey("username", "keyname")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/user/username/key/keyname")
	c.Assert(h.method, Equals, "DELETE")
	c.Assert(string(h.body), Equals, "null")
}

func (s *S) TestRemoveKeyWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveKey("proj2", "keyname")
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestGrantAccess(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	repositories := []string{"projectx", "projecty"}
	users := []string{"userx"}
	err := client.GrantAccess(repositories, users)
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/repository/grant")
	c.Assert(h.method, Equals, "POST")
	expected, err := json.Marshal(map[string][]string{"repositories": repositories, "users": users})
	c.Assert(err, IsNil)
	c.Assert(h.body, DeepEquals, expected)
	c.Assert(h.header.Get("Content-Type"), Equals, "application/json")
}

func (s *S) TestGrantAccessWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.GrantAccess([]string{"projectx", "projecty"}, []string{"userx"})
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRevokeAccess(c *C) {
	h := TestHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	repositories := []string{"projectx", "projecty"}
	users := []string{"userx"}
	err := client.RevokeAccess(repositories, users)
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/repository/revoke")
	c.Assert(h.method, Equals, "DELETE")
	expected, err := json.Marshal(map[string][]string{"repositories": repositories, "users": users})
	c.Assert(err, IsNil)
	c.Assert(h.body, DeepEquals, expected)
	c.Assert(h.header.Get("Content-Type"), Equals, "application/json")
}

func (s *S) TestRevokeAccessWithError(c *C) {
	h := ErrorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RevokeAccess([]string{"projectx", "projecty"}, []string{"usery"})
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}
