package oip5

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync"

	"github.com/azer/logger"
	patch "github.com/bitspill/protoPatch"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/oipwg/proto/go/pb_oip"
	"github.com/oipwg/proto/go/pb_oip5"
	"github.com/oipwg/proto/go/pb_oip5/pb_templates"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/modules/oip5/templates"
)

var editIndex = "oip5_edit"
var editCommitMutex sync.Mutex

const registeredPublisherTypeUrl = "type.googleapis.com/oipProto.templates.tmpl_433C2783"

func init() {
	events.SubscribeAsync("datastore:commit", onDatastoreCommitEdits)
	datastore.RegisterMapping("oip5_edit", "oip5_edit.json")
}

func intakeEdit(n *pb_oip5.EditProto, pubKey []byte, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	m := n.GetMod()
	if m == nil {
		return nil, errors.New("no mod")
	}

	var err error
	var jm jsonpb.Marshaler
	var jsonBuf bytes.Buffer
	var pm proto.Message

	switch mod := m.(type) {
	case *pb_oip5.EditProto_Patch:
		pm = mod.Patch
		err = jm.Marshal(&jsonBuf, n.GetPatch())
	case *pb_oip5.EditProto_Template:
		pm = mod.Template
		err = jm.Marshal(&jsonBuf, n.GetTemplate())
	}
	if err != nil {
		return nil, err
	}

	rawBuf, err := proto.Marshal(pm)
	if err != nil {
		return nil, err
	}

	rawB64 := base64.StdEncoding.EncodeToString(rawBuf)

	var el elasticOip5Edit
	switch m.(type) {
	case *pb_oip5.EditProto_Patch:
		el.PatchJson = jsonBuf.Bytes()
		el.PatchRaw = rawB64
	case *pb_oip5.EditProto_Template:
		el.TemplateJson = jsonBuf.Bytes()
		el.TemplateRaw = rawB64
	}
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
	PatchJson    json.RawMessage `json:"patch_json"`
	PatchRaw     string          `json:"patch_raw"`
	TemplateJson json.RawMessage `json:"template_json"`
	TemplateRaw  string          `json:"template_raw"`
	Reference    string          `json:"reference"`
	Meta         EMeta           `json:"meta"`
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

		if len(edit.TemplateRaw) != 0 {
			editTemplate(edit)
		} else {
			editRecord(edit)
		}
	}
}

func editTemplate(edit elasticOip5Edit) {
	tmpl, err := templates.GetTemplate(edit.Reference)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to obtain template for edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	if tmpl.SignedBy != edit.Meta.SignedBy {
		log.Error("edit not signed by template owner", logger.Attrs{"reference": edit.Reference, "txid": edit.Meta.Txid})
		markEditInvalid(edit.Meta.Txid)
		return
	}

	// ToDo validate changes are backwards compatible
	// ToDo rename field?

	err = templates.EditTemplate(tmpl, edit.TemplateRaw, edit.Meta.Txid)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to edit template", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
	}
}

func editRecord(edit elasticOip5Edit) {
	rec, err := GetRecord(edit.Reference)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to obtain record for edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	if rec.Meta.SignedBy != edit.Meta.SignedBy {
		log.Error("edit not signed by record owner", logger.Attrs{"reference": edit.Reference, "txid": edit.Meta.Txid})
		markEditInvalid(edit.Meta.Txid)
		return
	}

	b, err := base64.StdEncoding.DecodeString(edit.PatchRaw)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to decode raw patch", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	pp := &patch.ProtoPatch{}
	err = proto.Unmarshal(b, pp)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to decode proto patch for edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	p, err := patch.FromProto(pp)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to decode patch for edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	a, err := patch.ApplyPatch(*p, rec.Record)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to apply edit to record", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	newRec, ok := a.(*pb_oip5.RecordProto)
	if !ok {
		markEditInvalid(edit.Meta.Txid)
		log.Error("patch result is no longer a record", logger.Attrs{"reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	// Check to see if a publisher name was edited
	regPubNameChanged := false
	for i := range newRec.Details.Details {
		if len(rec.Record.Details.Details[i].TypeUrl) == 52 &&
			rec.Record.Details.Details[i].TypeUrl[44:] == registeredPublisherTypeUrl[44:] {
			regPub := &pb_templates.Tmpl_433C2783{}
			err := ptypes.UnmarshalAny(newRec.Details.Details[i], regPub)
			if err != nil {
				markEditInvalid(edit.Meta.Txid)
				log.Error("unable to decode reg pub any", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
				return
			}
			if rec.Meta.PublisherName != regPub.Name {
				rec.Meta.PublisherName = regPub.Name
				regPubNameChanged = true
			}
		}
	}

	rec.Record = newRec

	rec.Meta.History = append(rec.Meta.History, edit.Meta.Txid)
	rec.Meta.LastModified = edit.Meta.Time

	m := jsonpb.Marshaler{}

	var buf bytes.Buffer
	err = m.Marshal(&buf, newRec)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to marshal record json post edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	rawRec, err := proto.Marshal(newRec)
	if err != nil {
		markEditInvalid(edit.Meta.Txid)
		log.Error("unable to marshal record proto post edit", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		return
	}

	rawRec64 := base64.StdEncoding.EncodeToString(rawRec)
	rec.Meta.RecordRaw = rawRec64

	var el elasticOip5Record
	el.Record = buf.Bytes()
	el.Meta = rec.Meta

	bir := elastic.NewBulkIndexRequest().
		Index(datastore.Index(o5RecordIndexName)).
		Type("_doc").
		Id(edit.Meta.Txid).
		Doc(el)

	recordCache.Add(rec.Meta.Original, rec)

	datastore.AutoBulk.Add(bir)

	if len(rec.Meta.History) > 1 {
		for _, prevIteration := range rec.Meta.History[:len(rec.Meta.History)-1] {
			bur := elastic.NewBulkUpdateRequest().
				Index(datastore.Index(o5RecordIndexName)).
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

	if regPubNameChanged {
		datastore.AutoBulk.Commit()
		err := updatePublisherName(rec.Meta.SignedBy, rec.Meta.PublisherName)
		if err != nil {
			log.Error("unable to update publisher name", logger.Attrs{"err": err, "reference": edit.Reference, "txid": edit.Meta.Txid})
		}
	}
}

func updatePublisherName(pubAddr, pubName string) error {
	log.Info("updating publisher name", logger.Attrs{"pubAddr": pubAddr, "pubName": pubName})

	publisherCache.Add(pubAddr, pubName)

	s := elastic.NewScript("ctx._source.meta.publisher_name=params.pubName;").
		Param("pubName", pubName).
		Type("inline").
		Lang("painless")

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.signed_by", pubAddr),
		elastic.NewTermQuery("meta.latest", true),
	)
	cuq := datastore.Client().UpdateByQuery(datastore.Index("oip5_record")).Query(q).
		Type("_doc").Script(s).Refresh("true")

	res, err := cuq.Do(context.TODO())
	if err != nil {
		log.Error("unable to update publisher name", logger.Attrs{"err": err, "pubAddr": pubAddr, "pubName": pubName})
		return err
	}
	log.Info("update publisher name completed", logger.Attrs{"total": res.Total, "took": res.Took, "updated": res.Updated, "pubAddr": pubAddr, "pubName": pubName})
	return nil
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
