# ljus

Simple, idiomatic and lightweight http server.

Install it with:
```shell
go get -u anhgelus.world/ljus
```

## Usage

Quickstart:
```go
s := ljus.New()
s.Handle(ljus.NewRouteFunc("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello :D"))
}).SetName("root"))

l, err := net.Listen("tcp", ":8000")
if err != nil {
	panic(err)
}
defer l.Close()
err = s.Serve(context.TODO(), l)
if err != nil {
	panic(err)
}
```

Use FastCGI instead of HTTP/1.1:
```go
err = s.ServeFastCGI(context.TODO(), l)
if err != nil {
	panic(err)
}
```

Use a unix socket instead of TCP:
```go
l, err := ljus.ListenSocket("/run/foo/bar", os.FileMode(Oo766))
if err != nil {
	panic(err)
}
defer l.Close()
```

Group routes:
```go
g := ljus.NewGroup("/foo")
g.SetName("foo group")
g.Add(ljus.NewRouteFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("URI: /foo/bar"))
}).SetName("bar"))

s.Handle(g)
```

Use a middleware:
```go
s.Use(middleware.Log(slog.Default(), false, false))
```
