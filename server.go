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
	handler     Handler
	mux         *http.ServeMux
}

// Serve handles HTTP requests.
//
// See [Server.ServeFastCGI] to handles HTTP requests over FastCGI.
// See [Server.Handler] to get a [http.handler].
func (s *Server) Serve(l net.Listener) error {
	return http.Serve(l, s.Handler())
}

// Serve handles HTTP requests transported with FastCGI.
//
// See [Server.Serve] to handles HTTP requests.
// See [Server.Handler] to get a [http.handler].
func (s *Server) ServeFastCGI(l net.Listener) error {
	return fcgi.Serve(l, s.Handler())
}

// New returns a new [Server] with the default [http.ServeMux].
func New() *Server {
	return NewWithBase(http.NewServeMux())
}

// New returns a new [Server] with a given [http.ServeMux].
func NewWithBase(mux *http.ServeMux) *Server {
	return &Server{
		middlewares: list.New(),
		mux:         mux,
	}
}

// Handler returns the [http.Handler] behind the [Server].
//
// See [Server.Serve] and [Server.ServeFastCGI] to serve requests directly.
func (s *Server) Handler() http.Handler {
	s.handler = func(w *StatusWriter, r *http.Request) {
		s.mux.ServeHTTP(w, r)
	}
	for e := s.middlewares.Back(); e != nil; e = e.Prev() {
		s.handler = e.Value.(func(Handler) Handler)(s.handler)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handler(&StatusWriter{ResponseWriter: w}, r)
	})
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
