package datastore

import (
	"context"
	"sync"

	"github.com/azer/logger"
	"github.com/dustin/go-humanize"
	"github.com/oipwg/oip/events"
	"gopkg.in/olivere/elastic.v6"
	"time"
)

func BeginBulkIndexer() BulkIndexer {
	bi := BulkIndexer{
		bulk: client.Bulk(),
		m:    &sync.Mutex{},
	}

	return bi
}

type BulkIndexer struct {
	bulk               *elastic.BulkService
	m                  *sync.Mutex
	timedCommitRate    time.Duration
	timedCommitRunning bool
	timedCommitEnd     chan struct{}
}

func (bi *BulkIndexer) BeginTimedCommits(rate time.Duration) {
	bi.timedCommitRate = rate
	if bi.timedCommitRunning {
		return
	}
	go bi.timedCommit()
}

func (bi *BulkIndexer) timedCommit() {
	for {
		select {
		case <-bi.timedCommitEnd:
			bi.quickCommit()
			return
		case <-time.After(bi.timedCommitRate):
			bi.quickCommit()
		}
	}
}

func (bi *BulkIndexer) quickCommit() {
	bi.m.Lock()
	defer bi.m.Unlock()
	if bi.NumberOfActions() > 0 {
		t := log.Timer()
		estimatedSize := bi.EstimateSizeInBytes()
		log.Info("Indexing blocks/transactions", logger.Attrs{"human": humanize.Bytes(uint64(estimatedSize)), "bytes": estimatedSize})

		br, err := bi.Do(context.TODO())
		if err != nil {
			log.Error("error commiting to ES in quickCommit", logger.Attrs{"err": err})
			return
		}

		t.End("Indexed blocks/transactions", logger.Attrs{"items": len(br.Items), "took": br.Took, "errors": br.Errors})
	}
}

func (bi *BulkIndexer) EndTimedCommit() {
	bi.timedCommitEnd <- struct{}{}
}

func (bi *BulkIndexer) Do(ctx context.Context) (*elastic.BulkResponse, error) {
	br, err := bi.bulk.Do(ctx)
	if err == nil {
		events.Publish("datastore:commit")
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
		Index(Index("blocks")).
		Type("_doc").
		Id(bd.Block.Hash).
		Doc(bd)
	bi.Add(bir)
}

func (bi *BulkIndexer) DeleteBlock(hash string) {
	bir := elastic.NewBulkDeleteRequest().
		Index(Index("blocks")).
		Type("_doc").
		Id(hash)
	bi.Add(bir)
}

func (bi *BulkIndexer) StoreTransaction(td *TransactionData) {
	bir := elastic.NewBulkIndexRequest().
		Index(Index("transactions")).
		Type("_doc").
		Id(td.Transaction.Txid).
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
