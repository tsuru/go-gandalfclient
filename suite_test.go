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

type TestHandler struct {
	body    []byte
	method  string
	url     string
	content string
	header  http.Header
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.method = r.Method
	h.url = r.URL.String()
	h.body, _ = ioutil.ReadAll(r.Body)
	h.header = r.Header
	w.Write([]byte(h.content))
}

type ErrorHandler struct {
	body    []byte
	method  string
	url     string
	content string
	header  http.Header
}

func (h *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.method = r.Method
	h.url = r.URL.String()
	h.body, _ = ioutil.ReadAll(r.Body)
	h.header = r.Header
	http.Error(w, "Error performing requested operation", http.StatusBadRequest)
}
