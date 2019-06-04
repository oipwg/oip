package oip5

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/httpapi"
	"gopkg.in/olivere/elastic.v6"
)

const o5RecordIndexName = "oip5_record"
const o5TemplateIndexName = "oip5_templates"

var o5Router = httpapi.NewSubRoute("/o5")

func init() {
	o5Router.HandleFunc("/record/search", handleRecordSearch).Queries("q", "{query}")
	o5Router.HandleFunc("/template/search", handleTemplateSearch).Queries("q", "{query}")
	o5Router.HandleFunc("/record/get/latest", handleLatestRecord)
	o5Router.HandleFunc("/record/get/{id:[a-f0-9]+}", handleGetRecord)
	o5Router.HandleFunc("/record/mapping/{tmpl:tmpl_[a-fA-F0-9]{8}(?:,tmpl_[a-fA-F0-9]{8})*}", handleGetMapping)
	o5Router.HandleFunc("/template/get/latest", handleLatestTemplate)
	o5Router.HandleFunc("/template/get/{id:[a-f0-9]+}", handleGetTemplate)
}

var (
	o5Indices = []string{o5RecordIndexName}
	o5Fsc     = elastic.NewFetchSourceContext(true).
			Include("record.*", "template.*", "file_descriptor_set", "meta.signed_by", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")
)

func handleRecordSearch(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)
	// log.Info("handleSearchRecord", logger.Attrs{"opts": opts})
	searchQuery, err := url.PathUnescape(opts["query"])
	if err != nil {
		httpapi.RespondJSON(w, 400, map[string]interface{}{
			"error": "unable to decode query",
		})
		return
	}

	query := elastic.NewBoolQuery().Must(
		elastic.NewQueryStringQuery(searchQuery).
			AnalyzeWildcard(false),
		elastic.NewTermQuery("meta.deactivated", false),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		o5Indices,
		query,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid.keyword", Ascending: true},
		},
		o5Fsc,
	)

	httpapi.RespondSearch(w, searchService)
}

func handleTemplateSearch(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	searchQuery, err := url.PathUnescape(opts["query"])
	if err != nil {
		httpapi.RespondJSON(w, 400, map[string]interface{}{
			"error": "unable to decode query",
		})
		return
	}

	query := elastic.NewBoolQuery().Must(
		elastic.NewQueryStringQuery(searchQuery).
			AnalyzeWildcard(false),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		[]string{o5TemplateIndexName},
		query,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		o5Fsc,
	)

	httpapi.RespondSearch(w, searchService)
}

func handleLatestRecord(w http.ResponseWriter, r *http.Request) {

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		o5Indices,
		q,
		[]elastic.SortInfo{{Field: "meta.time", Ascending: false}},
		o5Fsc,
	)

	httpapi.RespondSearch(w, searchService)
}

func handleGetRecord(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		o5Indices,
		q,
		[]elastic.SortInfo{{Field: "meta.time", Ascending: false}},
		o5Fsc,
	)

	httpapi.RespondSearch(w, searchService)
}

func handleGetMapping(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)
	var fields []string

	for _, tmpl := range strings.Split(opts["tmpl"], ",") {
		fields = append(fields, "record.details."+tmpl+".*")
	}

	indexName := datastore.Index(o5RecordIndexName)

	res, err := datastore.Client().
		GetFieldMapping().
		Index(indexName).
		Type("_doc").
		Field(fields...).
		Do(r.Context())

	if err != nil {
		httpapi.RespondESError(w, err)
		return
	}

	fmt.Println(fields)

	m := res[indexName].(map[string]interface{})["mappings"].(map[string]interface{})["_doc"].(map[string]interface{})

	var ret = make(map[string]interface{})

	for key, value := range m {
		ret[key] = value
	}

	httpapi.RespondJSON(w, 200, ret)
}

func handleLatestTemplate(w http.ResponseWriter, r *http.Request) {

	q := elastic.NewExistsQuery("_id")

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		[]string{"oip5_templates"},
		q,
		[]elastic.SortInfo{{Field: "meta.time", Ascending: false}},
		o5Fsc,
	)

	httpapi.RespondSearch(w, searchService)
}

func handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		// elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		[]string{"oip5_templates"},
		q,
		[]elastic.SortInfo{{Field: "meta.time", Ascending: false}},
		o5Fsc,
	)

	httpapi.RespondSearch(w, searchService)
}
