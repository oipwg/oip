package oip

import (
	"encoding/json"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	log.Info("init alexandria-publisher")
	events.Bus.SubscribeAsync("modules:oip:alexandriaPublisher", onAlexandriaPublisher, false)
	datastore.RegisterMapping("alexandria-publisher", apMapping)
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
