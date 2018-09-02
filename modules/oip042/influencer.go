package oip042

import (
	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/json-iterator/go"
	"gopkg.in/olivere/elastic.v6"
)

func on42JsonRegisterInfluencer(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonRegisterInfluencer", logger.Attrs{"txid": tx.Transaction.Txid})

	sig := any.Get("signature").ToString()

	var el elasticOip042Influencer
	el.Influencer = any.GetInterface()
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

	bir := elastic.NewBulkIndexRequest().Index(oip042InfluencerIndex).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonEditInfluencer(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonEditInfluencer", logger.Attrs{"txid": tx.Transaction.Txid})

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
		Type:      "influencer",
	}

	bir := elastic.NewBulkIndexRequest().Index(oip042EditIndex).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonTransferInfluencer(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonTransferInfluencer", logger.Attrs{"txid": tx.Transaction.Txid})

	sig := any.Get("signature").ToString()

	var el elasticOip042Transfer
	el.Transfer = any.GetInterface()
	el.Meta = OMeta{
		Time:      tx.Transaction.Time,
		Txid:      tx.Transaction.Txid,
		Signature: sig,
		BlockHash: tx.BlockHash,
		Block:     tx.Block,
		Completed: false,
		Tx:        tx,
		Type:      "influencer",
	}

	bir := elastic.NewBulkIndexRequest().Index(oip042TransferIndex).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonDeactivateInfluencer(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonDeactivateInfluencer", logger.Attrs{"txid": tx.Transaction.Txid})

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
		Type:      "influencer",
	}

	bir := elastic.NewBulkIndexRequest().Index(oip042DeactivateIndex).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}
