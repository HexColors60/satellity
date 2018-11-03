package controllers

import (
	"net/http"

	"github.com/dimfeld/httptreemux"
	"github.com/godiscourse/godiscourse/controllers/admin"
	"github.com/godiscourse/godiscourse/views"
)

func RegisterRoutes(router *httptreemux.TreeMux) {
	router.GET("/_hc", health)

	registerUser(router)
	admin.RegisterAdminRoutes(router)
}

func health(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	views.RenderBlankResponse(w, r)
}
