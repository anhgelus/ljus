package ljus

import (
	"container/list"
	"context"
	"errors"
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
	w.Code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Middleware is a [Handler] taking the next [Handler] to call.
type Middleware = func(next http.Handler, w http.ResponseWriter, r *http.Request)

// Server is handling requests.
type Server struct {
	middlewares *list.List
	routes      *list.List
}

// New returns a new [Server] with the default [http.ServeMux].
func New() *Server {
	return &Server{
		middlewares: list.New(),
		routes:      list.New(),
	}
}

func (s *Server) serve(ctx context.Context, l net.Listener, fn func(net.Listener, http.Handler) error) error {
	errc := make(chan error, 1)
	go func() {
		errc <- fn(l, s.Handler())
	}()
	select {
	case <-ctx.Done():
		err := l.Close()
		cause := context.Cause(ctx)
		if err != nil {
			cause = errors.Join(cause, err)
		}
		return cause
	case err := <-errc:
		return err
	}
}

// Serve handles HTTP requests.
//
// See [Server.ServeFastCGI] to handles HTTP requests over FastCGI.
// See [Server.Handler] to get a [http.handler].
//
// See [SocketListener] if you want to use a unix socket.
func (s *Server) Serve(ctx context.Context, l net.Listener) error {
	return s.serve(ctx, l, http.Serve)
}

// Serve handles HTTP requests transported with FastCGI.
//
// See [Server.Serve] to handles HTTP requests.
// See [Server.Handler] to get a [http.handler].
//
// See [SocketListener] if you want to use a unix socket.
func (s *Server) ServeFastCGI(ctx context.Context, l net.Listener) error {
	return s.serve(ctx, l, fcgi.Serve)
}

type key uint

const (
	KeyRouteName key = iota
	KeyCancelContext
	KeyRealURI
)

// Handler returns the [http.Handler] behind the [Server].
//
// See [Server.Serve] and [Server.ServeFastCGI] to serve requests directly.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	for e := s.routes.Front(); e != nil; e = e.Next() {
		route := e.Value.(Route).Clone()
		mux.Handle(route.Pattern(), route.Handler(s.middlewares))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancelCause(context.Background())
		ctx = context.WithValue(ctx, KeyCancelContext, cancel)
		ctx = context.WithValue(ctx, KeyRealURI, r.RequestURI)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Use a [Middleware] for every requests.
func (s *Server) Use(middlewares ...Middleware) {
	for _, m := range middlewares {
		s.middlewares.PushFront(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { m(next, w, r) })
		})
	}
}

// Handle a [Route].
//
// See [NewRoute] and [NewGroup].
func (s *Server) Handle(routes ...Route) *Server {
	for _, m := range routes {
		s.routes.PushFront(m)
	}
	return s
}
