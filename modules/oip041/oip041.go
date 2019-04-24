package oip041

import (
	"net/http"
	"strconv"

	"github.com/azer/logger"
	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/filters"
	"github.com/oipwg/oip/flo"
	"github.com/oipwg/oip/httpapi"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

const oip41IndexName = "oip041"

var artRouter = httpapi.NewSubRoute("/oip041/artifact")

func init() {
	log.Info("init oip41")
	events.SubscribeAsync("modules:oip:oip041", on41, false)

	datastore.RegisterMapping(oip41IndexName, "oip041.json")

	artRouter.HandleFunc("/get/latest", handleLatest).Queries("nsfw", "{nsfw}")
	artRouter.HandleFunc("/get/latest", handleLatest)
	artRouter.HandleFunc("/get/{id:[a-f0-9]+}", handleGet)
}

var (
	o41Fsc = elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")
	o41Indices = []string{oip41IndexName}
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

	searchService := httpapi.BuildCommonSearchService(r.Context(), o41Indices, q, []elastic.SortInfo{{Field: "meta.time", Ascending: false}}, o41Fsc)
	httpapi.RespondSearch(w, searchService)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		// elastic.NewTermQuery("_id", opts["id"]),
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	searchService := httpapi.BuildCommonSearchService(r.Context(), o41Indices, q, []elastic.SortInfo{{Field: "meta.time", Ascending: false}}, o41Fsc)
	httpapi.RespondSearch(w, searchService)
}

func on41(floData string, tx *datastore.TransactionData) {
	log.Info("oip041 ", tx.Transaction.Txid)

	any := jsoniter.Get([]byte(floData))

	el, err := validateOip041(any, tx)
	if err != nil {
		log.Error("validate oip041 failed", logger.Attrs{"err": err})
		return
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("oip041")).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

type elasticOip041 struct {
	Artifact interface{} `json:"artifact"`
	Meta     OMeta       `json:"meta"`
}
type OMeta struct {
	Block       int64                      `json:"block"`
	BlockHash   string                     `json:"block_hash"`
	Blacklist   Blacklist                  `json:"blacklist"`
	Deactivated bool                       `json:"deactivated"`
	Signature   string                     `json:"signature"`
	Time        int64                      `json:"time"`
	Tx          *datastore.TransactionData `json:"-"`
	Txid        string                     `json:"txid"`
	Type        string                     `json:"type"`
}
type Blacklist struct {
	Blacklisted bool   `json:"blacklisted"`
	Filter      string `json:"filter"`
}

func validateOip041(any jsoniter.Any, tx *datastore.TransactionData) (elasticOip041, error) {
	var el elasticOip041

	o41 := any.Get("oip-041")
	sig := o41.Get("signature")

	art := o41.Get("artifact")
	if art.LastError() != nil {
		return el, errors.Wrap(art.LastError(), "oip-041.artifact")
	}

	if len(art.Get("info", "title").ToString()) == 0 {
		return el, errors.New("artifact.info.title missing")
	}

	ok, err := flo.CheckAddress(art.Get("publisher").ToString())
	if !ok {
		return el, errors.Wrap(err, "invalid FLO address")
	}

	bl, label := filters.ContainsWithLabel(tx.Transaction.Txid)

	el.Artifact = art.GetInterface()
	el.Meta = OMeta{
		Block:       tx.Block,
		BlockHash:   tx.BlockHash,
		Blacklist:   Blacklist{Blacklisted: bl, Filter: label},
		Deactivated: false,
		Signature:   sig.ToString(),
		Time:        tx.Transaction.Time,
		Tx:          tx,
		Txid:        tx.Transaction.Txid,
		Type:        "oip041",
	}

	return el, nil
}
