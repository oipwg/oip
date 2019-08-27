package datastore

import (
	"context"
	"sync"
	"time"

	"github.com/azer/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/dustin/go-humanize"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/events"
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
		log.Info("Quick Indexing %v of data containing %d blocks/transactions (%v bytes)", humanize.Bytes(uint64(estimatedSize)), bi.NumberOfActions(), estimatedSize)

		br, err := bi.Do(context.TODO())
		if err != nil {
			log.Error("Error commiting to ES in quickCommit", logger.Attrs{"err": spew.Sdump(err)})
			return
		}

		t.End("Quick Indexed %d blocks & transactions, took %v (errors=%v)", len(br.Items), br.Took, br.Errors)
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

func (bi *BulkIndexer) OrphanBlock(hash string) {
	// Orphan Block
	bir := elastic.NewBulkUpdateRequest().
		Index(Index("blocks")).
		Type("_doc").
		Id(hash).
		Doc(struct {
			Orphaned bool `json:"orphaned"`
		}{
			Orphaned: true,
		})
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
	_, err := bi.CheckSizeStore(context.TODO())
	if err != nil {
		log.Error("error storing after add")
	}
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

	// https://www.elastic.co/guide/en/elasticsearch/guide/2.x/indexing-performance.html#_using_and_sizing_bulk_requests
	// > Bulk sizing is dependent on your data, analysis, and cluster configuration, but a good starting point is 5–15 MB per bulk
	// https://www.elastic.co/guide/en/elasticsearch/reference/master/tune-for-indexing-speed.html#_use_bulk_requests
	// > it is advisable to avoid going beyond a couple tens of megabytes per request even if larger requests seem to perform better.
	// Set to 10mb to straddle between recomended amounts -skyoung
	if estimatedSize > 10*humanize.MByte {
		log.Info("Bulk Indexing %s of data, %d bulk actions", humanize.Bytes(uint64(estimatedSize)), bi.NumberOfActions())
		t := log.Timer()
		br, err := bi.Do(ctx)
		if err != nil {
			// sky todo, test this error logging, see if this is where the error is coming from
			log.Error("error on bulk indexing!", logger.Attrs{
				"spewError": spew.Sdump(err),
			})
			return BulkIndexerResponse{}, err
		}

		t.End("Bulk Indexed %d blocks & transactions, took %v (errors=%v)", len(br.Items), br.Took, br.Errors)

		if br.Errors {
			log.Error("Encountered errors during Bulk Action Processing!")
			for _, item := range br.Items {
				for _, value := range item {
					if value.Error != nil {
						log.Error("Error executing bulk action in index `%v` for ID `%v`! Error: `%v`",
							value.Index,
							value.Id,
							value.Error,
						)
					}
				}
			}
		}

		// Bulk request actions should get cleared
		if bi.NumberOfActions() > 0 {
			log.Error("Error Bulk Indexing, number of actions has not been cleared to 0! Remaining Actions: %d", bi.NumberOfActions())
		}

		// Check if there were any failed results in the bulk indexing process
		failedResults := br.Failed()
		if len(failedResults) > 0 {
			log.Error("Error, Bulk Indexing had failed %d results!", len(failedResults))
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
