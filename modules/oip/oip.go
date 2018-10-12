package oip

import (
	"encoding/base64"
	"strings"

	"github.com/azer/logger"
	"github.com/bitspill/oip/btc"
	"github.com/bitspill/oip/config"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/filters"
	"github.com/bitspill/oip/flo"
	"github.com/bitspill/oipProto/go/oipProto"
	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
)

const minFloDataLen = 35

func init() {
	log.Info("init oip")
	if config.IsTestnet() {
		events.Bus.SubscribeAsync("flo:floData", onFloDataTestNet, false)
	} else {
		events.Bus.SubscribeAsync("flo:floData", onFloDataMainNet, false)
	}
	events.Bus.SubscribeAsync("sync:floData:json", onJson, false)
	events.Bus.SubscribeAsync("sync:floData:p64", onP64, false)
}

func onFloDataMainNet(floData string, tx *datastore.TransactionData) {
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

	if tx.Block < 2731000 && tx.Transaction.Vin[0].IsCoinBase() {
		// oip-historian-3
		// oip-historian-2
		// oip-historian-1
		// alexandria-historian-v001
		if strings.HasPrefix(simplified, "oip-historian-") ||
			strings.HasPrefix(simplified, "alexandria-historian-") {
			events.Bus.Publish("modules:historian:stringDataPoint", floData, tx)
		}
	}

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
	// 	return
	// }
	if processPrefix("p64:", "sync:floData:p64", floData, tx) {
		return
	}

}

func onFloDataTestNet(floData string, tx *datastore.TransactionData) {
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
	// 	return
	// }
	if processPrefix("p64:", "sync:floData:p64", floData, tx) {
		return
	}

}

func processPrefix(prefix, namespace, floData string, tx *datastore.TransactionData) bool {
	if strings.HasPrefix(floData, prefix) {
		log.Info("prefix match", logger.Attrs{"txid": tx.Transaction.Txid, "prefix": prefix, "namespace": namespace})
		events.Bus.Publish(namespace, strings.TrimPrefix(floData, prefix), tx)
		return true
	}
	return false
}

func onJson(floData string, tx *datastore.TransactionData) {
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

func onP64(p64 string, tx *datastore.TransactionData) {
	t := log.Timer()
	defer t.End("onP64", logger.Attrs{"txid": tx.Transaction.Txid})

	b, err := base64.StdEncoding.DecodeString(p64)
	if err != nil {
		log.Error("unable to decode base 64 message",
			logger.Attrs{"txid": tx.Transaction.Txid, "p64": p64, "err": err})
		return
	}

	var msg oipProto.SignedMessage
	err = proto.Unmarshal(b, &msg)
	if err != nil {
		log.Error("unable to unmarshal protobuf message",
			logger.Attrs{"txid": tx.Transaction.Txid, "p64": p64, "err": err})
		return
	}

	signature := base64.StdEncoding.EncodeToString(msg.Signature)
	pubKey := string(msg.PubKey)
	signedMessage := base64.StdEncoding.EncodeToString(msg.SerializedMessage)

	switch msg.SignatureType {
	case oipProto.SignatureTypes_Btc:
		valid, err := btc.CheckSignature(pubKey, signature, signedMessage)
		if err != nil || !valid {
			log.Error("btc signature validation failed",
				logger.Attrs{"txid": tx.Transaction.Txid, "err": err, "sigType": msg.SignatureType,
					"pubKey": pubKey, "signature": signature,
					"message": signedMessage})
			return
		}
	case oipProto.SignatureTypes_Flo:
		valid, err := flo.CheckSignature(pubKey, signature, signedMessage)
		if err != nil || !valid {
			log.Error("flo signature validation failed",
				logger.Attrs{"txid": tx.Transaction.Txid, "err": err, "sigType": msg.SignatureType,
					"pubKey": pubKey, "signature": signature,
					"message": signedMessage})
			return
		}
	default:
		log.Error("unsupported proto signature type",
			logger.Attrs{"txid": tx.Transaction.Txid, "err": err, "sigType": msg.SignatureType})
		return
	}

	switch msg.MessageType {
	case oipProto.MessageTypes_Historian:
		var hdp = &oipProto.HistorianDataPoint{}
		err = proto.Unmarshal(msg.SerializedMessage, hdp)
		if err != nil {
			log.Error("unable to unmarshal protobuf historian message",
				logger.Attrs{"txid": tx.Transaction.Txid, "p64": p64, "err": err})
			return
		}
		events.Bus.Publish("modules:historian:protoDataPoint", hdp, tx)
	case oipProto.MessageTypes_OIP05:
		// ToDo
		log.Info("unexpected OIP 0.5 message", logger.Attrs{"txid": tx.Transaction.Txid})
	default:
		log.Error("unsupported proto message type",
			logger.Attrs{"txid": tx.Transaction.Txid, "err": err, "msgType": msg.MessageType})
	}
}
