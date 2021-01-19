package sync

import (
	"context"
	"time"

	"github.com/azer/logger"
	"github.com/bitspill/flod/chaincfg/chainhash"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/flo"
)

func InitialSync(ctx context.Context, count int64) (datastore.BlockData, error) {
	var lb datastore.BlockData
	var lbh int64 = -1

	lb, err := datastore.GetLastBlock(ctx)
	if err != nil {
		log.Error("Get last block failed", logger.Attrs{"err": err})
		return lb, err
	}

	recentBlocks.Push(&lb)

	if lb.Block != nil {
		lbh = lb.Block.Height

		hash, err := flo.GetBlockHash(lb.Block.Height)
		if err != nil {
			log.Error("Get block hash failed", logger.Attrs{"err": err, "height": lb.Block.Height})
			return lb, err
		}

		hash2, err := chainhash.NewHashFromStr(lb.Block.Hash)
		if err != nil {
			log.Error("hash from string failed", logger.Attrs{"err": err, "hash": lb.Block.Hash})
			return lb, err
		}
		if !hash.IsEqual(hash2) {
			log.Error("database and blockchain hash mismatch", logger.Attrs{"err": err, "floHash": hash, "dbHash": hash2, "height": lb.Block.Height})
			return lb, errors.Errorf("initialSync: Database and Blockchain hash mismatch %s != %s", hash, hash2)
		}
	}

	startup := time.Now()
	totalEstimatedSize := int64(0)

	Setup()

	for nh := lbh + 1; nh <= count; nh++ {
		if ctx.Err() != nil {
			log.Error("context error", logger.Attrs{"err": ctx.Err()})
			break
		}

		nlb, err := IndexBlockAtHeight(nh, lb)
		if err != nil {
			return lb, err
		}
		lb = nlb

		bir, err := datastore.AutoBulk.CheckSizeStore(ctx)
		if err != nil {
			return lb, err
		}

		if bir.Stored {
			totalEstimatedSize += bir.EstimatedSize
		}

		if nh%1000 == 0 {
			log.Info("Sync currently at height %s (%s) %s elapsed", humanize.Comma(nh), time.Unix(lb.Block.Time, 0), time.Since(startup))
		}
	}

	estimatedSize := datastore.AutoBulk.EstimateSizeInBytes()
	totalEstimatedSize += estimatedSize
	log.Info("Indexing blocks/transactions", logger.Attrs{"human": humanize.Bytes(uint64(estimatedSize)), "bytes": estimatedSize})

	if datastore.AutoBulk.NumberOfActions() > 0 {
		t := log.Timer()
		br, err := datastore.AutoBulk.Do(ctx)
		if err != nil {
			return lb, err
		}

		t.End("Indexed blocks/transactions", logger.Attrs{"items": len(br.Items), "took": br.Took, "errors": br.Errors})
		if br.Errors {
			log.Error("encountered errors, seeking")
			for _, item := range br.Items {
				for _, value := range item {
					if value.Error != nil {
						log.Error("error executing bulk action", logger.Attrs{
							"index":  value.Index,
							"id":     value.Id,
							"reason": value.Error.Reason,
							"error":  value.Error,
							// "errDump": spew.Sdump(err)
						})
					}
				}
			}
		}
	}

	end := time.Now()
	log.Info("Completed full sync of %s blocks ~%s of block/transaction data in %s",
		humanize.Comma(count), humanize.Bytes(uint64(totalEstimatedSize)), end.Sub(startup))

	return lb, nil
}
