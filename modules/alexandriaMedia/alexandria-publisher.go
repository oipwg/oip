package alexandriaMedia

import (
	"net/http"

	"github.com/azer/logger"
	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/httpapi"
	"gopkg.in/olivere/elastic.v6"
)

const apIndexName = "alexandria-publisher"

var pubRouter = httpapi.NewSubRoute("/alexandria/publisher")

func init() {
	log.Info("init alexandria-publisher")
	events.SubscribeAsync("modules:oip:alexandriaPublisher", onAlexandriaPublisher)
	datastore.RegisterMapping(apIndexName, "alexandria-publisher.json")
	pubRouter.HandleFunc("/get/latest/", handleLatestPublishers)
	pubRouter.HandleFunc("/get/{address:[A-Za-z0-9]+}", handleGetPublisher)
}

var (
	apIndices = []string{apIndexName}
	apFsc     = elastic.NewFetchSourceContext(true).
			Include("*")
)

func handleLatestPublishers(w http.ResponseWriter, r *http.Request) {
	q := elastic.NewBoolQuery().Must(
		elastic.NewExistsQuery("address"),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		apIndices,
		q,
		[]elastic.SortInfo{{Field: "timestamp", Ascending: false}},
		apFsc,
	)
	httpapi.RespondSearch(w, searchService)
}

func handleGetPublisher(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("address", opts["address"]),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		apIndices,
		q,
		[]elastic.SortInfo{{Field: "timestamp", Ascending: false}},
		apFsc,
	)
	httpapi.RespondSearch(w, searchService)
}

func onAlexandriaPublisher(floData string, tx *datastore.TransactionData) {
	pub := jsoniter.Get([]byte(floData), "alexandria-publisher")
	if pub.LastError() != nil {
		log.Error("invalid json", logger.Attrs{"floData": floData, "txid": tx.Transaction.Txid})
		return
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("alexandria-publisher")).Type("_doc").Doc(pub).Id(tx.Transaction.Txid)
	datastore.AutoBulk.Add(bir)
}
