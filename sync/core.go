package sync

import (
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/flo"
)

var (
	IsInitialSync = true
	MultipartSyncComplete = false
	EditSyncComplete = false
	recentBlocks  = blockBuffer{}
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

	var lbt int64
	if lb.Block == nil {
		lbt = b.Time
	} else {
		lbt = lb.Block.Time
	}

	bd := datastore.BlockData{
		Block:             b,
		SecSinceLastBlock: b.Time - lbt,
		Orphaned:          false,
	}

	datastore.AutoBulk.StoreBlock(bd)

	for i := range bd.Block.RawTx {
		rawTx := &bd.Block.RawTx[i]
		tx := &datastore.TransactionData{
			Block:       bd.Block.Height,
			BlockHash:   bd.Block.Hash,
			Confirmed:   true,
			IsCoinbase:  rawTx.Vin[0].IsCoinBase(),
			Transaction: rawTx,
		}

		datastore.AutoBulk.StoreTransaction(tx)
		if len(tx.Transaction.FloData) != 0 {
			events.Publish("flo:floData", tx.Transaction.FloData, tx)
		}
	}
	recentBlocks.Push(&bd)
	return bd, nil
}
