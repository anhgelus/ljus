package ljus

import (
	"container/list"
	"net"
	"net/http"
	"net/http/fcgi"
)

// StatusWriter is a [http.ResponseWriter] that stores the status code of the response.
type StatusWriter struct {
	http.ResponseWriter
	Code int
}

func (w *StatusWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.Code = statusCode
}

// Handler is a custom [http.HandlerFunc] taking a [StatusWriter] instead of a [http.ResponseWriter].
type Handler func(*StatusWriter, *http.Request)

type Middleware func(Handler, *StatusWriter, *http.Request)

// Server is handling requests.
type Server struct {
	middlewares *list.List
	routes      *list.List
	mux         *http.ServeMux
}

// Serve handles HTTP requests.
//
// See [Server.ServeFastCGI] to handles HTTP requests over FastCGI.
// See [Server.Handler] to get a [http.handler].
func (s *Server) Serve(l net.Listener) error {
	h := s.Handler()
	return http.Serve(l, h)
}

// Serve handles HTTP requests transported with FastCGI.
//
// See [Server.Serve] to handles HTTP requests.
// See [Server.Handler] to get a [http.handler].
func (s *Server) ServeFastCGI(l net.Listener) error {
	h := s.Handler()
	return fcgi.Serve(l, h)
}

// New returns a new [Server] with the default [http.ServeMux].
func New() *Server {
	return NewWithBase(http.NewServeMux())
}

// New returns a new [Server] with a given [http.ServeMux].
func NewWithBase(mux *http.ServeMux) *Server {
	return &Server{
		middlewares: list.New(),
		routes:      list.New(),
		mux:         mux,
	}
}

type key uint

var KeyRouteName key = 0

// Handler returns the [http.Handler] behind the [Server].
//
// See [Server.Serve] and [Server.ServeFastCGI] to serve requests directly.
func (s *Server) Handler() http.Handler {
	for e := s.routes.Front(); e != nil; e = e.Next() {
		route := e.Value.(*Route)
		s.mux.Handle(route.pattern, route.Handler(s.middlewares))
	}
	return s.mux
}

// Use a middleware for every requests.
func (s *Server) Use(middlewares ...Middleware) {
	for _, middle := range middlewares {
		s.middlewares.PushFront(func(next Handler) Handler {
			return func(w *StatusWriter, r *http.Request) {
				middle(next, w, r)
			}
		})
	}
}
