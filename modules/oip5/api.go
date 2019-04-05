package oip5

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/httpapi"
	"gopkg.in/olivere/elastic.v6"
)

const o5RecordIndexName = "oip5_record"

var o5Router = httpapi.NewSubRoute("/o5")

func init() {
	o5Router.HandleFunc("/record/get/latest", handleLatestRecord)
	o5Router.HandleFunc("/record/get/{id:[a-f0-9]+}", handleGetRecord)
	o5Router.HandleFunc("/record/mapping/{tmpl:tmpl_[a-f0-9]{16}(?:,tmpl_[a-f0-9]{16})*}", handleGetMapping)
	o5Router.HandleFunc("/template/get/latest", handleLatestTemplate)
	o5Router.HandleFunc("/record/get/{id:[a-f0-9]+}", handleGetTemplate)
}

var (
	o5Indices = []string{o5RecordIndexName}
	o5Fsc     = elastic.NewFetchSourceContext(true).
			Include("record.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")
)

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
