package oip

import (
	"strings"

	"github.com/azer/logger"
	"github.com/bitspill/oip/config"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/filters"
)

const minFloDataLen = 35

func init() {
	log.Info("init oip")
	if config.Testnet {
		events.Bus.SubscribeAsync("flo:floData", onFloDataTestNet, false)
	} else {
		events.Bus.SubscribeAsync("flo:floData", onFloDataMainNet, false)
	}
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
}
