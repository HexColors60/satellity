package main

import (
	"fmt"
	"net/http"

	"github.com/dimfeld/httptreemux"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/go-pg/pg"
	"github.com/godiscourse/godiscourse/config"
	"github.com/godiscourse/godiscourse/controllers"
	"github.com/godiscourse/godiscourse/durable"
	"github.com/godiscourse/godiscourse/middleware"
	"github.com/gorilla/handlers"
	"github.com/unrolled/render"
)

func startHttp(db *pg.DB) error {
	router := httptreemux.New()
	controllers.RegisterHanders(router)
	controllers.RegisterRoutes(router)

	handler := middleware.Authenticate(router)
	handler = middleware.Constraint(handler)
	handler = middleware.Context(handler, db, render.New())
	handler = middleware.State(handler)
	handler = middleware.Logger(handler, durable.NewLogger())
	handler = handlers.ProxyHeaders(handler)

	return gracehttp.Serve(&http.Server{Addr: fmt.Sprintf(":%d", config.HTTPListenPort), Handler: handler})
}
