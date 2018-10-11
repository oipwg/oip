package alexandriaMedia

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/httpapi"
	"github.com/gorilla/mux"
	"gopkg.in/olivere/elastic.v6"
)

const apIndexName = "alexandria-publisher"

var pubRouter = httpapi.NewSubRoute("/alexandria/publisher")

func init() {
	log.Info("init alexandria-publisher")
	events.Bus.SubscribeAsync("modules:oip:alexandriaPublisher", onAlexandriaPublisher, false)
	datastore.RegisterMapping(apIndexName, apMapping)
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
		Search(apIndexName).
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
		Search(apIndexName).
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

func onAlexandriaPublisher(floData string, tx datastore.TransactionData) {
	var ap map[string]json.RawMessage
	err := json.Unmarshal([]byte(floData), &ap)
	if err != nil {
		return
	}
	if pub, ok := ap["alexandria-publisher"]; ok {
		bir := elastic.NewBulkIndexRequest().Index("alexandria-publisher").Type("_doc").Doc(pub).Id(tx.Transaction.Txid)
		datastore.AutoBulk.Add(bir)
	}
}

const apMapping = `{
  "settings": {
    "number_of_shards": 2
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "address": {
          "type": "keyword",
          "ignore_above": 36
        },
        "bitcoin": {
          "type": "keyword",
          "ignore_above": 36
        },
        "bitmessage": {
          "type": "keyword",
          "ignore_above": 256
        },
        "emailmd5": {
          "type": "keyword",
          "ignore_above": 32
        },
        "email": {
          "type": "keyword",
          "ignore_above": 256
        },
        "name": {
          "type": "text",
          "fields": {
            "keyword": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        },
        "timestamp": {
          "type": "date",
          "format": "epoch_second"
        }
      }
    }
  }
}`
