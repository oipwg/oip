package oip042

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/oipwg/oip/httpapi"
	"gopkg.in/olivere/elastic.v6"
)

var artRouter = httpapi.NewSubRoute("/oip042/artifact")
var recordRouter = httpapi.NewSubRoute("/oip042/record")

func init() {
	artRouter.HandleFunc("/get/latest", handleLatest).Queries("nsfw", "{nsfw}")
	artRouter.HandleFunc("/get/latest", handleLatest)
	recordRouter.HandleFunc("/get/{editRecordTxid}", handleGet)
	recordRouter.HandleFunc("/get/{originalTxid}/latest", handleGetLatestEdit)
	recordRouter.HandleFunc("/get/{originalTxid}/version/{editRecordTxid}", handleGets)
}

var (
	o42ArtifactFsc = elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")
)

func handleLatest(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)

	if n, ok := opts["nsfw"]; ok {
		nsfw, _ := strconv.ParseBool(n)
		if !nsfw {
			q.MustNot(elastic.NewTermQuery("artifact.info.nsfw", true))
		}
		log.Info("nsfw: %t", nsfw)
	}

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		[]string{oip042ArtifactIndex},
		q,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		o42ArtifactFsc,
	)
	httpapi.RespondSearch(w, searchService)
}

/**
This method will return the version of the record requested with all edits applied. So, the most recent version of the record.
*/
func handleGet(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	editRecordTxid := vars["editRecordTxid"]
	var opts = mux.Vars(request)

	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewTermQuery("meta.txid", editRecordTxid),
	)

	if n, ok := opts["nsfw"]; ok {
		nsfw, _ := strconv.ParseBool(n)
		if !nsfw {
			query.MustNot(elastic.NewTermQuery("artifact.info.nsfw", true))
		}
		log.Info("nsfw: %t", nsfw)
	}

	searchService := httpapi.BuildCommonSearchService(
		request.Context(),
		[]string{oip042ArtifactIndex},
		query,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		o42ArtifactFsc,
	)
	httpapi.RespondSearch(response, searchService)
}

/**
This method will return the version of the record requested with all edits applied. So, the most recent version of the record.
*/
func handleGetLatestEdit(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	originalTxid := vars["originalTxid"]
	var opts = mux.Vars(request)

	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewTermQuery("meta.originalTxid", originalTxid),
		elastic.NewTermQuery("meta.latest", true),
	)

	if n, ok := opts["nsfw"]; ok {
		nsfw, _ := strconv.ParseBool(n)
		if !nsfw {
			query.MustNot(elastic.NewTermQuery("artifact.info.nsfw", true))
		}
		log.Info("nsfw: %t", nsfw)
	}

	searchService := httpapi.BuildCommonSearchService(
		request.Context(),
		[]string{oip042ArtifactIndex},
		query,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		o42ArtifactFsc,
	)
	httpapi.RespondSearch(response, searchService)
}

/**
This method will return the record requested with the version requested. The version will be specified using the transaction ID.
*/
func handleGets(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	originalTxid := vars["originalTxid"]
	editRecordTxid := vars["editRecordTxid"]
	var opts = mux.Vars(request)

	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewTermQuery("meta.originalTxid", originalTxid),
		elastic.NewTermQuery("meta.txid", editRecordTxid),
	)

	if n, ok := opts["nsfw"]; ok {
		nsfw, _ := strconv.ParseBool(n)
		if !nsfw {
			query.MustNot(elastic.NewTermQuery("artifact.info.nsfw", true))
		}
		log.Info("nsfw: %t", nsfw)
	}

	searchService := httpapi.BuildCommonSearchService(
		request.Context(),
		[]string{oip042ArtifactIndex},
		query,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		o42ArtifactFsc,
	)
	httpapi.RespondSearch(response, searchService)
}
