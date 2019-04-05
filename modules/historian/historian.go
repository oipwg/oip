package historian

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/gorilla/mux"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/httpapi"
	"github.com/oipwg/oip/oipProto"
	"gopkg.in/olivere/elastic.v6"
)

const histDataPointIndexName = "historian_data_point_"

var histRouter = httpapi.NewSubRoute("/historian")

func init() {
	log.Info("init historian")
	events.SubscribeAsync("modules:historian:stringDataPoint", onStringHdp, false)
	events.SubscribeAsync("modules:historian:protoDataPoint", onProtoHdp, false)

	datastore.RegisterMapping(histDataPointIndexName+"string", "historianDataPoint.json")
	datastore.RegisterMapping(histDataPointIndexName+"proto", "historianDataPoint.json")

	histRouter.HandleFunc("/get/latest", handleLatest)
	histRouter.HandleFunc("/get/{id:[a-f0-9]+}", handleGet)
	histRouter.HandleFunc("/24hr", handle24hr)
}

var (
	hdpFsc = elastic.NewFetchSourceContext(true).
		Include("data_point.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")
	hdpIndices = []string{histDataPointIndexName + "string", histDataPointIndexName + "proto"}
)

func handleLatest(w http.ResponseWriter, r *http.Request) {
	q := elastic.NewBoolQuery().Must(
		elastic.NewExistsQuery("_id"),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		hdpIndices,
		q,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: false},
		},
		hdpFsc,
	)
	httpapi.RespondSearch(w, searchService)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		hdpIndices,
		q,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Ascending: false, Field: "meta.txid"},
		},
		hdpFsc,
	)
	httpapi.RespondSearch(w, searchService)
}

func handle24hr(w http.ResponseWriter, r *http.Request) {
	// ToDo
	httpapi.RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
		"err": "not implemented",
	})
}

func onStringHdp(floData string, tx *datastore.TransactionData) {
	log.Info("historian dataPoint ", tx.Transaction.Txid)

	el, err := validateHdp(floData, tx)
	if err != nil {
		log.Error("validate historian dataPoint failed", logger.Attrs{"err": err})
		return
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(histDataPointIndexName + "string")).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func onProtoHdp(hdp *oipProto.HistorianDataPoint, tx *datastore.TransactionData) {
	log.Info("historian dataPoint ", tx.Transaction.Txid)

	var el elasticHdp
	el.DataPoint = hdp
	el.Meta = HMeta{
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
	}
	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(histDataPointIndexName + "proto")).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

type elasticHdp struct {
	DataPoint interface{} `json:"data_point"`
	Meta      HMeta       `json:"meta"`
}
type HMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"tx"`
	Txid      string                     `json:"txid"`
}
type DataPoint struct {
	Version      int     `json:"version,omitempty"`
	URL          string  `json:"url,omitempty"`
	Address      string  `json:"address,omitempty"`
	MrrLast10    float64 `json:"mrr_last_10,omitempty"`
	MrrLast24hr  float64 `json:"mrr_last_24hr,omitempty"`
	PoolHashrate float64 `json:"pool_hashrate,omitempty"`
	FbdHashrate  float64 `json:"fbd_hashrate,omitempty"`
	FmdWeighted  float64 `json:"fmd_weighted,omitempty"`
	FmdUsd       float64 `json:"fmd_usd,omitempty"`
	CmcLtc       float64 `json:"cmc_ltc,omitempty"`
	Signature    string  `json:"signature,omitempty"`
}

type hdpV int

const (
	alexV1 = iota
	oipV1  = iota
	oipV2  = iota
	oipV3  = iota
)

func validateHdp(floData string, tx *datastore.TransactionData) (elasticHdp, error) {
	if tx.Block > 2731000 {
		return elasticHdp{}, errors.New("deprecated")
	}

	// there are no bounds or error checks below since data points are no longer being published
	// all historical datapoints are known valid
	var el elasticHdp
	var hdp DataPoint
	var v hdpV

	// alexandria-historian-v001
	if floData[0] == 'a' {
		v = alexV1
		hdp.Version = 1
	}

	// oip-historian-3
	// oip-historian-2
	// oip-historian-1
	if floData[0] == 'o' {
		sv := floData[14]
		switch sv {
		case '1':
			v = oipV1
			hdp.Version = 1
		case '2':
			v = oipV2
			hdp.Version = 2
		case '3':
			v = oipV3
			hdp.Version = 3
		}
	}

	parts := strings.Split(floData, ":")

	if v == alexV1 {
		hdp.URL = parts[1]
	} else {
		hdp.Address = parts[1]
	}
	hdp.Signature = parts[len(parts)-1]

	i := 2
	hdp.MrrLast10, _ = strconv.ParseFloat(parts[i], 64)
	if v >= oipV2 {
		i++
		hdp.MrrLast24hr, _ = strconv.ParseFloat(parts[i], 64)
		if math.IsNaN(hdp.MrrLast24hr) {
			hdp.MrrLast24hr = 0
		}
	}
	i++
	hdp.PoolHashrate, _ = strconv.ParseFloat(parts[i], 64)
	i++
	hdp.FbdHashrate, _ = strconv.ParseFloat(parts[i], 64)
	i++
	hdp.FmdWeighted, _ = strconv.ParseFloat(parts[i], 64)
	i++
	hdp.FmdUsd, _ = strconv.ParseFloat(parts[i], 64)
	if v == oipV3 {
		i++
		hdp.CmcLtc, _ = strconv.ParseFloat(parts[i], 64)
	}

	el.DataPoint = hdp
	el.Meta = HMeta{
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
	}

	return el, nil
}
