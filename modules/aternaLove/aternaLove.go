package aternaLove

import (
	"strings"

	"github.com/azer/logger"
	"github.com/bitspill/oip/config"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	log.Info("init aterna")
	if !config.IsTestnet() {
		events.Bus.SubscribeAsync("flo:floData", onFloData, false)
		events.Bus.SubscribeAsync("modules:aternaLove:alove", onAlove, false)
		datastore.RegisterMapping("aterna", aternaMapping)
	}
}

func onFloData(floData string, tx *datastore.TransactionData) {
	if tx.Block < 500000 {
		return
	}
	if tx.Block > 1000000 {
		events.Bus.Unsubscribe("flo:floData", onFloData)
		events.Bus.Unsubscribe("modules:aternaLove:alove", onAlove)
	}

	prefix := "t1:ALOVE>"
	if strings.HasPrefix(floData, prefix) {
		events.Bus.Publish("modules:aternaLove:alove", strings.TrimPrefix(floData, prefix), tx)
		return
	}
}

func onAlove(floData string, tx *datastore.TransactionData) {
	chunks := strings.SplitN(floData, "|", 3)
	if len(chunks) != 3 {
		log.Error("invalid aterna", logger.Attrs{"txid": tx.Transaction.Txid, "floData": floData})
		return
	}

	a := Alove{
		Message: chunks[0],
		From:    chunks[2],
		To:      chunks[1],
		TxId:    tx.Transaction.Txid,
	}
	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("aterna")).Type("_doc").Id(tx.Transaction.Txid).Doc(a)
	datastore.AutoBulk.Add(bir)
}

type Alove struct {
	Message string `json:"message"`
	To      string `json:"to"`
	From    string `json:"from"`
	TxId    string `json:"txId"`
}

const aternaMapping = `{
  "settings": {
    "number_of_shards": 2
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "from": {
          "type": "text",
          "fields": {
            "keyword": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        },
        "message": {
          "type": "text",
          "fields": {
            "keyword": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        },
        "to": {
          "type": "text",
          "fields": {
            "keyword": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        },
        "txId": {
          "type": "keyword",
          "ignore_above": 64
        }
      }
    }
  }
}`
