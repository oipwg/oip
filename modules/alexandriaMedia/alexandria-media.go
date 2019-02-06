package alexandriaMedia

import (
	"net/http"
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
	artRouter.HandleFunc("/get/latest", handleLatest)
	artRouter.HandleFunc("/get/{id:[a-f0-9]+}", handleGet)
}

var (
	amIndices = []string{amIndexName}
	amFsc     = elastic.NewFetchSourceContext(true).
			Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")
)

func handleLatest(w http.ResponseWriter, r *http.Request) {

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)
	searchService := httpapi.BuildCommonSearchService(r.Context(), amIndices, q, []elastic.SortInfo{{Field: "meta.time", Ascending: false}}, amFsc)
	httpapi.RespondSearch(w, searchService)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)
	searchService := httpapi.BuildCommonSearchService(r.Context(), amIndices, q, []elastic.SortInfo{{Field: "meta.time", Ascending: false}}, amFsc)
	httpapi.RespondSearch(w, searchService)
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
