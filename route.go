package ljus

import (
	"container/list"
	"context"
	"net/http"
)

// Route represents a general HTTP route.
// It can be used concurrently.
//
// See [SimpleRoute].
// See [Group].
type Route interface {
	// Handler returns the [http.Handler] used.
	Handler(middlewares *list.List) http.Handler
	// SetName of the [Route].
	SetName(name string) Route
	// Name of the [Route].
	Name() string
	// With add a [Middleware] to the [Route].
	With(middlewares ...Middleware) Route
	// Pattern returns the [http.ServeMux] pattern.
	Pattern() string
	// Clone the [Route].
	Clone() Route
}

func copyList(old *list.List) *list.List {
	l := list.New()
	for e := old.Back(); e != nil; e = e.Prev() {
		l.PushFront(e.Value)
	}
	return l
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
	return &SimpleRoute{
		pattern:     pattern,
		handler:     func(w *StatusWriter, r *http.Request) { handler.ServeHTTP(w, r) },
		name:        "",
		middlewares: list.New(),
	}
}

// HandleFunc returns a new [SimpleRoute] called for the specified pattern.
//
// See [Server.Handle].
func NewRouteFunc(pattern string, handler http.HandlerFunc) *SimpleRoute {
	return NewRoute(pattern, http.HandlerFunc(handler))
}

// With adds a [Middleware] to the [SimpleRoute].
func (r *SimpleRoute) With(middlewares ...Middleware) Route {
	r.middlewares = copyList(r.middlewares)
	for _, m := range middlewares {
		r.middlewares.PushBack(m)
	}
	return r
}

// SetName the [SimpleRoute].
func (r *SimpleRoute) SetName(name string) Route {
	r.name = name
	return r
}

func (r *SimpleRoute) Name() string {
	return r.name
}

// Handler returns the [http.Handler] of the [SimpleRoute].
func (r *SimpleRoute) Handler(middlewares *list.List) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if r.name != "" {
			ctx = context.WithValue(ctx, KeyRouteName, r.name)
		}
		handler := generateHandler(r.handler, r.middlewares, middlewares)
		handler(&StatusWriter{ResponseWriter: w}, req.WithContext(ctx))
	})
}

func (r *SimpleRoute) Pattern() string {
	return r.pattern
}

func (r *SimpleRoute) Clone() Route {
	cl := *r
	cl.middlewares = copyList(cl.middlewares)
	return &cl
}

func generateHandler(base Handler, given *list.List, internal *list.List) Handler {
	handler := base
	for e := internal.Back(); e != nil; e = e.Prev() {
		handler = e.Value.(func(Handler) Handler)(handler)
	}
	for e := given.Back(); e != nil; e = e.Prev() {
		handler = e.Value.(func(Handler) Handler)(handler)
	}
	return handler
}
