package sync

import (
	"github.com/azer/logger"
	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/flod/wire"
	"github.com/bitspill/floutil"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
)

func init() {
	log.Info("Subscribing to events")
	events.Bus.SubscribeAsync("flo:notify:onFilteredBlockConnected", onFilteredBlockConnected, true)
	events.Bus.SubscribeAsync("flo:notify:onTxAcceptedVerbose", onTxAcceptedVerbose, false)

}

func onFilteredBlockConnected(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {
	log.Info("BlockConnected", logger.Attrs{"height": height})
	// ToDo: manage ilb properly
	// ToDo: check missed blocks between sync end and first notification
	// ToDo: commit each new block when live

	if ilb.Block.Hash != header.PrevBlock.String() {
		// ToDo: handle error regarding last/prev block hash mismatch
		log.Error("incoming block does not follow last block", &logger.Attrs{"incomingHash": header.PrevBlock.String(), "lastHash": ilb.Block.Hash, "incomingHeight": height, "lastHeight": ilb.Block.Height})
		return
	}

	_, err := IndexBlockAtHeight(int64(height), ilb)
	if err != nil {
	}
}

func onTxAcceptedVerbose(txDetails *flojson.TxRawResult) {
	tx := datastore.TransactionData{
		Block:       -1,
		BlockHash:   "",
		Confirmed:   false,
		Transaction: *txDetails,
	}

	datastore.AutoBulk.StoreTransaction(tx)
	if len(tx.Transaction.FloData) != 0 {
		events.Bus.Publish("flo:floData", tx.Transaction.FloData, tx)
	}
}
