package alexandriaMedia

import (
	"context"
	"net/http"
	"strconv"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/httpapi"
	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"gopkg.in/olivere/elastic.v6"
)

const apIndexName = "alexandria-publisher"

var pubRouter = httpapi.NewSubRoute("/alexandria/publisher")

func init() {
	log.Info("init alexandria-publisher")
	events.Bus.SubscribeAsync("modules:oip:alexandriaPublisher", onAlexandriaPublisher, false)
	datastore.RegisterMapping(apIndexName, "alexandria-publisher.json")
	pubRouter.HandleFunc("/get/latest/{limit:[0-9]+}", handleLatestPublishers)
	pubRouter.HandleFunc("/get/{address:[A-Za-z0-9]+}", handleGetPublisher)
}

func handleLatestPublishers(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	size, _ := strconv.ParseInt(opts["limit"], 10, 0)
	if size <= 0 || size > 1000 {
		size = -1
	}

	// q := elastic.NewBoolQuery().Must(
	// 	elastic.NewTermQuery("meta.deactivated", false),
	// )

	// fsc := elastic.NewFetchSourceContext(true).
	// 	Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(datastore.Index(apIndexName)).
		Type("_doc").
		// Query(q).
		Size(int(size)).
		Sort("timestamp", false).
		// FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		httpapi.RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(results.Hits.Hits),
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}

func handleGetPublisher(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("address", opts["address"]),
	)

	// fsc := elastic.NewFetchSourceContext(true).
	// 	Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(datastore.Index(apIndexName)).
		Type("_doc").
		Query(q).
		Size(1).
		Sort("timestamp", false).
		// FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		httpapi.RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
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
