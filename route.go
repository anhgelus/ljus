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
	// Use a [Middleware] with the [Route].
	Use(middlewares ...Middleware) Route
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
	handler     http.Handler
	name        string
	middlewares *list.List
}

// NewRoute returns a new [SimpleRoute] called for the specified pattern.
//
// See [Server.HandleFunc].
func NewRoute(pattern string, handler http.Handler) *SimpleRoute {
	return &SimpleRoute{
		pattern:     pattern,
		handler:     handler,
		name:        "",
		middlewares: list.New(),
	}
}

// HandleFunc returns a new [SimpleRoute] called for the specified pattern.
//
// See [Server.Handle].
func NewRouteFunc(pattern string, handler http.HandlerFunc) *SimpleRoute {
	return NewRoute(pattern, handler)
}

func (r *SimpleRoute) Use(middlewares ...Middleware) Route {
	r.middlewares = copyList(r.middlewares)
	for _, m := range middlewares {
		r.middlewares.PushBack(m)
	}
	return r
}

func (r *SimpleRoute) SetName(name string) Route {
	r.name = name
	return r
}

func (r *SimpleRoute) Name() string {
	return r.name
}

func (r *SimpleRoute) Handler(middlewares *list.List) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if r.name != "" {
			ctx = context.WithValue(ctx, KeyRouteName, r.name)
		}
		handler := generateHandler(r.handler, r.middlewares, middlewares)
		switch w.(type) {
		case *StatusWriter:
			handler.ServeHTTP(w, req.WithContext(ctx))
		default:
			handler.ServeHTTP(&StatusWriter{ResponseWriter: w}, req.WithContext(ctx))
		}
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

func generateHandler(base http.Handler, given *list.List, internal *list.List) http.Handler {
	handler := base
	for e := internal.Front(); e != nil; e = e.Next() {
		handler = e.Value.(func(http.Handler) http.Handler)(handler)
	}
	for e := given.Front(); e != nil; e = e.Next() {
		handler = e.Value.(func(http.Handler) http.Handler)(handler)
	}
	return handler
}
