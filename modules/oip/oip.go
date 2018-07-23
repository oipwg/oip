package oip

import (
	"strings"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
)

const MinFloDataLen = 35

func init() {
	log.Info("init oip")
	events.Bus.SubscribeAsync("flo:floData", onFloData, false)
	// events.Bus.SubscribeAsync("flo:newBlockCompleted", nil, false)
	// events.Bus.SubscribeAsync("flo:initialSyncCompleted", nil, false)
}

func onFloData(floData string, tx datastore.TransactionData) {
	if len(floData) < MinFloDataLen {
		// impossible to be a valid item at such a short length
		return
	}
	if tx.Block < 1000000 {
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
