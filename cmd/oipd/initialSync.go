package main

import (
	"context"
	"time"

	"github.com/azer/logger"
	"github.com/bitspill/flod/chaincfg/chainhash"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/sync"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
)

func InitialSync(ctx context.Context, count int64) (datastore.BlockData, error) {
	var lb datastore.BlockData
	var lbh int64 = -1

	lb, err := datastore.GetLastBlock(ctx)
	if err != nil {
		return lb, err
	}

	if lb.Block != nil {
		lbh = lb.Block.Height

		hash, err := FloRPC.GetBlockHash(lb.Block.Height)
		if err != nil {
			return lb, err
		}

		hash2, err := chainhash.NewHashFromStr(lb.Block.Hash)
		if err != nil {
			return lb, err
		}
		if !hash.IsEqual(hash2) {
			return lb, errors.Errorf("initialSync: Database and Blockchain hash mismatch %s != %s", hash, hash2)
		}
	}

	startup := time.Now()
	totalEstimatedSize := int64(0)

	sync.Setup(&FloRPC)

	for nh := lbh + 1; nh <= count; nh++ {
		if ctx.Err() != nil {
			log.Error("context error", logger.Attrs{"err": ctx.Err()})
			break
		}

		nlb, err := sync.IndexBlockAtHeight(nh, lb)
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

		if nh%10000 == 0 {
			log.Info("Sync currently at height %s %s elapsed", humanize.Comma(nh), time.Now().Sub(startup))
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
	}

	end := time.Now()
	log.Info("Completed full sync of %s blocks ~%s of block/transaction data in %s",
		humanize.Comma(count), humanize.Bytes(uint64(totalEstimatedSize)), end.Sub(startup))

	return lb, nil
}
