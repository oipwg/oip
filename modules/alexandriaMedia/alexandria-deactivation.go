package alexandriaMedia

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/flo"
	"gopkg.in/olivere/elastic.v6"
)

const adIndexName = `alexandria-deactivation`

var deactivationCommitMutex sync.Mutex

func init() {
	log.Info("init alexandria-deactivation")
	events.Bus.SubscribeAsync("modules:oip:alexandriaDeactivation", onAlexandriaDeactivation, false)
	events.Bus.SubscribeAsync("modules:oip:mpCompleted", onMpCompleted, false)
	datastore.RegisterMapping(adIndexName, adMapping)
}

func onAlexandriaDeactivation(floData string, tx *datastore.TransactionData) {
	var ad floAd
	err := json.Unmarshal([]byte(floData), &ad)
	if err != nil {
		log.Error("unable to unmarshal json", logger.Attrs{"txid": tx.Transaction.Txid})
		return
	}

	// signature pre-image for deactivation is <address>-<txid>
	ok, err := flo.CheckSignature(ad.AlexandriaDeactivation.Address, ad.Signature, ad.AlexandriaDeactivation.Address+"-"+ad.AlexandriaDeactivation.Txid)
	if !ok {
		log.Error("signature validation failed", logger.Attrs{"txid": tx.Transaction.Txid, "err": err})
		return
	}

	var ead = elasticAd{
		Address:   ad.AlexandriaDeactivation.Address,
		Reference: ad.AlexandriaDeactivation.Txid,
		Signature: ad.Signature,
		Meta: AdMeta{
			Block:     tx.Block,
			BlockHash: tx.BlockHash,
			Complete:  false,
			Stale:     false,
			Time:      tx.Transaction.Time,
			Tx:        tx,
			Txid:      tx.Transaction.Txid,
		},
	}

	bir := elastic.NewBulkIndexRequest().Index(adIndexName).Type("_doc").Doc(ead).Id(tx.Transaction.Txid)
	datastore.AutoBulk.Add(bir)
}

func onMpCompleted() {
	exist, err := datastore.Client().IndexExists(adIndexName).Do(context.TODO())
	if err != nil {
		log.Error("elastic index exists failed", logger.Attrs{"err": err, "index": adIndexName})
		return
	}
	if !exist {
		log.Info("elastic index doesn't exist", logger.Attrs{"index": adIndexName})
		return
	}

	deactivationCommitMutex.Lock()
	defer deactivationCommitMutex.Unlock()

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.complete", false),
		elastic.NewTermQuery("meta.stale", false),
	)
	results, err := datastore.Client().Search(adIndexName).Type("_doc").Query(q).Size(10000).Sort("meta.time", false).Do(context.TODO())
	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		return
	}

	if len(results.Hits.Hits) == 0 {
		// early abort
		return
	}

	log.Info("Collecting deactivates to attempt applying", logger.Attrs{"pendingDeactivations": len(results.Hits.Hits)})

	for _, v := range results.Hits.Hits {
		var ea elasticAd
		err := json.Unmarshal(*v.Source, &ea)
		if err != nil {
			log.Info("failed to unmarshal elastic hit", logger.Attrs{"err": err, "source": *v.Source, "id": v.Id})
			continue
		}

		// deactivate the artifact
		s := elastic.NewScript("ctx._source.meta.deactivated=true;").Type("inline").Lang("painless")
		up := elastic.NewBulkUpdateRequest().Index(amIndexName).Id(ea.Reference).Type("_doc").Script(s)
		datastore.AutoBulk.Add(up)
		// All attempted oip-041 deactivation appear to be invalid
		// up = elastic.NewBulkUpdateRequest().Index("oip041").Id(ea.Reference).Type("_doc").Script(s)
		// datastore.AutoBulk.Add(up)

		// tag deactivation as completed
		s = elastic.NewScript("ctx._source.meta.complete=true;").Type("inline").Lang("painless")
		up = elastic.NewBulkUpdateRequest().Index(adIndexName).Id(ea.Meta.Txid).Type("_doc").Script(s)
		datastore.AutoBulk.Add(up)
	}
}

type floAd struct {
	AlexandriaDeactivation struct {
		Txid    string `json:"txid"`
		Address string `json:"address"`
	} `json:"alexandria-deactivation"`
	Signature string `json:"signature"`
}

type elasticAd struct {
	Reference string `json:"reference"`
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Meta      AdMeta `json:"meta"`
}

type AdMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	Complete  bool                       `json:"complete"`
	Stale     bool                       `json:"stale"`
	Txid      string                     `json:"txid"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"tx"`
}

const adMapping = `{
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
        "reference": {
          "type": "keyword",
          "ignore_above": 36
        },
        "signature": {
          "type": "text",
          "index": false
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
            "complete": {
              "type": "boolean"
            },
            "stale": {
              "type": "boolean"
            },
            "time": {
              "type": "date",
              "format": "epoch_second"
            },
            "txid": {
              "type": "keyword",
              "ignore_above": 64
            },
            "tx": {
              "type": "object",
              "enabled": false
            }
          }
        }
      }
    }
  }
}`
