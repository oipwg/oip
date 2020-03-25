package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/azer/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	json "github.com/json-iterator/go"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"gopkg.in/olivere/elastic.v6"
)

var rootRouter = mux.NewRouter().PathPrefix("/oip").Subrouter()
var daemonRoutes = NewSubRoute("/daemon")

var (
	apiStartup time.Time
)

func init() {
	rootRouter.Use(logRequests)
	rootRouter.Use(commonParameterParser)
	rootRouter.NotFoundHandler = http.HandlerFunc(handle404)

	daemonRoutes.HandleFunc("/version", handleVersion)
}

func Serve() {
	apiStartup = time.Now()
	listen := viper.GetString("oip.api.listen")
	err := http.ListenAndServe(listen, handlers.CompressHandler(cors.Default().Handler(rootRouter)))
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
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(500)
		n, err := w.Write([]byte("Internal server error"))
		if err != nil {
			log.Error("Unable to write json response", logger.Attrs{"n": n, "err": err, "payload": payload, "code": code})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	n, err := w.Write(b)
	if err != nil {
		log.Error("Unable to write json response", logger.Attrs{"n": n, "err": err, "payload": payload, "code": code})
	}
}

func RespondESError(w http.ResponseWriter, err error) {
	if elasticErr, ok := err.(*elastic.Error); ok {
		if elasticErr.Status == 400 {
			RespondJSON(w, 400, map[string]interface{}{
				"error": "invalid search request",
			})
			return
		}
	}
	RespondJSON(w, 500, map[string]interface{}{
		"error": "unable to execute search",
	})
}

func RespondSearch(w http.ResponseWriter, searchService *elastic.SearchService) {
	results, err := searchService.Do(context.TODO())
	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err, "results": results})
		RespondESError(w, err)
		return
	}
	sources, nextAfter := ExtractSources(results)
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(results.Hits.Hits),
		"total":   results.Hits.TotalHits,
		"results": sources,
		"next":    nextAfter,
	})
}
