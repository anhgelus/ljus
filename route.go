package ljus

import "net/http"

// Route represents an HTTP route.
type Route struct {
	pattern string
	handler Handler
}

// Handle returns a new [Route] called for the specified pattern.
//
// See [Server.HandleFunc].
func (s *Server) Handle(pattern string, handler http.Handler) *Route {
	r := &Route{pattern, func(w *StatusWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}}
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
func (r *Route) With(middleware Middleware) *Route {
	r.handler = func(w *StatusWriter, req *http.Request) {
		middleware(r.handler, w, req)
	}
	return r
}
