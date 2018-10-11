package historian

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/httpapi"
	"github.com/bitspill/oipProto/go/oipProto"
	"github.com/gorilla/mux"
	"gopkg.in/olivere/elastic.v6"
)

const histDataPointIndexName = "historian_data_point_"

var histRouter = httpapi.NewSubRoute("/historian")

func init() {
	log.Info("init historian")
	events.Bus.SubscribeAsync("modules:historian:stringDataPoint", onStringHdp, false)
	events.Bus.SubscribeAsync("modules:historian:protoDataPoint", onProtoHdp, false)

	datastore.RegisterMapping(histDataPointIndexName+"string", histDataPointMapping)
	datastore.RegisterMapping(histDataPointIndexName+"proto", histDataPointMapping)

	histRouter.HandleFunc("/get/latest/{limit:[0-9]+}", handleLatest)
	histRouter.HandleFunc("/get/{id:[a-f0-9]+}", handleGet)
	histRouter.HandleFunc("/24hr", handle24hr)
}

func handleLatest(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	size, _ := strconv.ParseInt(opts["limit"], 10, 0)
	if size <= 0 || size > 1000 {
		size = -1
	}

	// q := elastic.NewBoolQuery().Must(
	// 	elastic.NewTermQuery("meta.deactivated", false),
	// )

	fsc := elastic.NewFetchSourceContext(true).
		Include("data_point.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(histDataPointIndexName).
		Type("_doc").
		// Query(q).
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
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	fsc := elastic.NewFetchSourceContext(true).
		Include("data_point.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(histDataPointIndexName).
		Type("_doc").
		Query(q).
		Size(1).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
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

	bir := elastic.NewBulkIndexRequest().Index(histDataPointIndexName + "string").Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func onProtoHdp(hdp *oipProto.HistorianDataPoint, tx *datastore.TransactionData) {
	log.Info("historian dataPoint ", tx.Transaction.Txid)

	var el elasticHdp
	el.DataPoint = hdp
	el.Meta = HMeta{
		Time:      tx.Transaction.Time,
		Txid:      tx.Transaction.Txid,
		BlockHash: tx.BlockHash,
		Block:     tx.Block,
		Tx:        tx,
	}
	bir := elastic.NewBulkIndexRequest().Index(histDataPointIndexName + "proto").Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

type elasticHdp struct {
	DataPoint interface{} `json:"data_point"`
	Meta      HMeta       `json:"meta"`
}
type HMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	Txid      string                     `json:"txid"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"tx"`
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
		Time:      tx.Transaction.Time,
		Txid:      tx.Transaction.Txid,
		BlockHash: tx.BlockHash,
		Block:     tx.Block,
		Tx:        tx,
	}

	return el, nil
}

const histDataPointMapping = `{
  "settings": {
    "number_of_shards": 2
  },
  "mappings": {
    "_doc": {
      "dynamic": "true",
      "properties": {
        "data_point" : {
			"type": "object"
        },
        "meta": {
          "properties": {
            "block": {
              "type": "long"
            },
            "block_hash": {
              "type": "keyword",
              "ignore_above": 64
            },
            "signature": {
              "type": "text",
              "index": false
            },
            "time": {
              "type": "date",
              "format": "epoch_second"
            },
            "tx": {
              "type": "object",
              "enabled": false
            },
            "txid": {
              "type": "keyword",
              "ignore_above": 64
            }
          }
        }
      }
    }
  }
}
`
