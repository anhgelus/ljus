package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"anhgelus.world/ljus"
	"anhgelus.world/ljus/middleware"
)

func main() {
	r := ljus.New()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	r.Use(middleware.Log(log, nil, true, true))
	r.Use(middleware.SecurityHeaders("example.org", 15*time.Minute))
	r.Handle(ljus.NewRouteFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello :D")
	}).Name("root"))

	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	err = r.Serve(l)
	if err != nil {
		panic(err)
	}
}
