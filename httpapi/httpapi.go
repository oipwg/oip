package httpapi

import (
	"net/http"
	"time"

	"github.com/azer/logger"
	"github.com/bitspill/oip/version"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	json "github.com/json-iterator/go"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

var rootRouter = mux.NewRouter()
var daemonRoutes = NewSubRoute("/daemon")

var (
	apiStartup time.Time
)

func init() {
	rootRouter.Use(logRequests)
	rootRouter.NotFoundHandler = http.HandlerFunc(handle404)

	daemonRoutes.HandleFunc("/version", handleVersion)
}

func Serve() {
	apiStartup = time.Now()
	listen := viper.GetString("oip.api.listen")
	err := http.ListenAndServe(listen, cors.Default().Handler(rootRouter))
	if err != nil {
		log.Error("Error serving http api", logger.Attrs{"err": err, "listen": listen})
	}
}

func NewSubRoute(prefix string) *mux.Router {
	return rootRouter.PathPrefix(prefix).Subrouter()
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
		"Started":       apiStartup.Format(time.RFC1123Z),
		"Uptime":        time.Since(apiStartup).String(),
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
