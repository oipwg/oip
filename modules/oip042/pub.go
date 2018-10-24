package oip042

import (
	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/json-iterator/go"
	"gopkg.in/olivere/elastic.v6"
)

func on42JsonRegisterPub(any jsoniter.Any, tx *datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonRegisterPub", logger.Attrs{"txid": tx.Transaction.Txid})

	// name := any.Get("name").ToString()
	// if len(name) == 0 {
	//	log.Println("oip042 no pub.name")
	//	return
	// }

	sig := any.Get("signature").ToString()

	var el elasticOip042Pub
	el.Pub = any.GetInterface()
	el.Meta = AMeta{
		Time:        tx.Transaction.Time,
		Txid:        tx.Transaction.Txid,
		Signature:   sig,
		BlockHash:   tx.BlockHash,
		Block:       tx.Block,
		Deactivated: false,
		Tx:          tx,
		Type:        "oip042",
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042PublisherIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonEditPub(any jsoniter.Any, tx *datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonEditPub", logger.Attrs{"txid": tx.Transaction.Txid})

	sig := any.Get("signature").ToString()

	var el elasticOip042Edit
	el.Edit = any.GetInterface()
	el.Meta = OMeta{
		Time:      tx.Transaction.Time,
		Txid:      tx.Transaction.Txid,
		Signature: sig,
		BlockHash: tx.BlockHash,
		Block:     tx.Block,
		Completed: false,
		Tx:        tx,
		Type:      "pub",
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042EditIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonTransferPub(any jsoniter.Any, tx *datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonTransferPub", logger.Attrs{"txid": tx.Transaction.Txid})

	sig := any.Get("signature").ToString()

	var el elasticOip042Edit
	el.Edit = any.GetInterface()
	el.Meta = OMeta{
		Time:      tx.Transaction.Time,
		Txid:      tx.Transaction.Txid,
		Signature: sig,
		BlockHash: tx.BlockHash,
		Block:     tx.Block,
		Completed: false,
		Tx:        tx,
		Type:      "pub",
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042TransferIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonDeactivatePub(any jsoniter.Any, tx *datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonDeactivatePub", logger.Attrs{"txid": tx.Transaction.Txid})

	sig := any.Get("signature").ToString()

	var el elasticOip042DeactivateInterface
	el.Deactivate = any.GetInterface()
	el.Meta = OMeta{
		Time:      tx.Transaction.Time,
		Txid:      tx.Transaction.Txid,
		Signature: sig,
		BlockHash: tx.BlockHash,
		Block:     tx.Block,
		Completed: false,
		Tx:        tx,
		Type:      "pub",
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042DeactivateIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}
