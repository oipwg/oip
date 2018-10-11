package oip

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/flo"
	"github.com/bitspill/oip/httpapi"
	oipSync "github.com/bitspill/oip/sync"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

const multipartIndex = "oip-multipart-single"

var multiPartCommitMutex sync.Mutex
var mpRouter = httpapi.NewSubRoute("/multipart")

func init() {
	log.Info("init multipart")
	datastore.RegisterMapping(multipartIndex, multipartMapping)
	events.Bus.SubscribeAsync("modules:oip:multipartSingle", onMultipartSingle, false)
	events.Bus.SubscribeAsync("datastore:commit", onDatastoreCommit, false)

	mpRouter.HandleFunc("/get/ref/{ref:[a-f0-9]+}/{limit:[0-9]+}", handleGetRef)
	mpRouter.HandleFunc("/get/ref/{ref:[a-f0-9]+}", handleGetRef)
	mpRouter.HandleFunc("/get/id/{id:[a-f0-9]+}", handleGetId)
}

func handleGetId(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	// fsc := elastic.NewFetchSourceContext(true).
	// 	Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(multipartIndex).
		Type("_doc").
		Query(q).
		Size(1).
		Sort("meta.time", false).
		// FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		httpapi.RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}

func handleGetRef(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	limit, _ := opts["limit"]
	size, _ := strconv.ParseInt(limit, 10, 0)
	if size <= 0 || size > 1000 {
		size = -1
	}

	q := elastic.NewBoolQuery().Must(
		elastic.NewPrefixQuery("reference", opts["ref"]),
	)

	// fsc := elastic.NewFetchSourceContext(true).
	// 	Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(multipartIndex).
		Type("_doc").
		Query(q).
		Size(int(size)).
		Sort("meta.time", false).
		// FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		httpapi.RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		// "limit": size,
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}

func onDatastoreCommit() {
	multiPartCommitMutex.Lock()
	defer multiPartCommitMutex.Unlock()

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.complete", false),
		elastic.NewTermQuery("meta.stale", false),
	)
	results, err := datastore.Client().Search(multipartIndex).Type("_doc").Query(q).Size(10000).Sort("meta.time", false).Do(context.TODO())
	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		return
	}

	log.Info("Collecting multiparts to attempt assembly", logger.Attrs{"pendingParts": len(results.Hits.Hits)})

	multiparts := make(map[string]Multipart)
	for _, v := range results.Hits.Hits {
		var mps MultipartSingle
		err := json.Unmarshal(*v.Source, &mps)
		if err != nil {
			log.Info("failed to unmarshal elastic hit", logger.Attrs{"err": err})
			continue
		}
		mp, ok := multiparts[mps.Reference]
		if !ok {
			mp.Total = mps.Max + 1
		}
		if mps.Part < mp.Total {
			mp.Count++
			mp.Parts = append(mp.Parts, mps)
			multiparts[mps.Reference] = mp
		}
	}

	potentialChanges := false
	for k, mp := range multiparts {
		if mp.Count >= mp.Total {
			if mp.Count > mp.Total {
				log.Info("extra parts", k)
			}
			tryCompleteMultipart(mp)
			potentialChanges = true
		}
	}

	if potentialChanges {
		ref, err := datastore.Client().Refresh(multipartIndex).Do(context.TODO())
		if err != nil {
			log.Info("multipart refresh failed")
			spew.Dump(err)
		} else {
			tot := ref.Shards.Total
			fai := ref.Shards.Failed
			suc := ref.Shards.Successful
			log.Info("refresh complete", logger.Attrs{"total": tot, "failed": fai, "successful": suc})
		}

		events.Bus.Publish("modules:oip:mpCompleted")
	}

	if !oipSync.IsInitialSync {
		markStale()
	}
}

func tryCompleteMultipart(mp Multipart) {
	rebuild := make([]string, mp.Total)
	var part0 MultipartSingle
	for i := range mp.Parts {
		value := &mp.Parts[i]
		if value.Part == 0 {
			part0 = *value
		}
		if rebuild[value.Part] != "" {
			log.Info("dupe", value.Meta.Txid)
		}
		rebuild[value.Part] = value.Data
	}

	for _, v := range rebuild {
		if v == "" {
			return
		}
	}

	log.Info("completed mp ", logger.Attrs{"reference": mp.Parts[0].Reference})

	dataString := strings.Join(rebuild, "")
	s := elastic.NewScript("ctx._source.meta.complete=true;"+
		"ctx._source.meta.assembled=params.assembled").Type("inline").Param("assembled", dataString).Lang("painless")

	q := elastic.NewTermQuery("reference", part0.Reference)
	cuq := datastore.Client().UpdateByQuery(multipartIndex).Query(q).
		Type("_doc").Script(s)

	// elastic.NewBulkUpdateRequest()

	res, err := cuq.Do(context.TODO())

	if err != nil {
		log.Error("error updating multipart", logger.Attrs{
			"reference": part0.Reference,
			"block":     part0.Meta.Block,
			"err":       err,
			"errDump":   spew.Sdump(err)})
		return
	}

	events.Bus.Publish("flo:floData", dataString, part0.Meta.Tx)

	log.Info("marked as completed", logger.Attrs{"reference": part0.Reference, "updated": res.Updated, "took": res.Took})
}

