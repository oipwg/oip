package sync

import (
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/flo"
)

var (
	rpc *flo.RPC
	ilb datastore.BlockData

	IsInitialSync = true
)

func Setup(r *flo.RPC) {
	rpc = r
}

func IndexBlockAtHeight(height int64, lb datastore.BlockData) (datastore.BlockData, error) {
	hash, err := rpc.GetBlockHash(height)
	if err != nil {
		return lb, err
	}

	b, err := rpc.GetBlockVerboseTx(hash)
	if err != nil {
		return lb, err
	}

	var lbt int64
	if lb.Block == nil {
		lbt = b.Time
	} else {
		lbt = lb.Block.Time
	}

	bd := datastore.BlockData{
		Block:             b,
		SecSinceLastBlock: b.Time - lbt,
	}

	datastore.AutoBulk.StoreBlock(bd)

	for _, v := range bd.Block.RawTx {
		tx := datastore.TransactionData{
			Block:       bd.Block.Height,
			BlockHash:   bd.Block.Hash,
			Confirmed:   true,
			Transaction: v,
		}

		datastore.AutoBulk.StoreTransaction(tx)
		if len(tx.Transaction.FloData) != 0 {
			events.Bus.Publish("flo:floData", tx.Transaction.FloData, tx)
		}
	}
	ilb = bd
	return bd, nil
}
