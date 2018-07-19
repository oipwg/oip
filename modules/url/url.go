package url

import (
	"strings"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
)

func init() {
	events.Bus.SubscribeAsync("flo:floData", onFloData, false)
	events.Bus.SubscribeAsync("modules:url", onUrl, false)
}

func onFloData(floData string, tx datastore.TransactionData) {
	if strings.HasPrefix(floData, "http://") || strings.HasPrefix(floData, "https://") {
		events.Bus.Publish("modules:url", floData)
		return
	}
}

func onUrl(floData string) {

}
