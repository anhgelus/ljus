package ljus

import (
	"container/list"
	"net/http"
	"strings"
)

// Group containing [Route].
type Group struct {
	pattern     string
	name        string
	routes      *list.List
	middlewares *list.List
}

// Group returns a new [Group] called for the specified pattern.
func NewGroup(pattern string) *Group {
	g := &Group{
		routes:      list.New(),
		middlewares: list.New(),
	}
	_, p, ok := strings.Cut(pattern, " ")
	if !ok {
		p = pattern
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	g.pattern = p
	return g
}

// Add a [Route] to the [Group].
func (g *Group) Add(routes ...Route) *Group {
	for _, r := range routes {
		g.routes.PushBack(r)
	}
	return g
}

func (g *Group) Handler(middlewares *list.List) http.Handler {
	mux := http.NewServeMux()
	for e := g.routes.Front(); e != nil; e = e.Next() {
		route := e.Value.(Route).Clone()
		route.SetName(g.name + " -> " + route.Name())
		cp := copyList(middlewares)
		cp.PushFrontList(g.middlewares)
		mux.Handle(route.Pattern(), route.Handler(cp))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.RequestURI, "/") {
			r.RequestURI = "/" + r.RequestURI
		}
		r.RequestURI = "/" + strings.TrimPrefix(r.RequestURI, g.pattern)
		r.URL.Path, _, _ = strings.Cut(r.RequestURI, "?")
		mux.ServeHTTP(w, r)
	})
}

func (g *Group) Use(middlewares ...Middleware) Route {
	for _, m := range middlewares {
		g.middlewares.PushFront(m)
	}
	return g
}

func (g *Group) SetName(name string) Route {
	g.name = name
	return g
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Pattern() string {
	return g.pattern
}

func (g *Group) Clone() Route {
	cl := *g
	cl.middlewares = copyList(cl.middlewares)
	cl.routes = copyList(cl.routes)
	return &cl
}
