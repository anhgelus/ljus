package ljus

import (
	"container/list"
	"context"
	"net/http"
)

// Route represents an HTTP route.
type Route struct {
	pattern     string
	handler     Handler
	name        string
	middlewares *list.List
}

// Handle returns a new [Route] called for the specified pattern.
//
// See [Server.HandleFunc].
func (s *Server) Handle(pattern string, handler http.Handler) *Route {
	r := &Route{
		pattern:     pattern,
		handler:     func(w *StatusWriter, r *http.Request) { handler.ServeHTTP(w, r) },
		name:        "",
		middlewares: list.New(),
	}
	s.routes.PushFront(r)
	return r
}

// HandleFunc returns a new [Route] called for the specified pattern.
//
// See [Server.Handle].
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) *Route {
	return s.Handle(pattern, http.HandlerFunc(handler))
}

// With adds a [Middleware] to the [Route].
func (r *Route) With(middlewares ...Middleware) *Route {
	for _, m := range middlewares {
		r.middlewares.PushBack(m)
	}
	return r
}

// Name the [Route].
func (r *Route) Name(name string) *Route {
	r.name = name
	return r
}

func (r *Route) Handler(middlewares *list.List) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if r.name != "" {
			ctx = context.WithValue(ctx, KeyRouteName, r.name)
		}
		handler := r.handler
		for e := middlewares.Back(); e != nil; e = e.Prev() {
			handler = e.Value.(func(Handler) Handler)(handler)
		}
		for e := r.middlewares.Back(); e != nil; e = e.Prev() {
			handler = e.Value.(func(Handler) Handler)(handler)
		}
		handler(&StatusWriter{ResponseWriter: w}, req.WithContext(ctx))
	})
}
