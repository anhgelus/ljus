package ljus

import (
	"container/list"
	"context"
	"net/http"
)

// Route represents a general HTTP route.
//
// See [SimpleRoute].
// See [Group].
type Route interface {
	// Handler returns the [http.Handler] used.
	Handler(middlewares *list.List) http.Handler
	// Sets the Name of the [Route].
	Name(name string) Route
	// With add a [Middleware] to the [Route].
	With(middlewares ...Middleware) Route
	// Pattern returns the [http.ServeMux] pattern.
	Pattern() string
}

// SimpleRoute is just a [Route].
//
// See [Group] to group multiple [Route].
type SimpleRoute struct {
	pattern     string
	handler     Handler
	name        string
	middlewares *list.List
}

// NewRoute returns a new [SimpleRoute] called for the specified pattern.
//
// See [Server.HandleFunc].
func NewRoute(pattern string, handler http.Handler) *SimpleRoute {
	r := &SimpleRoute{
		pattern:     pattern,
		handler:     func(w *StatusWriter, r *http.Request) { handler.ServeHTTP(w, r) },
		name:        "",
		middlewares: list.New(),
	}
	return r
}

// HandleFunc returns a new [SimpleRoute] called for the specified pattern.
//
// See [Server.Handle].
func NewRouteFunc(pattern string, handler http.HandlerFunc) *SimpleRoute {
	return NewRoute(pattern, http.HandlerFunc(handler))
}

// With adds a [Middleware] to the [SimpleRoute].
func (r *SimpleRoute) With(middlewares ...Middleware) Route {
	for _, m := range middlewares {
		r.middlewares.PushBack(m)
	}
	return r
}

// Name the [SimpleRoute].
func (r *SimpleRoute) Name(name string) Route {
	r.name = name
	return r
}

// Handler returns the [http.Handler] of the [SimpleRoute].
func (r *SimpleRoute) Handler(middlewares *list.List) http.Handler {
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

func (r *SimpleRoute) Pattern() string {
	return r.pattern
}
