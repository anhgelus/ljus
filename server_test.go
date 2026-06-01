package ljus

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func do(t *testing.T, fn func(string) (*http.Response, error), url string, exp string) {
	resp, err := fn(url)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != exp {
		t.Errorf("invalid response for %s: %s, wanted %s", url, b, exp)
	}
}

func TestServer(t *testing.T) {
	s := New()
	s.Handle(NewRouteFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("foo"))
	}).Name("get foo"))
	s.Handle(NewRouteFunc("POST /foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("FOO"))
	}).Name("post foo"))

	g := NewGroup("/group")
	g.Add(NewRouteFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	}).Name("group root"))
	g.Add(NewRouteFunc("GET /bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	}).Name("get bar"))
	g.Add(NewRouteFunc("POST /bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("BAR"))
	}).Name("post bar"))

	s.Handle(g)

	ts := httptest.NewServer(s.Handler())

	do(t, http.Get, ts.URL+"/foo", "foo")
	do(t, func(s string) (*http.Response, error) {
		return http.Post(s, "", nil)
	}, ts.URL+"/foo", "FOO")

	do(t, http.Get, ts.URL+"/group", "root")
	do(t, http.Get, ts.URL+"/group/bar", "bar")
	do(t, func(s string) (*http.Response, error) {
		return http.Post(s, "", nil)
	}, ts.URL+"/group/bar", "BAR")
}
