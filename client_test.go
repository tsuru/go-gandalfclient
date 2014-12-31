// Copyright 2014 go-gandalfclient authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gandalf

import (
	"bytes"
	"encoding/json"
	"errors"
	. "launchpad.net/gocheck"
	"net/http/httptest"
)

type unmarshable struct{}

func (u unmarshable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("Unmarshable.")
}

func (s *S) TestDoRequest(c *C) {
	h := testHandler{content: `some return message`}
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
	h := testHandler{content: `some return message`}
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
	h := testHandler{content: `some return message`}
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
	h := errorHandler{}
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

func (s *S) TestPostMarshalingFailure(c *C) {
	client := Client{Endpoint: "http://127.0.0.1:747399"}
	err := client.post(unmarshable{}, "/users/something")
	c.Assert(err, NotNil)
	e, ok := err.(*json.MarshalerError)
	c.Assert(ok, Equals, true)
	c.Assert(e.Err.Error(), Equals, "Unmarshable.")
}

func (s *S) TestDelete(c *C) {
	h := testHandler{content: `some return message`}
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

func (s *S) TestDeleteWithMarshalingError(c *C) {
	client := Client{Endpoint: "http://127.0.0.1:747399"}
	err := client.delete(unmarshable{}, "/users/something")
	c.Assert(err, NotNil)
	e, ok := err.(*json.MarshalerError)
	c.Assert(ok, Equals, true)
	c.Assert(e.Err.Error(), Equals, "Unmarshable.")
}

func (s *S) TestDeleteWithResponseError(c *C) {
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.delete(nil, "/user/someuser")
	c.Assert(err, ErrorMatches, "^Error performing requested operation\n$")
	c.Assert(string(h.body), Equals, "null")
}

func (s *S) TestDeleteWithBody(c *C) {
	h := testHandler{content: `some return message`}
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
	h := testHandler{content: `{"fookey": "bar keycontent"}`}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	out, err := client.get("/user/someuser")
	c.Assert(err, IsNil)
	c.Assert(string(out), Equals, `{"fookey": "bar keycontent"}`)
	c.Assert(h.url, Equals, "/user/someuser")
	c.Assert(h.method, Equals, "GET")
}

func (s *S) TestGetWithError(c *C) {
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.get("/user/someuser")
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

func (s *S) TestFormatBodyMarshalingFailure(c *C) {
	client := &Client{}
	b, err := client.formatBody(unmarshable{})
	c.Assert(b, IsNil)
	c.Assert(err, NotNil)
	e, ok := err.(*json.MarshalerError)
	c.Assert(ok, Equals, true)
	c.Assert(e.Err.Error(), Equals, "Unmarshable.")
}

func (s *S) TestNewRepository(c *C) {
	h := testHandler{}
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
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.NewRepository("proj1", []string{"someuser"}, false)
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestGetRepository(c *C) {
	content := `{"name":"repo-name","git_url":"git@test.com:repo-name.git","ssh_url":"git://test.com/repo-name.git"}`
	h := testHandler{content: content}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	r, err := client.GetRepository("repo-name")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/repository/repo-name?:name=repo-name")
	c.Assert(h.method, Equals, "GET")
	c.Assert(r.Name, Equals, "repo-name")
	c.Assert(r.GitURL, Equals, "git@test.com:repo-name.git")
	c.Assert(r.SshURL, Equals, "git://test.com/repo-name.git")
}

func (s *S) TestGetRepositoryOnUnmarshalError(c *C) {
	h := testHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	r, err := client.GetRepository("repo-name")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^Caught error decoding returned json: unexpected end of JSON input$")
	c.Assert(r.Name, Equals, "")
	c.Assert(r.GitURL, Equals, "")
	c.Assert(r.SshURL, Equals, "")
}

func (s *S) TestGetRepositoryOnHTTPError(c *C) {
	content := `null`
	h := errorHandler{content: content}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.GetRepository("repo-name")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^Caught error getting repository metadata: Error performing requested operation\n$")
}

func (s *S) TestNewUser(c *C) {
	h := testHandler{}
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
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.NewUser("someuser", map[string]string{"testkey": "ssh-rsa somekey"})
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRemoveUser(c *C) {
	h := testHandler{}
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
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveUser("someuser")
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRemoveRepository(c *C) {
	h := testHandler{}
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
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveRepository("proj2")
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestAddKey(c *C) {
	h := testHandler{}
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
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.AddKey("proj2", map[string]string{"key": "ssh-rsa keycontent user@host"})
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRemoveKey(c *C) {
	h := testHandler{}
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
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RemoveKey("proj2", "keyname")
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestListKeys(c *C) {
	h := testHandler{content: `{"fookey":"bar keycontent"}`}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	keys, err := client.ListKeys("userx")
	c.Assert(err, IsNil)
	expected := map[string]string{"fookey": "bar keycontent"}
	c.Assert(expected, DeepEquals, keys)
	c.Assert(h.url, Equals, "/user/userx/keys")
	c.Assert(h.method, Equals, "GET")
}

func (s *S) TestListKeysWithError(c *C) {
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.ListKeys("userx")
	c.Assert(err.Error(), Equals, "Error performing requested operation\n")
}

func (s *S) TestGrantAccess(c *C) {
	h := testHandler{}
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
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.GrantAccess([]string{"projectx", "projecty"}, []string{"userx"})
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRevokeAccess(c *C) {
	h := testHandler{}
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
	h := errorHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	err := client.RevokeAccess([]string{"projectx", "projecty"}, []string{"usery"})
	expected := "^Error performing requested operation\n$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestGetDiff(c *C) {
	content := "diff_test"
	h := testHandler{content: content}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	diffOutput, err := client.GetDiff("repo-name", "1b970b076bbb30d708e262b402d4e31910e1dc10", "545b1904af34458704e2aa06ff1aaffad5289f8f")
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/repository/repo-name/diff/commits?:name=repo-name&previous_commit=1b970b076bbb30d708e262b402d4e31910e1dc10&last_commit=545b1904af34458704e2aa06ff1aaffad5289f8f")
	c.Assert(h.method, Equals, "GET")
	c.Assert(diffOutput, Equals, content)
}

func (s *S) TestGetDiffOnHTTPError(c *C) {
	content := `null`
	h := errorHandler{content: content}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.GetDiff("repo-name", "1b970b076bbb30d708e262b402d4e31910e1dc10", "545b1904af34458704e2aa06ff1aaffad5289f8f")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^Caught error getting repository metadata: Error performing requested operation\n$")
}

func (s *S) TestHealthCheck(c *C) {
	content := "test"
	h := testHandler{content: content}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	result, err := client.GetHealthCheck()
	c.Assert(err, IsNil)
	c.Assert(h.url, Equals, "/healthcheck")
	c.Assert(h.method, Equals, "GET")
	c.Assert(string(result), Equals, content)
}

func (s *S) TestHealthCheckOnHTTPError(c *C) {
	content := `null`
	h := errorHandler{content: content}
	ts := httptest.NewServer(&h)
	defer ts.Close()
	client := Client{Endpoint: ts.URL}
	_, err := client.GetHealthCheck()
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^Error performing requested operation\n$")
}
