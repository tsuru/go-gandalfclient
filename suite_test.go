// Copyright 2013 go-gandalfclient authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gandalf

import (
	"io/ioutil"
	. "launchpad.net/gocheck"
	"net/http"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

type testHandler struct {
	body    []byte
	method  string
	url     string
	content string
	header  http.Header
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.method = r.Method
	h.url = r.URL.String()
	h.body, _ = ioutil.ReadAll(r.Body)
	h.header = r.Header
	w.Write([]byte(h.content))
}

type errorHandler struct {
	body    []byte
	method  string
	url     string
	content string
	header  http.Header
}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.method = r.Method
	h.url = r.URL.String()
	h.body, _ = ioutil.ReadAll(r.Body)
	h.header = r.Header
	http.Error(w, "Error performing requested operation", http.StatusBadRequest)
}
