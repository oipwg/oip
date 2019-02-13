package httpapi

import (
	"net/http"
	"time"

	"github.com/azer/logger"
	"github.com/bitspill/oip/version"
)

func handleVersion(w http.ResponseWriter, _ *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]string{
		"BuiltBy":       version.BuiltBy,
		"BuildDate":     version.BuildDate,
		"GoVersion":     version.GoVersion,
		"GitCommitHash": version.GitCommitHash,
		"Started":       apiStartup.Format(time.RFC1123Z),
		"Uptime":        time.Since(apiStartup).String(),
	})
}

func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("404 not found"))
	log.Info("404", logger.Attrs{
		"url":           r.URL,
		"httpMethod":    r.Method,
		"remoteAddr":    r.RemoteAddr,
		"contentLength": r.ContentLength,
		"userAgent":     r.UserAgent(),
	})
}
