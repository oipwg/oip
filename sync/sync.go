package sync

import (
	"context"

	"github.com/azer/logger"
	"github.com/bitspill/flod/wire"
	"github.com/bitspill/floutil"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/dustin/go-humanize"
)

func init() {
	log.Info("Subscribing to events")
	events.Bus.SubscribeAsync("flo:notify:onFilteredBlockConnected", onFilteredBlockConnected, true)
}

func onFilteredBlockConnected(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {
	log.Info("BlockConnected", logger.Attrs{"height": height})
	// ToDo: manage ilb properly
	// ToDo: check missed blocks between sync end and first notification
	// ToDo: commit each new block when live
	_, _ = IndexBlockAtHeight(int64(height), ilb)

	estimatedSize := datastore.AutoBulk.EstimateSizeInBytes()
	log.Info("Indexing blocks/transactions", logger.Attrs{"human": humanize.Bytes(uint64(estimatedSize)), "bytes": estimatedSize})

	if datastore.AutoBulk.NumberOfActions() > 0 {
		t := log.Timer()
		br, err := datastore.AutoBulk.Do(context.TODO())
		if err != nil {
			return
		}

		t.End("Indexed blocks/transactions", logger.Attrs{"items": len(br.Items), "took": br.Took, "errors": br.Errors})
	}
}
