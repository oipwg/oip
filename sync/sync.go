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
	events.SubscribeAsync("flo:notify:onFilteredBlockConnected", onFilteredBlockConnected, true)
	events.SubscribeAsync("flo:notify:onFilteredBlockDisconnected", onFilteredBlockDisconnected, true)
	events.SubscribeAsync("flo:notify:onTxAcceptedVerbose", onTxAcceptedVerbose, false)
}

func onFilteredBlockConnected(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {
	attr := logger.Attrs{"incomingHeight": height}

	log.Info("BlockConnected", attr)

	ilb := recentBlocks.PeekFront()

	if ilb.Block.Hash == header.PrevBlock.String() {
		// easy case new block follows; add it
		_, err := IndexBlockAtHeight(int64(height), *ilb)
		if err != nil {
			attr["err"] = err
			log.Error("onFilteredBlockConnected unable to index block, follow", attr)
		}

		return
	}

	// more difficult cases; new block does not follow
	// maybe orphan, fork, or future block

	attr["incomingHash"] = header.PrevBlock.String()
	attr["lastHash"] = ilb.Block.Hash
	attr["lastHeight"] = ilb.Block.Height

	if int64(height) > ilb.Block.Height+1 {
		log.Info("gap in block heights syncing...", attr)

		for i := ilb.Block.Height + 1; i <= int64(height); i++ {
			attr["i"] = i
			attr["lastHash"] = ilb.Block.Hash
			attr["lastHeight"] = ilb.Block.Height
			log.Info("filling gap", attr)
			nlb, err := IndexBlockAtHeight(int64(i), *ilb)
			if err != nil {
				attr["err"] = err
				log.Error("onFilteredBlockConnected unable to index block, gap", attr)
				return
			}
			ilb = &nlb
		}
		return
	}

	// ToDo: test rewind/re-org
	attr["recentBlocksLen"] = recentBlocks.Len()
	for i := -1; i > -recentBlocks.Len(); i-- {
		b := recentBlocks.Get(i)
		attr["i"] = i
		log.Info("unrolling block", attr)
		if b.Block.Hash == header.PrevBlock.String() {
			attr["rewind"] = -i
			log.Info("re-org detected", attr)
			for ; i < 0; i++ {
				attr["pop"] = i
				log.Info("popping block", attr)
				recentBlocks.PopFront()
			}
			_, err := IndexBlockAtHeight(int64(height), *ilb)
			if err != nil {
				attr["err"] = err
				log.Error("onFilteredBlockConnected unable to index block, re-org", attr)
			}

			return
		}
	}

	log.Error("potential fork, unable to connect block", attr)
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
	attr := logger.Attrs{"oldHeight": height, "oldHash": header.BlockHash().String()}

	log.Info("BlockDisconnected", attr)

	// ToDo mark as disconnected in database along with all associated records
}
