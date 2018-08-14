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
	el.Meta = OMeta{
		Time:        tx.Transaction.Time,
		Txid:        tx.Transaction.Txid,
		Signature:   sig,
		BlockHash:   tx.BlockHash,
		Block:       tx.Block,
		Deactivated: false,
		Tx:          tx,
	}

	bir := elastic.NewBulkIndexRequest().Index(oip042ArtifactIndex).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

type elasticOip042Artifact struct {
	Artifact interface{} `json:"artifact"`
	Meta     OMeta       `json:"meta"`
}
type OMeta struct {
	Block       int64                     `json:"block"`
	BlockHash   string                    `json:"block_hash"`
	Deactivated bool                      `json:"deactivated"`
	Signature   string                    `json:"signature"`
	Txid        string                    `json:"txid"`
	Time        int64                     `json:"time"`
	Tx          datastore.TransactionData `json:"tx"`
}
