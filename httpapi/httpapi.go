package httpapi

import (
	"net/http"

	"github.com/azer/logger"
	"github.com/bitspill/oip/config"
	"github.com/bitspill/oip/version"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	json "github.com/json-iterator/go"
)

var router = mux.NewRouter()

func init() {
	router.Use(logRequests)
	router.NotFoundHandler = http.HandlerFunc(handle404)

	router.HandleFunc("/version", handleVersion)
}

func Serve() {
	listen := config.Get("api.listen").String("127.0.0.1:1606")
	http.ListenAndServe(listen, router)
}

func NewSubRoute(prefix string) *mux.Router {
	return router.PathPrefix(prefix).Subrouter()
}

func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	b, err := json.Marshal(payload)
	if err != nil {
		log.Error("Unable to marshal response payload", logger.Attrs{"err": err, "payload": spew.Sdump(payload)})
		w.WriteHeader(500)
		w.Write([]byte("Internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(b)
}

func handleVersion(w http.ResponseWriter, _ *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]string{
		"BuiltBy":       version.BuiltBy,
		"BuildDate":     version.BuildDate,
		"GoVersion":     version.GoVersion,
		"GitCommitHash": version.GitCommitHash,
	})
}

func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 not found"))
	log.Info("404", logger.Attrs{
		"url":           r.URL,
		"httpMethod":    r.Method,
		"remoteAddr":    r.RemoteAddr,
		"contentLength": r.ContentLength,
		"userAgent":     r.UserAgent(),
	})
}

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
