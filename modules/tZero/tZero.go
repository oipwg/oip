package tZero

import (
	"regexp"
	"strings"

	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/config"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

func init() {
	log.Info("init tZero")
	if !config.IsTestnet() {
		events.SubscribeAsync("flo:floData", floDataProcessor)
		events.SubscribeAsync("modules:tZero:cancel", onCancel)
		events.SubscribeAsync("modules:tZero:inventoryPosted", onInventoryPosted)
		events.SubscribeAsync("modules:tZero:executionReport", onExecutionReport)
		events.SubscribeAsync("modules:tZero:clientInterest", onClientInterest)
		datastore.RegisterMapping("tzero", "tZero.json")
	}
}

func floDataProcessor(floData string, tx *datastore.TransactionData) {
	if tx.Block < 2000000 {
		return
	}

	if strings.HasPrefix(floData, "Cancel: ") {
		events.Publish("modules:tZero:cancel", floData, tx)
		return
	}
	if strings.HasPrefix(floData, "Inventory Posted: ") {
		events.Publish("modules:tZero:inventoryPosted", floData, tx)
		return
	}
	if strings.HasPrefix(floData, "Execution Report: ") {
		events.Publish("modules:tZero:executionReport", floData, tx)
		return
	}
	if strings.HasPrefix(floData, "Client Interest: ") {
		events.Publish("modules:tZero:clientInterest", floData, tx)
		return
	}
}

func onCancel(floData string, tx *datastore.TransactionData) {
	gi := extractGeneralInfo(floData)

	gi.Action = "Cancel"
	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("tzero")).Type("_doc").Id(tx.Transaction.Txid).Doc(gi)
	datastore.AutoBulk.Add(bir)
}

func onInventoryPosted(floData string, tx *datastore.TransactionData) {
	gi := extractGeneralInfo(floData)

	gi.Action = "InventoryPosted"
	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("tzero")).Type("_doc").Id(tx.Transaction.Txid).Doc(gi)
	datastore.AutoBulk.Add(bir)
}

func onClientInterest(floData string, tx *datastore.TransactionData) {
	gi := extractGeneralInfo(floData)

	gi.Action = "ClientInterest"
	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("tzero")).Type("_doc").Id(tx.Transaction.Txid).Doc(gi)
	datastore.AutoBulk.Add(bir)
}

func onExecutionReport(floData string, tx *datastore.TransactionData) {
	gi := extractGeneralInfo(floData)

	gi.Action = "ExecutionReport"
	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("tzero")).Type("_doc").Id(tx.Transaction.Txid).Doc(gi)
	datastore.AutoBulk.Add(bir)
}

var giRe = regexp.MustCompile(`([A-Za-z]+)\((.*?)\)`)

func extractGeneralInfo(s string) tZeroTransaction {
	matches := giRe.FindAllStringSubmatch(s, -1)

	tgi := tZeroTransaction{}
	for _, value := range matches {
		if len(value) == 3 {
			switch value[1] {
			case "SOI":
				tgi.SOI = value[2]
			case "STI":
				tgi.STI = value[2]
			case "Broker":
				tgi.Broker = value[2]
			case "Account":
				tgi.Account = value[2]
			case "Time":
				tgi.Time = value[2]
			case "Side":
				tgi.Side = value[2]
			case "Symbol":
				tgi.Symbol = value[2]
			case "Qty":
				tgi.Qty = value[2]
			case "Price":
				tgi.Price = value[2]
			case "OrderType":
				tgi.OrderType = value[2]
			case "TimeInForce":
				tgi.TimeInForce = value[2]
			}
		}
	}
	return tgi
}

type tZeroTransaction struct {
	SOI         string `json:"soi"`
	STI         string `json:"sti"`
	Broker      string `json:"broker"`
	Account     string `json:"account"`
	Time        string `json:"time"`
	Side        string `json:"side"`
	Symbol      string `json:"symbol"`
	Qty         string `json:"qty"`
	Price       string `json:"price"`
	OrderType   string `json:"order_type"`
	TimeInForce string `json:"time_in_force"`
	Action      string `json:"action"`
}
