package httpapi

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/oipwg/oip/datastore"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	rootRouter.HandleFunc("/artifact/get/latest", handleLatest)
	rootRouter.HandleFunc("/artifact/get/latest", handleLatest)
	rootRouter.HandleFunc("/artifact/get/{id:[a-f0-9]+}", handleGet)
	rootRouter.HandleFunc("/artifact/search", handleArtifactSearch).Queries("q", "{query}")
	rootRouter.HandleFunc("/artifact/cardinality", handleCardinality).Queries("f", "{field:[a-zA-Z\\.]+}")
}

var (
	artifactIndices = []string{"oip041", "oip042_artifact"}
	artifactFsc     = elastic.NewFetchSourceContext(true).
			Include("artifact.*", "meta.block_hash", "meta.txid", "meta.originalTxid", "meta.block", "meta.time", "meta.type")
)

func handleArtifactSearch(w http.ResponseWriter, r *http.Request) {
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
			// DefaultField("artifact.info.description").
			AnalyzeWildcard(false),
		elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewTermQuery("meta.blacklist.blacklisted", false),
	)

	searchService := BuildCommonSearchService(
		r.Context(),
		artifactIndices,
		query,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		artifactFsc,
	)

	RespondSearch(w, searchService)
}

func handleLatest(w http.ResponseWriter, r *http.Request) {

	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)

	// if n, ok := opts["nsfw"]; ok {
	// 	nsfw, _ := strconv.ParseBool(n)
	// 	query.Must(elastic.NewTermQuery("", true))
	// 	log.Info("nsfw: %t", nsfw)
	// }

	searchService := BuildCommonSearchService(
		r.Context(),
		artifactIndices,
		query,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		artifactFsc,
	)

	RespondSearch(w, searchService)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
		elastic.NewTermQuery("meta.blacklist.blacklisted", false),
	)

	searchService := BuildCommonSearchService(
		r.Context(),
		artifactIndices,
		query,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		artifactFsc,
	)

	RespondSearch(w, searchService)
}

func handleCardinality(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)

	s, err := datastore.Client().
		Search(datastore.Index("oip042_artifact")).
		Size(0).
		Query(query).
		Aggregation(
			"cardinality",
			elastic.NewCardinalityAggregation().
				Field(opts["field"]),
		).
		Do(context.TODO())

	if err != nil {
		RespondESError(w, err)
		return
	}

	agg, ok := s.Aggregations.Cardinality("cardinality")
	if !ok {
		RespondESError(w, errors.New("cardinality not found"))
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"c": agg.Value,
		"f": opts["field"],
	})
}
