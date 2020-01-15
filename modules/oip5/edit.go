package oip5

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/azer/logger"
	patch "github.com/bitspill/protoPatch"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/oipwg/proto/go/pb_oip"
	"github.com/oipwg/proto/go/pb_oip5"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

var editIndex = "oip5_edit"
var editCommitMutex sync.Mutex

func init() {
	events.SubscribeAsync("datastore:commit", onDatastoreCommitEdits)
	datastore.RegisterMapping("oip5_edit", "oip5_edit.json")
}

func intakeEdit(n *pb_oip5.EditProto, pubKey []byte, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	m := jsonpb.Marshaler{}

	var jsonBuf bytes.Buffer
	err := m.Marshal(&jsonBuf, n)
	if err != nil {
		return nil, err
	}

	fmt.Println(jsonBuf.String())

	rawBuf, err := proto.Marshal(n.Patch)
	if err != nil {
		return nil, err
	}

	rawB64 := base64.StdEncoding.EncodeToString(rawBuf)

	var el elasticOip5Edit
	el.PatchJson = jsonBuf.Bytes()
	el.PatchRaw = rawB64
	el.Reference = pb_oip.TxidToString(n.Reference)
	el.Meta = EMeta{
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Applied:   false,
		Invalid:   false,
		SignedBy:  string(pubKey),
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
	}

	bir := elastic.NewBulkIndexRequest().
		Index(datastore.Index(editIndex)).
		Type("_doc").
		Id(tx.Transaction.Txid).
		Doc(el)

	return bir, nil
}

type elasticOip5Edit struct {
	PatchJson json.RawMessage `json:"patch_json"`
	PatchRaw  string          `json:"patch_raw"`
	Reference string          `json:"reference"`
	Meta      EMeta           `json:"meta"`
}

type EMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	Applied   bool                       `json:"applied"`
	Invalid   bool                       `json:"invalid"`
	SignedBy  string                     `json:"signed_by"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"-"`
	Txid      string                     `json:"txid"`
}

func onDatastoreCommitEdits() {
	editCommitMutex.Lock()
	defer editCommitMutex.Unlock()

	var edits []elasticOip5Edit

	var after []interface{}

moreEdits:
	edits, after, err := queryEdits(edits, after)
	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
	}
	if after != nil {
		goto moreEdits
	}

	for _, edit := range edits {
		log.Info("processing edit", logger.Attrs{"edit": edit.Meta.Txid})

		rec, err := GetRecord(edit.Reference)
		if err != nil {
			markEditInvalid(edit.Meta.Txid)
			log.Error("unable to obtain record for edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
			continue
		}

		b, err := base64.StdEncoding.DecodeString(edit.PatchRaw)
		if err != nil {
			markEditInvalid(edit.Meta.Txid)
			log.Error("unable to decode raw patch", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
			continue
		}

		pp := &patch.ProtoPatch{}
		err = proto.Unmarshal(b, pp)
		if err != nil {
			markEditInvalid(edit.Meta.Txid)
			log.Error("unable to decode proto patch for edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
			continue
		}

		p, err := patch.FromProto(pp)
		if err != nil {
			markEditInvalid(edit.Meta.Txid)
			log.Error("unable to decode patch for edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
			continue
		}

		a, err := patch.ApplyPatch(*p, rec.Record)
		if err != nil {
			markEditInvalid(edit.Meta.Txid)
			log.Error("unable to apply edit to record", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
			continue
		}

		newRec, ok := a.(*pb_oip5.RecordProto)
		if !ok {
			markEditInvalid(edit.Meta.Txid)
			log.Error("patch result is no longer a record", logger.Attrs{"reference": edit.Reference, "txid": edit.Meta.Txid})
			continue
		}

		rec.Record = newRec

		rec.Meta.History = append(rec.Meta.History, edit.Meta.Txid)

		m := jsonpb.Marshaler{}

		var buf bytes.Buffer
		err = m.Marshal(&buf, newRec)
		if err != nil {
			markEditInvalid(edit.Meta.Txid)
			log.Error("unable to marshal record json post edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
			continue
		}

		rawRec, err := proto.Marshal(newRec)
		if err != nil {
			markEditInvalid(edit.Meta.Txid)
			log.Error("unable to marshal record proto post edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
			continue
		}

		rawRec64 := base64.StdEncoding.EncodeToString(rawRec)
		rec.Meta.RecordRaw = rawRec64

		var el elasticOip5Record
		el.Record = buf.Bytes()
		el.Meta = rec.Meta

		bir := elastic.NewBulkIndexRequest().
			Index(datastore.Index("oip5_record")).
			Type("_doc").
			Id(edit.Meta.Txid).
			Doc(el)

		recordCache.Add(rec.Meta.Original, rec)

		datastore.AutoBulk.Add(bir)

		if len(rec.Meta.History) > 1 {
			for _, prevIteration := range rec.Meta.History[:len(rec.Meta.History)-1] {
				bur := elastic.NewBulkUpdateRequest().
					Index(datastore.Index("oip5_record")).
					Type("_doc").
					Id(prevIteration).
					Doc(MetaLatest{Latest{false}})

				datastore.AutoBulk.Add(bur)
			}
		}

		bur := elastic.NewBulkUpdateRequest().
			Index(datastore.Index("oip5_edit")).
			Type("_doc").
			Id(edit.Meta.Txid).
			Doc(MetaApplied{Applied{true}})

		datastore.AutoBulk.Add(bur)
	}
}

func markEditInvalid(txid string) {
	bur := elastic.NewBulkUpdateRequest().
		Index(datastore.Index(editIndex)).
		Type("_doc").
		Id(txid).
		Doc(MetaInvalid{Invalid{true}})

	datastore.AutoBulk.Add(bur)
}

func queryEdits(edits []elasticOip5Edit, after []interface{}) ([]elasticOip5Edit, []interface{}, error) {
	var nextAfter []interface{}
	searchSize := 10000

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.applied", false),
		elastic.NewTermQuery("meta.invalid", false),
	)
	search := datastore.Client().
		Search(datastore.Index(editIndex)).
		Type("_doc").
		Query(q).
		Size(searchSize).
		Sort("meta.time", false).
		Sort("reference", false)

	if after != nil {
		search.SearchAfter(after...)
	}

	results, err := search.Do(context.TODO())
	if err != nil {
		return nil, nil, err
	}

	log.Info("Collecting edits to attempt application", logger.Attrs{"newEdits": len(results.Hits.Hits), "totalEdits": len(results.Hits.Hits) + len(edits)})

	for i, v := range results.Hits.Hits {
		var elEdit elasticOip5Edit
		err := json.Unmarshal(*v.Source, &elEdit)
		if err != nil {
			log.Info("failed to unmarshal elastic hit", logger.Attrs{"err": err})
			continue
		}
		edits = append(edits, elEdit)

		if i == len(results.Hits.Hits)-1 && len(results.Hits.Hits) == searchSize {
			nextAfter = v.Sort
		}
	}

	return edits, nextAfter, nil
}

type Latest struct {
	Latest bool `json:"latest"`
}
type MetaLatest struct {
	Meta Latest `json:"meta"`
}

type Invalid struct {
	Invalid bool `json:"invalid"`
}
type MetaInvalid struct {
	Meta Invalid `json:"meta"`
}

type Applied struct {
	Applied bool `json:"applied"`
}
type MetaApplied struct {
	Meta Applied `json:"meta"`
}
