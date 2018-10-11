package floDataDump

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
)

var spams = []string{
	`http`,
	`text:`,
	`Prohashing`,
	`Mined at CoinMine.PW`,
	`Multipool`,
	`{"total_length"`,
	`{"first_txid": `,
	`{"alexandria-media"`,
	`{ "alexandria-media": `,
	`{ "alexandria-publisher":`,
	`{ "alexandria-deactivation":`,
	`{"alexandria-historian"`,
	`{"alexandria-history-record`,
	` { "alexandria-historian"`,
	` { "alexandria-history-record`,
	`{"alexandria-media": `,
	`{"alexandria-publisher":`,
	`{"alexandria-deactivation":`,
	` { "alexandria-media": `,
	` { "alexandria-publisher":`,
	` { "alexandria-deactivation":`,
	`alexandria-media-multipart(`,
	`This document has been flotorized:`,
	`{"alexandria-publisher"`,
	`oip-mp(`,
	`A_blockchain_spamtest_by_MaGNeT_so_if_you_read_this_it_works`,
	`xxxxxxxxxxxxxxxxxxxxxxxxxx_A_blockchain_spamtest_by_MaGNeT_so_if_you_read_this_it_works`,
	`t1:`,
	`botpool`,
	`fun_pool`,
	`RT @`,
	`@`,
	`CryptoPools`,
	`{"title" : "Alexandria"`,
	`{"app" : "Alexandria"`,
	`Thanks for using the FLO faucet at http://florincoin.info/faucet`,
	`Thanks for participating in the FLO 1st birthday giveaway`,
	`Alexandria:{"v":"1.0","p":`,
	`localhost/?tp=`,
	`{"column 1":`,
	`{"PAGE_IMAGES"`,
	`{"id":"`,
	`pool.alexandria.media`,
	`alexandria-historian-v001`,
	`{"Slidechain ID"`,
	`{"line"`,
	`U2FsdGVkX1`,
	`{"source":"blockchainlogger.com"`,
	`{"timestamp"`,
	`{ "artifact"`,
	`{"oip-041":`,
	`{"artifact"`,
	`{
"alexandria-autominer"`,
	`{
"alexandria-retailer"`,
	`oip-historian-1`,
	`oip-historian-2`,
	`Test run - flotorizer`,
	`penlau profit pool`,
	`api.alexandria.io/pool/`,
	`caaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`,
	`undefined`,
	`Hello from oip-mw :)`,
	`stringabc`,
	`asdfghjklasdfghjk`,
	`[object Object]`,
	`This is a message.`,
	`098765432109876`,
	`Test1`,
	`Test2`,
	`Test3`,
	`Test4`,
	`Test5`,
	`Test6`,
	`Test7`,
	`Test8`,
	`Test9`,
	`Test10`,
	`Test11`,
	`Test12`,
	`oip-historian-3`,
	`This is the message`,
	`Yet another test`,
	`d9dede25-034a-488d-b873-1643f39ec79c`,
	`10d8cbc7-ff97-4df4-b61f-86db7f40ebd2`,
	`bc024743-bd10-4d40-8fb8-e0f4b7d45041`,
	`FloSecretShares1.0v-beta`,
	`SharedSecret1.0v-beta`,
	`Hello world!`,
	`Cancel: SOI(`,
	`Inventory Posted: SOI(`,
	`Execution Report: SOI(`,
	`Client Interest: SOI(`,
	`Session opening date:`,
	`MiningCore payout`,
	`Flo MiningCore`,
	`tPool.io`,
	`p64:`,
}

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
	// if !spamCheck(floData) {
	//	log.Info(floData)
	// }

	totalFloData++

	if tx.Transaction.Vin[0].IsCoinBase() {
		coinbaseFloData++
	}

	if strings.HasPrefix(floData, "text:") {
		textFloData++
		f.WriteString(fmt.Sprintf("%8d %s - %s\n", tx.Block, tx.Transaction.Txid, floData))
	}
}

func spamCheck(floData string) bool {
	for _, s := range spams {
		if strings.HasPrefix(floData, s) {
			return true
		}
	}
	return false
}
