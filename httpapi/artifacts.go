package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/gorilla/mux"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	rootRouter.HandleFunc("/artifact/get/latest/{limit:[0-9]+}", handleLatest).Queries("nsfw", "{nsfw}")
	rootRouter.HandleFunc("/artifact/get/latest/{limit:[0-9]+}", handleLatest)
	rootRouter.HandleFunc("/artifact/get/{id:[a-f0-9]+}", handleGet)
	rootRouter.HandleFunc("/artifact/search", handleArtifactSearch).Queries("q", "{query}", "limit", "{limit:[0-9]+}")
	rootRouter.HandleFunc("/artifact/search", handleArtifactSearch).Queries("q", "{query}")
}

func handleArtifactSearch(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	lim, _ := opts["limit"]
	size, _ := strconv.ParseInt(lim, 10, 0)
	if size <= 0 || size > 1000 {
		size = -1
	}

	searchQuery, err := url.PathUnescape(opts["query"])
	if err != nil {
		RespondJSON(w, 400, map[string]interface{}{
			"error": "unable to decode query",
		})
		return
	}

	q := elastic.NewBoolQuery().Must(
		elastic.NewQueryStringQuery(searchQuery).
			// DefaultField("artifact.info.description").
			AnalyzeWildcard(false),
		elastic.NewTermQuery("meta.deactivated", false),
	)

	log.Info(searchQuery)
	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")

	results, err := datastore.Client().
		Search(datastore.Index("oip041"), datastore.Index("oip042_artifact")).
		Type("_doc").
		Query(q).
		Size(int(size)).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(results.Hits.Hits),
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}

func handleLatest(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	size, _ := strconv.ParseInt(opts["limit"], 10, 0)
	if size <= 0 || size > 1000 {
		size = -1
	}

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)

	// if n, ok := opts["nsfw"]; ok {
	// 	nsfw, _ := strconv.ParseBool(n)
	// 	q.Must(elastic.NewTermQuery("", true))
	// 	log.Info("nsfw: %t", nsfw)
	// }

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")

	results, err := datastore.Client().
		Search(datastore.Index("oip041"), datastore.Index("oip042_artifact")). // "alexandria-media",
		Type("_doc").
		Query(q).
		Size(int(size)).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(results.Hits.Hits),
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")

	results, err := datastore.Client().
		Search(datastore.Index("oip041"), datastore.Index("oip042_artifact"), datastore.Index("alexandria-media")).
		Type("_doc").
		Query(q).
		Size(1).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}
