package gandalfclient

import (
	. "launchpad.net/gocheck"
	"net/http/httptest"
)

func (s *S) TestNewRepository(c *C) {
	h := TestHandler{content: `Repository "proj1" created successfuly`}
	ts := httptest.NewServer(&h)
	client := Client{Endpoint: ts.URL}
	_, err := client.NewRepository("proj1", []string{"someuser"}, false)
	c.Assert(err, IsNil)
    c.Assert(string(h.body), Equals, `{"name":"proj1","users":["someuser"],"ispublic":false}`)
}
