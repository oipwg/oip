package oip

import (
	"strings"

	"github.com/azer/logger"
	"github.com/bitspill/oip/config"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/filters"
	"github.com/json-iterator/go"
)

const minFloDataLen = 35

func init() {
	log.Info("init oip")
	if config.Testnet {
		events.Bus.SubscribeAsync("flo:floData", onFloDataTestNet, false)
	} else {
		events.Bus.SubscribeAsync("flo:floData", onFloDataMainNet, false)
	}
	events.Bus.SubscribeAsync("sync:floData:json", onJson, false)
}

func onFloDataMainNet(floData string, tx datastore.TransactionData) {
	if len(floData) < minFloDataLen {
		// impossible to be a valid item at such a short length
		return
	}
	if tx.Block < 1000000 {
		return
	}

	if filters.Contains(tx.Transaction.Txid) {
		log.Error("Filtered out transaction", logger.Attrs{"txid": tx.Transaction.Txid})
		return
	}

	simplified := strings.TrimSpace(floData[0:35])
	simplified = strings.Replace(simplified, " ", "", -1)

	if (tx.Block > 2263000 && strings.HasPrefix(simplified, "oip-mp(")) ||
		(tx.Block < 2400000 && strings.HasPrefix(simplified, "alexandria-media-multipart(")) {
		events.Bus.Publish("modules:oip:multipartSingle", floData, tx)
		return
	}

	if strings.HasPrefix(simplified, `{"alexandria-publisher":`) {
		events.Bus.Publish("modules:oip:alexandriaPublisher", floData, tx)
		return
	}

	if tx.Block < 2400000 {
		if strings.HasPrefix(simplified, `{"alexandria-deactivation":`) {
			events.Bus.Publish("modules:oip:alexandriaDeactivation", floData, tx)
			return
		}
		if strings.HasPrefix(simplified, `{"alexandria-media":`) {
			events.Bus.Publish("modules:oip:alexandriaMedia", floData, tx)
			return
		}
	}

	if tx.Block < 2000000 {
		return
	}

	if strings.HasPrefix(simplified, `{"oip-041":`) {
		events.Bus.Publish("modules:oip:oip041", floData, tx)
		return
	}

	if processPrefix("json:", "sync:floData:json", floData, tx) {
		return
	}
	// if processPrefix("gz:", "sync:floData:gz", floData, tx) {
	//	return
	// }
	if processPrefix("p64:", "sync:floData:p64", floData, tx) {
		return
	}

}

func onFloDataTestNet(floData string, tx datastore.TransactionData) {
	if len(floData) < minFloDataLen {
		// impossible to be a valid item at such a short length
		return
	}

	if filters.Contains(tx.Transaction.Txid) {
		log.Error("Filtered out transaction", logger.Attrs{"txid": tx.Transaction.Txid})
		return
	}

	simplified := strings.TrimSpace(floData[0:35])
	simplified = strings.Replace(simplified, " ", "", -1)

	if strings.HasPrefix(simplified, "oip-mp(") ||
		strings.HasPrefix(simplified, "alexandria-media-multipart(") {
		events.Bus.Publish("modules:oip:multipartSingle", floData, tx)
		return
	}
	if strings.HasPrefix(simplified, `{"alexandria-publisher":`) {
		events.Bus.Publish("modules:oip:alexandriaPublisher", floData, tx)
		return
	}
	if strings.HasPrefix(simplified, `{"alexandria-deactivation":`) {
		events.Bus.Publish("modules:oip:alexandriaDeactivation", floData, tx)
		return
	}
	if strings.HasPrefix(simplified, `{"alexandria-media":`) {
		events.Bus.Publish("modules:oip:alexandriaMedia", floData, tx)
		return
	}
	if strings.HasPrefix(simplified, `{"oip-041":`) {
		events.Bus.Publish("modules:oip:oip041", floData, tx)
		return
	}
	if processPrefix("json:", "sync:floData:json", floData, tx) {
		return
	}
	// if processPrefix("gz:", "sync:floData:gz", floData, tx) {
	//	return
	// }
	if processPrefix("p64:", "sync:floData:p64", floData, tx) {
		return
	}

}

func processPrefix(prefix string, namespace string, floData string, tx datastore.TransactionData) bool {
	if strings.HasPrefix(floData, prefix) {
		log.Info("prefix match", logger.Attrs{"txid": tx.Transaction.Txid, "prefix": prefix, "namespace": namespace})
		events.Bus.Publish(namespace, strings.TrimPrefix(floData, prefix), tx)
		return true
	}
	return false
}

func onJson(floData string, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("onJson", logger.Attrs{"txid": tx.Transaction.Txid})
	var dj map[string]jsoniter.RawMessage
	err := jsoniter.Unmarshal([]byte(floData), &dj)
	if err != nil {
		return
	}

	if o42, ok := dj["oip042"]; ok {
		log.Info("sending oip042 message", logger.Attrs{"txid": tx.Transaction.Txid})
		events.Bus.Publish("modules:oip042:json", o42, tx)
		return
	}

	log.Error("no supported json type", logger.Attrs{"txid": tx.Transaction.Txid})
}
