package sync

import (
	"github.com/azer/logger"
	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/flod/wire"
	"github.com/bitspill/floutil"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

func init() {
	log.Info("Subscribing to events")
	events.SubscribeAsync("flo:notify:onFilteredBlockConnected", onFilteredBlockConnected, false)
	events.SubscribeAsync("flo:notify:onFilteredBlockDisconnected", onFilteredBlockDisconnected, true)
	events.SubscribeAsync("flo:notify:onTxAcceptedVerbose", onTxAcceptedVerbose, false)
}

func onFilteredBlockConnected(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {
	headerHash := header.BlockHash().String()

	attr := logger.Attrs{ "iHeight": height, "iHash": headerHash }

	//log.Info("Incoming Block: %v (%d) %v", headerHash, height, header.Timestamp)

	lastBlock := recentBlocks.PeekFront()

	if lastBlock.Block.Hash == header.PrevBlock.String() {
		// easy case new block follows; add it
		_, err := IndexBlockAtHeight(int64(height), *lastBlock)
		if err != nil {
			attr["err"] = err
			log.Error("onFilteredBlockConnected unable to index block, follow", attr)
		}

		log.Info("Indexed Block:  %v (%d) %v", headerHash, height, header.Timestamp)

		return
	}

	// more difficult cases; new block does not follow
	// maybe orphan, fork, or future block
	attr["lHash"] = lastBlock.Block.Hash
	attr["lHeight"] = lastBlock.Block.Height

	if int64(height) > lastBlock.Block.Height+1 {
		log.Info("Incoming Block %v (%d) leaves a gap, syncing missing blocks %d to %d", headerHash, height, lastBlock.Block.Height+1, height-1)

		for missingBlockHeight := lastBlock.Block.Height + 1; missingBlockHeight < int64(height); missingBlockHeight++ {
			//log.Info("Requesting Gap Block at Height %d | Last Block: %v (%d)", missingBlockHeight, lastBlock.Block.Hash, lastBlock.Block.Height)

			nlb, err := IndexBlockAtHeight(int64(missingBlockHeight), *lastBlock)
			if err != nil {
				attr["err"] = err
				log.Error("onFilteredBlockConnected unable to index block, gap", attr)
				return
			}
			lastBlock = &nlb

			log.Info("Indexed Gap Block: %v (%d) | Head Block: %v (%d)", nlb.Block.Hash, nlb.Block.Height, headerHash, height)
		}

		_, err := IndexBlockAtHeight(int64(height), *lastBlock)
		if err != nil {
			attr["err"] = err
			log.Error("onFilteredBlockConnected unable to index block, follow", attr)
		}

		log.Info("Indexed Block:  %v (%d) %v", headerHash, height, header.Timestamp)

		return
	}
}

func onTxAcceptedVerbose(txDetails *flojson.TxRawResult) {
	tx := &datastore.TransactionData{
		Block:       -1,
		BlockHash:   "",
		Confirmed:   false,
		IsCoinbase:  txDetails.Vin[0].IsCoinBase(),
		Transaction: txDetails,
	}

	datastore.AutoBulk.StoreTransaction(tx)
	if len(tx.Transaction.FloData) != 0 {
		events.Publish("flo:floData", tx.Transaction.FloData, tx)
	}
}

func onFilteredBlockDisconnected(height int32, header *wire.BlockHeader) {
	// Because of how this code works, the subscription needs to be Serial (in order, one after the other)
	if recentBlocks.PeekFront().Block.Hash == header.BlockHash().String() {
		nlb := recentBlocks.QuickPopFront()
		log.Info("Disconnected Block: %v (%d) | New Head Block: %v (%d)", header.BlockHash().String(), height, nlb.Block.Hash, nlb.Block.Height)
	}

	// Mark the blocks as orphaned and send off the update requests
	datastore.AutoBulk.OrphanBlock(header.BlockHash().String())
	// Todo orphan transactions, artifacts, multiparts, etc
	log.Info("Marked Block as Orphaned: %v (%d) %v", header.BlockHash().String(), height, header.Timestamp)
}
