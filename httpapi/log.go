package httpapi

import (
	"net/http"

	"github.com/azer/logger"
)

var log = logger.New("httpApi")

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := log.Timer()
		next.ServeHTTP(w, r)
		t.End("req", logger.Attrs{
			"url":           r.URL,
			"httpMethod":    r.Method,
			"remoteAddr":    r.RemoteAddr,
			"contentLength": r.ContentLength,
			"userAgent":     r.UserAgent(),
		})
	})
}
