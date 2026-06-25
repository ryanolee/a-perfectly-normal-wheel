package middleware

import (
	"net/http"
	"strings"
)

const (
	BasicAuthPath = "/admin"
)

func BasicAuthMiddleware(next http.Handler, username string, password string) (handlerFunc http.HandlerFunc) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, BasicAuthPath) {
			next.ServeHTTP(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
