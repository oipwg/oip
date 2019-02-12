package httpapi

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/azer/logger"
	"github.com/gorilla/mux"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	rootRouter.HandleFunc("/floData/get/{id:[a-f0-9]+}", handleGetFloData)
	rootRouter.HandleFunc("/floData/latest", handleFloDataLatest)
	rootRouter.HandleFunc("/floData/search", handleFloDataSearch).Queries("q", "{query}")
}

var (
	txFsc = elastic.NewFetchSourceContext(true).
		Include("tx.floData", "tx.txid", "tx.time", "tx.blockhash", "tx.size", "is_coinbase")
)

func handleGetFloData(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)
	log.Info("handleGetFloData", logger.Attrs{"opts": opts})

	q := elastic.NewBoolQuery().Must(
		elastic.NewPrefixQuery("tx.txid", opts["id"]),
	)

	searchService := BuildCommonSearchService(
		r.Context(),
		[]string{"transactions"},
		q,
		[]elastic.SortInfo{{Field: "tx.txid", Ascending: false}},
		txFsc,
	)
	RespondSearch(w, searchService)
}

func handleFloDataSearch(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	searchQuery, err := url.PathUnescape(opts["query"])
	if err != nil {
		RespondJSON(w, 400, map[string]interface{}{
			"error": "unable to decode query",
		})
		return
	}

	query := elastic.NewBoolQuery().Must(
		elastic.NewQueryStringQuery(searchQuery).
			DefaultField("tx.floData").
			AnalyzeWildcard(false),
	)

	searchService := BuildCommonSearchService(
		r.Context(),
		[]string{"transactions"},
		query,
		[]elastic.SortInfo{
			{Ascending: false, Field: "tx.time"},
			{Field: "tx.txid", Ascending: true},
		},
		txFsc,
	)

	RespondSearch(w, searchService)
}

func handleFloDataLatest(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	query := elastic.NewBoolQuery().Must(
		elastic.NewExistsQuery("tx.floData"),
	)

	if c, ok := opts["coinbase"]; ok {
		coinbase, _ := strconv.ParseBool(c)
		if !coinbase {
			query.Must(elastic.NewTermQuery("is_coinbase", coinbase))
		}
	}

	ctx := r.Context()
	searchService := BuildCommonSearchService(
		ctx,
		[]string{"transactions"},
		query,
		[]elastic.SortInfo{{Ascending: false, Field: "tx.time"}, {Ascending: true, Field: "tx.txid"}},
		txFsc,
	)

	RespondSearch(w, searchService)
}
