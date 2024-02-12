package middleware

import (
	"net/http"

	httpresp "github.com/apoldev/go-http/internal/app/lib/http-resp"
)

type limiter interface {
	Take() bool
	Release()
}

func LimitMiddleware(l limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Take() {
			httpresp.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		defer l.Release()
		next.ServeHTTP(w, r)
	})
}
