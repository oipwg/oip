package oip042

import (
	"strings"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/json-iterator/go"
	"gopkg.in/olivere/elastic.v6"
)

const oip042ArtifactIndex = `oip042_artifact`
const oip042PublisherIndex = `oip042_publisher`

func init() {
	log.Info("init oip042 json")
	events.Bus.SubscribeAsync("flo:floData", onFloData, false)
	events.Bus.SubscribeAsync("sync:floData:json", onJson, false)
	events.Bus.SubscribeAsync("modules:oip042:json", on42Json, false)

	datastore.RegisterMapping(oip042ArtifactIndex, publishOip042ArtifactMapping)
}

func on42Json(message jsoniter.RawMessage, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42Json", logger.Attrs{"txid": tx.Transaction.Txid})
	if !jsoniter.Valid(message) {
		log.Info("invalid json %s", tx.Transaction.Txid)
		return
	}

	publish := jsoniter.Get(message, "publish")
	err := publish.LastError()
	if err == nil {
		on42JsonPublish(publish, tx)
		return
	}
	register := jsoniter.Get(message, "register")
	err = publish.LastError()
	if err == nil {
		on42JsonRegister(register, tx)
		return
	}

	log.Error("no publisher/register message %s", tx.Transaction.Txid)
}

func on42JsonRegister(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonRegister", logger.Attrs{"txid": tx.Transaction.Txid})

	pub := any.Get("pub")
	err := pub.LastError()
	if err == nil {
		on42JsonRegisterPub(pub, tx)
		return
	}

	log.Error("no publish %s", tx.Transaction.Txid)
}

func on42JsonRegisterPub(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonRegisterPub", logger.Attrs{"txid": tx.Transaction.Txid})

	// name := any.Get("name").ToString()
	// if len(name) == 0 {
	//	log.Println("oip042 no pub.name")
	//	return
	// }

	bir := elastic.NewBulkIndexRequest().Index(oip042PublisherIndex).Type("_doc").Id(tx.Transaction.Txid).Doc(any.GetInterface())
	datastore.AutoBulk.Add(bir)
}

func on42JsonPublish(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonPublish", logger.Attrs{"txid": tx.Transaction.Txid})

	artifact := any.Get("artifact")
	err := artifact.LastError()
	if err != nil {
		log.Error("%s - %s", tx.Transaction.Txid, err.Error())
		return
	}

	title := artifact.Get("info", "title").ToString()
	if len(title) == 0 {
		log.Error("oip042 no title", logger.Attrs{"txid": tx.Transaction.Txid})
		return
	}

	bir := elastic.NewBulkIndexRequest().Index(oip042ArtifactIndex).Type("_doc").Id(tx.Transaction.Txid).Doc(artifact.GetInterface())
	datastore.AutoBulk.Add(bir)
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

	log.Error("no oip042", logger.Attrs{"txid": tx.Transaction.Txid})
}

func onFloData(floData string, tx datastore.TransactionData) {
	if tx.Block < 2000000 {
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
