package alexandriaMedia

import (
	"encoding/json"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"gopkg.in/olivere/elastic.v6"
)

const adIndexName = `alexandria-deactivation`

func init() {
	log.Info("init alexandria-deactivation")
	events.Bus.SubscribeAsync("modules:oip:alexandriaDeactivation", onAlexandriaDeactivation, false)
	datastore.RegisterMapping(adIndexName, adMapping)
}

func onAlexandriaDeactivation(floData string, tx datastore.TransactionData) {
	var ap map[string]json.RawMessage
	err := json.Unmarshal([]byte(floData), &ap)
	if err != nil {
		return
	}
	if d, ok := ap["alexandria-deactivation"]; ok {
		bir := elastic.NewBulkIndexRequest().Index(adIndexName).Type("_doc").Doc(d).Id(tx.Transaction.Txid)
		datastore.AutoBulk.Add(bir)
	}
}

const adMapping = `{
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
        "txid": {
          "type": "keyword",
          "ignore_above": 64
        }
      }
    }
  }
}`
