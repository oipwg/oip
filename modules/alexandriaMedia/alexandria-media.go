package alexandriaMedia

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/httpapi"
	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"gopkg.in/olivere/elastic.v6"
)

const amIndexName = "alexandria-media"

var artRouter = httpapi.NewSubRoute("/alexandria/artifact")

func init() {
	log.Info("init alexandria-media")
	events.Bus.SubscribeAsync("modules:oip:alexandriaMedia", onAlexandriaMedia, false)
	datastore.RegisterMapping(amIndexName, "alexandria-media.json")
	artRouter.HandleFunc("/get/latest/{limit:[0-9]+}", handleLatest)
	artRouter.HandleFunc("/get/{id:[a-f0-9]+}", handleGet)
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

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")

	results, err := datastore.Client().
		Search(datastore.Index(amIndexName)).
		Type("_doc").
		Query(q).
		Size(int(size)).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
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

func handleGet(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")

	results, err := datastore.Client().
		Search(datastore.Index(amIndexName)).
		Type("_doc").
		Query(q).
		Size(1).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
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

func onAlexandriaMedia(floData string, tx *datastore.TransactionData) {
	attr := logger.Attrs{"txid": tx.Transaction.Txid}
	log.Info("onAlexandriaMedia", attr)

	bytesFloData := []byte(floData)
	a := jsoniter.Get(bytesFloData)
	am := a.Get("alexandria-media")
	title := am.Get("info", "title").ToString()
	artTime := am.Get("timestamp").ToInt64()
	if artTime > 999999999999 {
		// correct ms timestamps to s
		attr["artTime"] = artTime
		attr["time"] = time.Unix(artTime, 0)
		log.Info("Scaling timestamp", attr)
		artTime /= 1000
	}
	if len(title) != 0 {
		el := elasticAm{
			Artifact: am.GetInterface(),
			Meta: AmMeta{
				Block:       tx.Block,
				BlockHash:   tx.BlockHash,
				Deactivated: false,
				Signature:   a.Get("signature").ToString(),
				Time:        artTime,
				Tx:          tx,
				Txid:        tx.Transaction.Txid,
				Type:        "alexandria-media",
			},
		}

		bir := elastic.NewBulkIndexRequest().Index(datastore.Index(amIndexName)).Type("_doc").Doc(el).Id(tx.Transaction.Txid)
		datastore.AutoBulk.Add(bir)
	} else {
		log.Info("no title", attr)
	}
}

type elasticAm struct {
	Artifact interface{} `json:"artifact"`
	Meta     AmMeta      `json:"meta"`
}

type AmMeta struct {
	Block       int64                      `json:"block"`
	BlockHash   string                     `json:"block_hash"`
	Deactivated bool                       `json:"deactivated"`
	Signature   string                     `json:"signature"`
	Time        int64                      `json:"time"`
	Tx          *datastore.TransactionData `json:"tx"`
	Txid        string                     `json:"txid"`
	Type        string                     `json:"type"`
}
