package sync

import (
	goSync "sync"
	"sync/atomic"
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
	events.SubscribeAsync("flo:notify:onFilteredBlockDisconnected", onFilteredBlockDisconnected, false)
	events.SubscribeAsync("flo:notify:onTxAcceptedVerbose", onTxAcceptedVerbose, false)
}

var gapConnecting = false
var onBlockConnectMutex goSync.Mutex
var blocksDisconnecting int32 = 0

func onFilteredBlockConnected(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {
	// Check if we should be waiting for block disconnects to finish processing before connecting new blocks
	if atomic.LoadInt32(&blocksDisconnecting) > 0 {
		// Start the Lock to prevent processing
		onBlockConnectMutex.Lock()
		// Clean up our own Lock
		defer onBlockConnectMutex.Unlock()
	}

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
		// Check if we are currently connecting a gap, and if we are, skip processing the new gap connection since we don't want to mess up recentBlocks
		if gapConnecting {
			return
		}
		gapConnecting = true

		log.Info("Incoming Block  %v (%d) leaves a gap, syncing missing blocks %d to %d", headerHash, height, lastBlock.Block.Height+1, height-1)

		for missingBlockHeight := lastBlock.Block.Height + 1; missingBlockHeight < int64(height); missingBlockHeight++ {
			//log.Info("Requesting Gap Block at Height %d | Last Block: %v (%d)", missingBlockHeight, lastBlock.Block.Hash, lastBlock.Block.Height)
			if recentBlocks.PeekFront().Block.Height >= missingBlockHeight {
				continue
			}

			nlb, err := IndexBlockAtHeight(int64(missingBlockHeight), *lastBlock)
			if err != nil {
				attr["err"] = err
				log.Error("onFilteredBlockConnected unable to index block, gap", attr)

				gapConnecting = false
				return
			}
			lastBlock = &nlb

			log.Info("Indexed Gap Block: %v (%d) | Incoming Head Block: %v (%d)", nlb.Block.Hash, nlb.Block.Height, headerHash, height)
		}

		_, err := IndexBlockAtHeight(int64(height), *lastBlock)
		if err != nil {
			attr["err"] = err
			log.Error("onFilteredBlockConnected unable to index block, follow", attr)
		}

		log.Info("Indexed Block:  %v (%d) %v", headerHash, height, header.Timestamp)

		gapConnecting = false
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
	// Check if we are at the start of a Block Disconnect, and if so, lock processing of new blocks being connected
	if atomic.LoadInt32(&blocksDisconnecting) == 0 {
		// Start the initial lock
		onBlockConnectMutex.Lock()
	}

	// Increment the count of our disconnecting blocks
	atomic.AddInt32(&blocksDisconnecting, 1)

	// Pop the front of recent blocks off. Occasionally block disconnects run out of order, so it is exepected to occasionally see a
	// different popped block than the disconnected block. As long as all disconnects occur before any new connects occur, it is safe
	// to disconnect blocks in this manner
	nlb := recentBlocks.PopFront()
	log.Info("Disconnected Block: %v (%d) | Popped Block: %v (%d)", header.BlockHash().String(), height, nlb.Block.Hash, nlb.Block.Height)

	// Mark the blocks as orphaned and send off the update requests
	datastore.AutoBulk.OrphanBlock(header.BlockHash().String())
	// todo: orphan transactions, artifacts, multiparts, etc
	log.Info("Marked Block as Orphaned: %v (%d) %v", header.BlockHash().String(), height, header.Timestamp)

	// Decrement the count of our disconnecting blocks
	atomic.AddInt32(&blocksDisconnecting, -1)

	// Check if we are done with our Block Disconnects, and if so, clear the lock stopping new blocks from being connected
	if atomic.LoadInt32(&blocksDisconnecting) == 0 {
		// Clear the initial lock
		onBlockConnectMutex.Unlock()
	}
}
