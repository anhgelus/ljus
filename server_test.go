package ljus_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "anhgelus.world/ljus"
	"anhgelus.world/ljus/middleware"
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

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	s.Use(middleware.Log(log, true, true))

	s.Handle(NewRouteFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("foo"))
	}).SetName("get foo"))
	s.Handle(NewRouteFunc("POST /foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("FOO"))
	}).SetName("post foo"))

	g := NewGroup("/group")
	g.Add(NewRouteFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	}).SetName("group root"))
	g.Add(NewRouteFunc("GET /bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	}).SetName("get bar"))
	g.Add(NewRouteFunc("POST /bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("BAR"))
	}).SetName("post bar"))

	s.Handle(g)

	ts := httptest.NewServer(s.Handler(t.Context()))

	do(t, http.Get, ts.URL+"/foo", "foo")
	do(t, func(s string) (*http.Response, error) {
		return http.Post(s, "", nil)
	}, ts.URL+"/foo", "FOO")

	do(t, http.Get, ts.URL+"/group/", "root")
	do(t, http.Get, ts.URL+"/group/bar", "bar")
	do(t, func(s string) (*http.Response, error) {
		return http.Post(s, "", nil)
	}, ts.URL+"/group/bar", "BAR")

	resp, err := http.Get(ts.URL + "/empty")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("invalid status code: %d, wanted %d", resp.StatusCode, 404)
	}
}
