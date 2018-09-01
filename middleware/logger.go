package middleware

import (
	"net/http"

	"github.com/godiscourse/godiscourse/durable"
	"github.com/godiscourse/godiscourse/session"
)

func Logger(handler http.Handler, logger *durable.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := session.WithLogger(r.Context(), logger)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
