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
	w.ResponseWriter.WriteHeader(statusCode)
	w.Code = statusCode
}

// Handler is a custom [http.HandlerFunc] taking a [StatusWriter] instead of a [http.ResponseWriter].
type Handler func(*StatusWriter, *http.Request)

type Middleware func(Handler, *StatusWriter, *http.Request)

// Server is handling requests.
type Server struct {
	// NotFoundHandler is called when no other handler is available.
	NotFoundHandler http.Handler
	middlewares     *list.List
	routes          *list.List
}

// New returns a new [Server] with the default [http.ServeMux].
func New() *Server {
	return &Server{
		NotFoundHandler: http.NotFoundHandler(),
		middlewares:     list.New(),
		routes:          list.New(),
	}
}

func (s *Server) serve(ctx context.Context, l net.Listener, fn func(net.Listener, http.Handler) error) error {
	errc := make(chan error, 1)
	go func() {
		errc <- fn(l, s.Handler(ctx))
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
func (s *Server) Serve(ctx context.Context, l net.Listener) error {
	return s.serve(ctx, l, http.Serve)
}

// Serve handles HTTP requests transported with FastCGI.
//
// See [Server.Serve] to handles HTTP requests.
// See [Server.Handler] to get a [http.handler].
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
func (s *Server) Handler(ctx context.Context) http.Handler {
	mux := http.NewServeMux()
	var homeHandler http.Handler
	for e := s.routes.Front(); e != nil; e = e.Next() {
		route := e.Value.(Route).Clone()
		if route.Pattern() != "/" {
			mux.Handle(route.Pattern(), route.Handler(s.middlewares))
		} else {
			homeHandler = route.Handler(s.middlewares)
		}
	}
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != "/" || homeHandler == nil {
			s.NotFoundHandler.ServeHTTP(w, r)
		} else {
			homeHandler.ServeHTTP(w, r)
		}
	}))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cancel context.CancelCauseFunc
		ctx, cancel = context.WithCancelCause(ctx)
		ctx = context.WithValue(ctx, KeyCancelContext, cancel)

		ctx = context.WithValue(ctx, KeyRealURI, r.RequestURI)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})

}

// Use a [Middleware] for every requests.
func (s *Server) Use(middlewares ...Middleware) {
	for _, m := range middlewares {
		s.middlewares.PushFront(func(next Handler) Handler {
			return func(w *StatusWriter, r *http.Request) { m(next, w, r) }
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
