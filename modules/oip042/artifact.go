package oip042

import (
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/json-iterator/go"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/filters"
	"github.com/oipwg/oip/flo"
	"github.com/oipwg/oip/modules/oip042/validators"
	"gopkg.in/olivere/elastic.v6"
)

func on42JsonPublishArtifact(artifact jsoniter.Any, tx *datastore.TransactionData) {
	attr := logger.Attrs{"txid": tx.Transaction.Txid}

	title := artifact.Get("info", "title").ToString()
	if len(title) == 0 {
		log.Error("oip042 no title", attr)
		return
	}

	floAddr := artifact.Get("floAddress").ToString()
	ok, err := flo.CheckAddress(floAddr)
	if !ok {
		attr["err"] = err
		log.Error("invalid FLO address", attr)
		return
	}

	v := []string{artifact.Get("storage", "location").ToString(), floAddr,
		strconv.FormatInt(artifact.Get("timestamp").ToInt64(), 10)}
	preImage := strings.Join(v, "-")

	sig := artifact.Get("signature").ToString()
	ok, err = flo.CheckSignature(floAddr, sig, preImage)
	if !ok {
		attr["err"] = err
		attr["preimage"] = preImage
		attr["address"] = floAddr
		attr["sig"] = sig
		log.Error("invalid signature", attr)
		return
	}

	t := artifact.Get("type").ToString()
	st := artifact.Get("subType").ToString()
	valid := validators.IsValidArtifact(t, st, &artifact, tx.Transaction.Txid)
	if !valid {
		attr["type"] = t
		attr["subtype"] = st
		log.Error("artifact validation failed", attr)
		return
	}

	bl, label := filters.ContainsWithLabel(tx.Transaction.Txid)

	var el elasticOip042Artifact
	el.Artifact = artifact.GetInterface()
	el.Meta = AMeta{
		Block:       tx.Block,
		BlockHash:   tx.BlockHash,
		Blacklist:   Blacklist{Blacklisted: bl, Filter: label},
		Deactivated: false,
		Signature:   sig,
		Time:        tx.Transaction.Time,
		Tx:          tx,
		Txid:        tx.Transaction.Txid,
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
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Completed: false,
		Signature: sig,
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
		OTxid:     any.Get("txid").ToString(),
		PTxid:     "",
		Type:      "artifact",
	}

	el.Patch = any.Get("patch").ToString()
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
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Completed: false,
		Signature: sig,
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
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
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Completed: false,
		Signature: sig,
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
		Type:      "artifact",
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(oip042DeactivateIndex)).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}
