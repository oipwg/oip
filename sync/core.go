package sync

import (
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/flo"
	"github.com/pkg/errors"
)

var (
	ilb datastore.BlockData

	IsInitialSync = true
)

func Setup() {
}

func IndexBlockAtHeight(height int64, lb datastore.BlockData) (datastore.BlockData, error) {
	hash, err := flo.GetBlockHash(height)
	if err != nil {
		return lb, err
	}

	b, err := flo.GetBlockVerboseTx(hash)
	if err != nil {
		return lb, err
	}

	if lb.Block.Hash != b.PreviousHash {
		return lb, errors.New("block does not follow last known block")
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
