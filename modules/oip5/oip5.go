package oip5

import (
	"bytes"
	"encoding/json"

	"github.com/azer/logger"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	_ "github.com/jhump/protoreflect/dynamic"
	"github.com/oipwg/oip/oipProto"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

func init() {
	log.Info("init oip5")
	events.SubscribeAsync("modules:oip5:msg", on5msg, false)

	_ = datastore.RegisterMapping("oip5_templates", "oip5_templates.json")

}

func on5msg(msg oipProto.SignedMessage, tx *datastore.TransactionData) {
	attr := logger.Attrs{"txid": tx.Transaction.Txid}
	log.Info("oip5 ", attr)

	var o5 = &OipFive{}

	err := proto.Unmarshal(msg.SerializedMessage, o5)
	if err != nil {
		attr["err"] = err
		log.Error("unable to unmarshal serialized message", attr)
		return
	}

	nonNilAction := false
	if o5.RecordTemplate != nil {
		nonNilAction = true
		bir, err := intakeRecordTemplate(o5.RecordTemplate, tx)
		if err != nil {
			attr["err"] = err
			log.Error("unable to process RecordTemplate", attr)
		} else {
			attr["templateName"] = o5.RecordTemplate.FriendlyName
			log.Info("adding RecordTemplate", attr)
			datastore.AutoBulk.Add(bir)
		}
	}

	if o5.Record != nil {
		nonNilAction = true
		bir, err := intakeRecord(o5.Record, tx)
		if err != nil {
			attr["err"] = err
			log.Error("unable to process Record", attr)
		} else {
			datastore.AutoBulk.Add(bir)
		}
	}

	if !nonNilAction {
		log.Error("no supported oip5 action to process", attr)
	}
}

func intakeRecord(r *RecordProto, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	m := jsonpb.Marshaler{}

	var buf bytes.Buffer
	err := m.Marshal(&buf, r)
	if err != nil {
		return nil, err
	}

	var el elasticOip5Record
	el.Record = buf.Bytes()
	el.Meta = RMeta{
		Block:       tx.Block,
		BlockHash:   tx.BlockHash,
		Deactivated: false,
		Time:        tx.Transaction.Time,
		Tx:          tx,
		Txid:        tx.Transaction.Txid,
		Type:        "oip5",
	}

	bir := elastic.NewBulkIndexRequest().
		Index(datastore.Index("oip5_record")).
		Type("_doc").
		Id(tx.Transaction.Txid).
		Doc(el)
	return bir, nil
}

type elasticOip5Record struct {
	Record json.RawMessage `json:"record"`
	Meta   RMeta           `json:"meta"`
}

type RMeta struct {
	Block       int64                      `json:"block"`
	BlockHash   string                     `json:"block_hash"`
	Deactivated bool                       `json:"deactivated"`
	Time        int64                      `json:"time"`
	Tx          *datastore.TransactionData `json:"tx"`
	Txid        string                     `json:"txid"`
	Type        string                     `json:"type"`
}
