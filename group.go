package ljus

import (
	"container/list"
	"net/http"
	"slices"
	"strings"
)

type Group struct {
	inner *SimpleRoute
	mux   *http.ServeMux
	sub   *list.List
}

// Group returns a new [Group] called for the specified pattern.
func NewGroup(pattern string) *Group {
	g := &Group{
		sub: list.New(),
		mux: http.NewServeMux(),
	}
	v, p, _ := strings.Cut(pattern, " ")
	if !slices.Contains([]string{
		http.MethodConnect, http.MethodDelete, http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPatch,
		http.MethodPost, http.MethodPut, http.MethodTrace,
	}, v) {
		p = pattern
	}
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}
	g.inner = NewRouteFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.RequestURI, "/") {
			r.RequestURI = "/" + r.RequestURI
		}
		r.RequestURI = strings.TrimPrefix(r.RequestURI, p)
		r.URL.Path, _, _ = strings.Cut(r.RequestURI, "?")
		g.mux.ServeHTTP(w, r)
	})
	return g
}

// Add [Route] to the [Group].
func (g *Group) Add(routes ...Route) *Group {
	for _, r := range routes {
		g.sub.PushBack(r)
	}
	return g
}

func (g *Group) Handler(middlewares *list.List) http.Handler {
	for e := g.sub.Front(); e != nil; e = e.Next() {
		route := e.Value.(Route)
		g.mux.Handle(route.Pattern(), route.Handler(list.New()))
	}
	return g.inner.Handler(middlewares)
}

func (g *Group) With(middlewares ...Middleware) Route {
	return g.inner.With(middlewares...)
}

func (g *Group) Name(name string) Route {
	return g.inner.Name(name)
}

func (g *Group) Pattern() string {
	return g.inner.pattern
}
