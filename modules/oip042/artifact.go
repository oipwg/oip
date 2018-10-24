package oip042

import (
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/flo"
	"github.com/json-iterator/go"
	"gopkg.in/olivere/elastic.v6"
)

func on42JsonPublishArtifact(artifact jsoniter.Any, tx *datastore.TransactionData) {
	title := artifact.Get("info", "title").ToString()
	if len(title) == 0 {
		log.Error("oip042 no title", logger.Attrs{"txid": tx.Transaction.Txid})
		return
	}

	floAddr := artifact.Get("floAddress").ToString()
	ok, err := flo.CheckAddress(floAddr)
	if !ok {
		log.Error("invalid FLO address", logger.Attrs{"txid": tx.Transaction.Txid, "err": err})
		return
	}

	v := []string{artifact.Get("storage", "location").ToString(), floAddr,
		strconv.FormatInt(artifact.Get("timestamp").ToInt64(), 10)}
	preImage := strings.Join(v, "-")

	sig := artifact.Get("signature").ToString()
	ok, err = flo.CheckSignature(floAddr, sig, preImage)
	if !ok {
		log.Error("invalid signature", logger.Attrs{"txid": tx.Transaction.Txid, "preimage": preImage,
			"address": floAddr, "sig": sig, "err": err})
		return
	}

	var el elasticOip042Artifact
	el.Artifact = artifact.GetInterface()
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

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonEditArtifact(any jsoniter.Any, tx *datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonEditArtifact", logger.Attrs{"txid": tx.Transaction.Txid})

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
		Type:      "artifact",
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042EditIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonTransferArtifact(any jsoniter.Any, tx *datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonTransferArtifact", logger.Attrs{"txid": tx.Transaction.Txid})

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
		Type:      "artifact",
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042TransferIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

func on42JsonDeactivateArtifact(any jsoniter.Any, tx *datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonDeactivateArtifact", logger.Attrs{"txid": tx.Transaction.Txid})

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
		Type:      "artifact",
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042DeactivateIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}
