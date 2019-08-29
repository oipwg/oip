package oip5

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/azer/logger"
	"github.com/bitspill/protoPatch"
	"github.com/golang/protobuf/jsonpb"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/modules/oip"
)

var editIndex = "oip5_edit"
var editCommitMutex sync.Mutex

func init() {
	events.SubscribeAsync("datastore:commit", onDatastoreCommitEdits)
}

func applyEdit(edit *EditProto, record *RecordProto) {
	var ops []*patch.Op

	for _, o := range edit.Ops {
		op := &patch.Op{}
		for _, p := range o.Path {
			op.Path = append(op.Path, patch.Step{
				Tag:      p.Tag,
				Action:   actionProtoToPatch(p.Action),
				SrcIndex: int(p.SrcIndex),
				DstIndex: int(p.DstIndex),
				MapKey:   nil,
			})
		}
		ops = append(ops, op)
	}

	p := patch.Patch{
		NewValues: edit.NewValues,
		Ops:       ops,
	}

	res, err := patch.ApplyPatch(p, record)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func actionProtoToPatch(action Step_Action) patch.Action {
	switch action {
	case Step_ACTION_ERROR:
		return patch.ActionInvalid
	case Step_ACTION_REPLACE_ALL:
		return patch.ActionReplaceAll
	case Step_ACTION_APPEND_ALL:
		return patch.ActionAppendAll
	case Step_ACTION_REMOVE_ALL:
		return patch.ActionRemoveAll
	case Step_ACTION_REMOVE_ONE:
		return patch.ActionRemoveOne
	case Step_ACTION_REPLACE_ONE:
		return patch.ActionReplaceOne
	case Step_ACTION_STRING_PATCH:
		return patch.ActionStrPatch
	case Step_ACTION_STEP_INTO:
		return patch.ActionStepInto
	default:
		return patch.ActionInvalid
	}
}

func intakeEdit(n *EditProto, pubKey []byte, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	m := jsonpb.Marshaler{}

	var buf bytes.Buffer
	err := m.Marshal(&buf, n)
	if err != nil {
		return nil, err
	}

	fmt.Println(buf.String())

	var el elasticOip5Edit
	el.Edit = buf.Bytes()
	el.Reference = oip.TxidToString(n.Reference)
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
	Edit      json.RawMessage `json:"edit"`
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
	after, err := queryEdits(edits, after)
	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
	}
	if after != nil {
		goto moreEdits
	}

	for _, value := range edits {
		fmt.Println(value)

		// ToDo: apply edit
	}

}

func queryEdits(edits []elasticOip5Edit, after []interface{}) ([]interface{}, error) {
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
		return nil, err
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

	return nextAfter, nil
}
