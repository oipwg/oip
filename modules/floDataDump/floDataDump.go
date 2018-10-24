package floDataDump

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
)

var f *os.File

var (
	totalFloData    int64
	textFloData     int64
	coinbaseFloData int64
)

func init() {
	log.Info("init floDataDump")
	events.Bus.SubscribeAsync("flo:floData", onFloData, false)
	events.Bus.SubscribeAsync("datastore:commit", onCommit, false)

	var err error
	f, err = os.OpenFile("textComments.txt", os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
}

func onCommit() {
	if totalFloData-coinbaseFloData > 0 {
		log.Info("%d/%d %d %f%%\n", textFloData, totalFloData-coinbaseFloData, totalFloData,
			float64(textFloData*10000/(totalFloData-coinbaseFloData))/10000)
		err := f.Sync()
		if err != nil {
			panic(err)
		}
	}
}

func onFloData(floData string, tx *datastore.TransactionData) {
	totalFloData++

	if tx.Transaction.Vin[0].IsCoinBase() {
		coinbaseFloData++
	}

	if strings.HasPrefix(floData, "text:") {
		textFloData++
		f.WriteString(fmt.Sprintf("%8d %s - %s\n", tx.Block, tx.Transaction.Txid, floData))
	}
}