func onMultipartSingle(floData string, tx datastore.TransactionData) {
	ms, err := multipartSingleFromString(floData)
	if err != nil {
		log.Info("multipartSingleFromString error", logger.Attrs{"err": err, "txid": tx.Transaction.Txid})
		return
	}

	if ms.Part == 0 {
		ms.Reference = tx.Transaction.Txid[0:10]
	} else {
		ms.Reference = ms.Reference[0:10]
	}

	ms.Meta = MSMeta{
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Complete:  false,
		Time:      tx.Transaction.Time,
		Txid:      tx.Transaction.Txid,
		Tx:        tx,
	}

	bir := elastic.NewBulkIndexRequest().Index(multipartIndex).Type("_doc").Doc(ms).Id(tx.Transaction.Txid)
	datastore.AutoBulk.Add(bir)
}

func multipartSingleFromString(s string) (MultipartSingle, error) {
	var ret MultipartSingle

	// trim prefix off
	s = strings.TrimPrefix(s, "alexandria-media-multipart(")
	s = strings.TrimPrefix(s, "oip-mp(")

	comChunks := strings.Split(s, "):")
	if len(comChunks) < 2 {
		return ret, errors.New("malformed multi-part")
	}

	metaString := comChunks[0]
	dataString := strings.Join(comChunks[1:], "):")

	meta := strings.Split(metaString, ",")
	lm := len(meta)
	// 4 if omitting reference, 5 with all fields, 6 if erroneous fluffy-enigma trailing comma
	if lm != 4 && lm != 5 && lm != 6 {
		return ret, errors.New("malformed multi-part meta")
	}

	// check part and max
	partS := meta[0]
	part, err := strconv.Atoi(partS)
	if err != nil {
		return ret, errors.New("cannot convert part to int")
	}
	maxS := meta[1]
	max, err2 := strconv.Atoi(maxS)
	if err2 != nil {
		return ret, errors.New("cannot convert max to int")
	}

	if max <= 0 {
		return ret, errors.New("max must be positive")
	}

	if part > max {
		return ret, errors.New("part must not exceed max")
	}

	// get and check address
	address := meta[2]
	if ok, err := flo.CheckAddress(address); !ok {
		return ret, errors.Wrap(err, "ErrInvalidAddress")
	}

	reference := meta[3]
	signature := meta[lm-1]
	if signature == "" {
		// fluffy-enigma for a while appended an erroneous trailing comma
		signature = meta[lm-2]
	}

	// signature pre-image is <part>-<max>-<address>-<txid>-<data>
	// in the case of multipart[0], txid is 64 zeros
	// in the case of multipart[n], where n != 0, txid is the reference txid (from multipart[0])
	preimage := partS + "-" + maxS + "-" + address + "-" + reference + "-" + dataString

	if ok, err := flo.CheckSignature(address, signature, preimage); !ok {
		if part != 0 {
			return ret, errors.Wrap(err, "ErrBadSignature")
		}
		preimage := partS + "-" + maxS + "-" + address + "-" + strings.Repeat("0", 64) + "-" + dataString
		if ok, err := flo.CheckSignature(address, signature, preimage); !ok {
			return ret, errors.Wrap(err, "ErrBadSignature")
		}
	}

	if max == 0 {
		panic(s)
	}

	ret = MultipartSingle{
		Part:      part,
		Max:       max,
		Reference: reference,
		Address:   address,
		Signature: signature,
		Data:      dataString,
	}

	return ret, nil
}

func markStale() {
	s := elastic.NewScript("ctx._source.meta.stale=true;").Type("inline").Lang("painless")

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.complete", false),
		elastic.NewTermQuery("meta.stale", false),
		elastic.NewRangeQuery("meta.time").Lte("now-1w"),
	)
	cuq := datastore.Client().UpdateByQuery(multipartIndex).Query(q).
		Type("_doc").Script(s) // .Refresh("wait_for")

	res, err := cuq.Do(context.TODO())
	if err != nil {
		spew.Dump(err)
		panic("")
	}
	log.Info("mark stale complete", logger.Attrs{"total": res.Total, "took": res.Took, "updated": res.Updated})
}

type MultipartSingle struct {
	Part      int    `json:"part"`
	Max       int    `json:"max"`
	Reference string `json:"reference"`
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Data      string `json:"data"`
	Meta      MSMeta `json:"meta"`
}

type MSMeta struct {
	Block     int64                     `json:"block"`
	BlockHash string                    `json:"block_hash"`
	Complete  bool                      `json:"complete"`
	Stale     bool                      `json:"stale"`
	Txid      string                    `json:"txid"`
	Time      int64                     `json:"time"`
	Tx        datastore.TransactionData `json:"tx"`
}

type Multipart struct {
	Parts []MultipartSingle
	Count int
	Total int
}

const multipartMapping = `{
  "settings": {
    "number_of_shards": 2
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "address": {
          "type": "keyword",
          "ignore_above": 36
        },
        "data": {
          "type": "text",
          "index": false
        },
        "max": {
          "type": "long"
        },
        "meta": {
          "properties": {
            "assembled": {
              "type": "text",
              "index": false
            },
            "block": {
              "type": "long"
            },
            "block_hash": {
              "type": "keyword",
              "ignore_above": 64
            },
            "complete": {
              "type": "boolean"
            },
            "stale": {
              "type": "boolean"
            },
            "time": {
              "type": "date",
              "format": "epoch_second"
            },
            "txid": {
              "type": "keyword",
              "ignore_above": 64
            },
            "tx": {
              "type": "object",
              "enabled": false
            }
          }
        },
        "part": {
          "type": "long"
        },
        "reference": {
          "type": "keyword",
          "ignore_above": 64
        },
        "signature": {
          "type": "text",
          "index": false
        },
        "txid": {
          "type": "keyword",
          "ignore_above": 64
        }
      }
    }
  }
}`
