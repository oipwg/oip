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
	router.HandleFunc("/floData/search", handleFloDataSearch).Queries("q", "{query}", "limit", "{limit:[0-9]+}")
	router.HandleFunc("/floData/search", handleFloDataSearch).Queries("q", "{query}")
}

func handleFloDataSearch(w http.ResponseWriter, r *http.Request) {
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
			DefaultField("tx.floData").
			AnalyzeWildcard(false),
	)

	log.Info(searchQuery)
	fsc := elastic.NewFetchSourceContext(true).
		Include("tx.floData", "tx.txid", "tx.time", "tx.blockhash", "tx.size")

	results, err := datastore.Client().
		Search(datastore.Index("transactions")).
		Type("_doc").
		Query(q).
		Size(int(size)).
		Sort("tx.time", false).
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
