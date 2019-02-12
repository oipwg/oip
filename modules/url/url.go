package url

import (
	"strings"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
)

func init() {
	events.SubscribeAsync("flo:floData", onFloData, false)
	events.SubscribeAsync("modules:url", onUrl, false)
}

func onFloData(floData string, tx *datastore.TransactionData) {
	if strings.HasPrefix(floData, "http://") || strings.HasPrefix(floData, "https://") {
		events.Publish("modules:url", floData)
		return
	}
}

func onUrl(floData string) {

}
