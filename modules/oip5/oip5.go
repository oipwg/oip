package oip5

import (
	"github.com/azer/logger"
	"github.com/golang/protobuf/proto"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/modules/oip"
)

func init() {
	log.Info("init oip5")
	events.SubscribeAsync("modules:oip5:msg", on5msg)

	datastore.RegisterMapping("oip5_templates", "oip5_templates.json")
	datastore.RegisterMapping("oip5_record", "oip5_record.json")
}

func on5msg(msg oip.SignedMessage, tx *datastore.TransactionData) {
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
		bir, err := intakeRecordTemplate(o5.RecordTemplate, msg.PubKey, tx)
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
		bir, err := intakeRecord(o5.Record, msg.PubKey, tx)
		if err != nil {
			attr["err"] = err
			log.Error("unable to process Record", attr)
		} else {
			attr["deets"] = o5.Record.Details
			log.Info("adding o5 record", attr)
			datastore.AutoBulk.Add(bir)

			events.Publish("modules:oip5:record", o5.Record, msg.PubKey, tx)
		}
	}

	if o5.Normalize != nil {
		nonNilAction = true
		bir, err := intakeNormalize(o5.Normalize, msg.PubKey, tx)
		if err != nil {
			attr["err"] = err
			log.Error("unable to process Normalize", attr)
		} else {
			log.Info("adding o5 normalize", attr)
			datastore.AutoBulk.Add(bir)
		}
	}

	if o5.Edit != nil {
		nonNilAction = true
		bir, err := intakeEdit(o5.Edit, msg.PubKey, tx)
		if err != nil {
			attr["err"] = err
			log.Error("unable to process Edit", attr)
		} else {
			log.Info("adding o5 edit", attr)
			datastore.AutoBulk.Add(bir)
		}
	}

	if !nonNilAction {
		log.Error("no supported oip5 action to process", attr)
	}
}
