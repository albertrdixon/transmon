package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.google.com/p/go-uuid/uuid"
	"github.com/albertrdixon/gearbox/url"
	"github.com/stretchr/testify/assert"
	"github.com/zenazn/goji/web"
)

func testServer(route, output string) *httptest.Server {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	m := web.New()
	m.Post(route, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, output)
	})
	mux.Handle("/", m)

	return server
}

func TestRequestPort(t *testing.T) {
	is := assert.New(t)
	route := "/port/forward"
	server := testServer("/port/forward", `{"port":1234}`)
	defer server.Close()

	u, er := url.Parse(fmt.Sprintf("%s%s", server.URL, route))
	if !is.NoError(er) {
		t.Log(er.Error())
		t.FailNow()
	}

	t.Logf("URL: %v", u)
	p := &PIA{URL: u, User: "user", Pass: "pass", ClientID: uuid.New()}
	port, er := requestPort("1.2.3.4", p)
	is.NoError(er)
	is.Equal(1234, port)
}
