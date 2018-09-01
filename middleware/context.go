package middleware

import (
	"net/http"

	"github.com/go-pg/pg"
	"github.com/godiscourse/godiscourse/session"
	"github.com/unrolled/render"
)

func Context(handler http.Handler, db *pg.DB, r *render.Render) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := session.WithDatabase(req.Context(), db)
		ctx = session.WithRender(ctx, r)
		handler.ServeHTTP(w, req.WithContext(ctx))
	})
}
