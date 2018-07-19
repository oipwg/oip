package datastore

import (
	"context"
	"sync"

	"github.com/azer/logger"
	"github.com/bitspill/oip/events"
	"github.com/dustin/go-humanize"
	"gopkg.in/olivere/elastic.v6"
)

func BeginBulkIndexer() BulkIndexer {
	bi := BulkIndexer{
		bulk: client.Bulk(),
	}

	return bi
}

type BulkIndexer struct {
	bulk *elastic.BulkService
	m    sync.Mutex
}

func (bi *BulkIndexer) Do(ctx context.Context) (*elastic.BulkResponse, error) {
	br, err := bi.bulk.Do(ctx)
	if err == nil {
		events.Bus.Publish("datastore:commit")
	}
	return br, err
}

func (bi *BulkIndexer) NumberOfActions() int {
	return bi.bulk.NumberOfActions()
}

func (bi *BulkIndexer) EstimateSizeInBytes() int64 {
	return bi.bulk.EstimatedSizeInBytes()
}

func (bi *BulkIndexer) StoreBlock(bd BlockData) {
	bir := elastic.NewBulkIndexRequest().
		Index("blocks").
		Type("_doc").
		Id(bd.Block.Hash).
		Doc(bd)
	bi.Add(bir)
}

func (bi *BulkIndexer) StoreTransaction(td TransactionData) {
	bir := elastic.NewBulkIndexRequest().
		Index("transactions").
		Type("_doc").
		Id(td.Transaction.Hash).
		Doc(td)
	bi.Add(bir)
}

func (bi *BulkIndexer) Add(bir ...elastic.BulkableRequest) {
	bi.m.Lock()
	bi.bulk.Add(bir...)
	bi.m.Unlock()
	bi.CheckSizeStore(context.TODO())
}

type BulkIndexerResponse struct {
	*elastic.BulkResponse
	EstimatedSize int64
	Stored        bool
}

func (bi *BulkIndexer) CheckSizeStore(ctx context.Context) (BulkIndexerResponse, error) {
	bi.m.Lock()
	defer bi.m.Unlock()
	estimatedSize := bi.EstimateSizeInBytes()

	if estimatedSize > 80*humanize.MByte {
		log.Info("Indexing %s of data", humanize.Bytes(uint64(estimatedSize)))
		t := log.Timer()
		br, err := bi.Do(ctx)
		if err != nil {
			return BulkIndexerResponse{}, err
		}

		t.End("Indexed blocks/transactions", logger.Attrs{"items": len(br.Items), "took": br.Took, "errors": br.Errors})

		if br.Errors {
			log.Error("encountered errors, seeking")
			for _, item := range br.Items {
				for _, value := range item {
					if value.Error != nil {
						log.Error("error executing bulk action", logger.Attrs{
							"index": value.Index,
							"error": value.Error,
							//"errDump": spew.Sdump(err)
						})
					}
				}
			}
		}

		// deactivateArtifact: string -- bad
		// b998b28cdbc0b60638df2bbea2997e75937a3115ee8d331ee83d62b538407371

		return BulkIndexerResponse{
			br,
			estimatedSize,
			true,
		}, nil
	}
	return BulkIndexerResponse{}, nil
}
