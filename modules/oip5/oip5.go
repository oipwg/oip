package oip5

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/azer/logger"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	lru "github.com/hashicorp/golang-lru"
	jsoniter "github.com/json-iterator/go"
	"github.com/oipwg/oip/oipProto"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

const recordCacheDepth = 10000
const publisherCacheDepth = 1000

var recordCache *lru.Cache
var publisherCache *lru.Cache

var normalizers = make(map[uint32][]*NormalizeRecordProto)

func init() {
	log.Info("init oip5")
	events.SubscribeAsync("modules:oip5:msg", on5msg)

	datastore.RegisterMapping("oip5_templates", "oip5_templates.json")
	datastore.RegisterMapping("oip5_record", "oip5_record.json")

	recordCache, _ = lru.New(recordCacheDepth)
	publisherCache, _ = lru.New(publisherCacheDepth)
}

func on5msg(msg oipProto.SignedMessage, tx *datastore.TransactionData) {
	attr := logger.Attrs{"txid": tx.Transaction.Txid}
	log.Info("oip5 ", attr)

	var o5 = &OipFive{}

	err := proto.Unmarshal(msg.SerializedMessage, o5)
	if err != nil {
		attr["err"] = err
		log.Error("unable to unmarshal serialized message", attr)
		return
	}

	nonNilAction := false
	if o5.RecordTemplate != nil {
		nonNilAction = true
		bir, err := intakeRecordTemplate(o5.RecordTemplate, msg.PubKey, tx)
		if err != nil {
			attr["err"] = err
			log.Error("unable to process RecordTemplate", attr)
		} else {
			attr["templateName"] = o5.RecordTemplate.FriendlyName
			log.Info("adding RecordTemplate", attr)
			datastore.AutoBulk.Add(bir)
		}
	}

	if o5.Record != nil {
		nonNilAction = true
		bir, err := intakeRecord(o5.Record, msg.PubKey, tx)
		if err != nil {
			attr["err"] = err
			log.Error("unable to process Record", attr)
		} else {
			attr["deets"] = o5.Record.Details
			log.Info("adding o5 record", attr)
			datastore.AutoBulk.Add(bir)

			err := normalizeRecord(o5.Record, tx)
			if err != nil {
				attr["err"] = err
				log.Error("ERROR", attr)
			}
		}
	}

	if o5.Normalize != nil {
		nonNilAction = true
		bir, err := intakeNormalize(o5.Normalize, msg.PubKey, tx)
		if err != nil {
			attr["err"] = err
			log.Error("unable to process Normalize", attr)
		} else {
			log.Info("adding o5 normalize", attr)
			datastore.AutoBulk.Add(bir)
		}
	}

	if !nonNilAction {
		log.Error("no supported oip5 action to process", attr)
	}
}

func intakeRecord(r *RecordProto, pubKey []byte, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	m := jsonpb.Marshaler{}

	var buf bytes.Buffer
	err := m.Marshal(&buf, r)
	if err != nil {
		return nil, err
	}

	strPubKey := string(pubKey)

	pubName, err := GetPublisherName(strPubKey)
	if err != nil {
		log.Error("error getting publisher name", logger.Attrs{"txid": tx.Transaction.Txid, "pubkey": strPubKey, "err": err})
	}

	var el elasticOip5Record
	el.Record = buf.Bytes()
	el.Meta = RMeta{
		Block:         tx.Block,
		BlockHash:     tx.BlockHash,
		Deactivated:   false,
		SignedBy:      strPubKey,
		PublisherName: pubName,
		Time:          tx.Transaction.Time,
		Tx:            tx,
		Txid:          tx.Transaction.Txid,
		Type:          "oip5",
	}

	bir := elastic.NewBulkIndexRequest().
		Index(datastore.Index("oip5_record")).
		Type("_doc").
		Id(tx.Transaction.Txid).
		Doc(el)

	cr := &oip5Record{
		Record: r,
		Meta:   el.Meta,
	}

	recordCache.Add(el.Meta.Txid, cr)

	return bir, nil
}

func GetRecord(txid string) (*oip5Record, error) {
	r, found := recordCache.Get(txid)
	if found {
		return r.(*oip5Record), nil
	}

	get, err := datastore.Client().Get().Index(datastore.Index("oip5_record")).Type("_doc").Id(txid).Do(context.Background())
	if err != nil {
		return nil, err
	}
	if get.Found {
		var eRec elasticOip5Record
		err := json.Unmarshal(*get.Source, &eRec)
		if err != nil {
			return nil, err
		}

		rec := &oip5Record{
			Meta:   eRec.Meta,
			Record: &RecordProto{},
		}
		// templates oipProto.templates.tmpl_... not being added to protobuf types

		umarsh := jsonpb.Unmarshaler{
			AnyResolver: &o5AnyResolver{},
		}

		err = umarsh.Unmarshal(bytes.NewReader(eRec.Record), rec.Record)
		if err != nil {
			return nil, err
		}

		recordCache.Add(rec.Meta.Txid, rec)

		return rec, nil
	}
	return nil, errors.New("ID not found")
}

func GetPublisherName(pubKey string) (string, error) {
	pni, found := publisherCache.Get(pubKey)
	if found {
		return pni.(string), nil
	}

	q := elastic.NewBoolQuery().Must(
		elastic.NewExistsQuery("record.details.tmpl_433C2783.name"),
		elastic.NewTermQuery("meta.signed_by", pubKey),
	)
	results, err := datastore.Client().
		Search(datastore.Index("oip5_record")).
		Type("_doc").
		Query(q).
		Size(1).
		Sort("meta.time", false).
		Do(context.TODO())
	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		return "", err
	}

	if len(results.Hits.Hits) > 0 {
		src := *results.Hits.Hits[0].Source
		pn := jsoniter.Get(src, "record", "details", "tmpl_433C2783", "name").ToString()
		publisherCache.Add(pubKey, pn)
		return pn, nil
	}

	return "", errors.New("unable to locate publisher")
}

type elasticOip5Record struct {
	Record json.RawMessage `json:"record"`
	Meta   RMeta           `json:"meta"`
}

type oip5Record struct {
	Record *RecordProto `json:"record"`
	Meta   RMeta        `json:"meta"`
}

type RMeta struct {
	Block         int64                      `json:"block"`
	BlockHash     string                     `json:"block_hash"`
	Deactivated   bool                       `json:"deactivated"`
	SignedBy      string                     `json:"signed_by"`
	PublisherName string                     `json:"publisher_name"`
	Time          int64                      `json:"time"`
	Tx            *datastore.TransactionData `json:"-"`
	Txid          string                     `json:"txid"`
	Type          string                     `json:"type"`
	Normalizer    int64                      `json:"normalizer_id,omitempty"`
}

func (m *OipDetails) MarshalJSONPB(marsh *jsonpb.Marshaler) ([]byte, error) {
	var detMap = make(map[string]*json.RawMessage)

	// "@type": "type.googleapis.com/oipProto.templates.tmpl_deadbeef",
	// oipProto.templates.tmpl_deadbeef
	for _, detAny := range m.Details {
		name, err := ptypes.AnyMessageName(detAny)
		if err != nil {
			return nil, err
		}

		msg, err := CreateNewMessage(name)
		if err != nil {
			return nil, err
		}
		err = ptypes.UnmarshalAny(detAny, msg)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := marsh.Marshal(&buf, msg); err != nil {
			return nil, err
		}
		jr := json.RawMessage(buf.Bytes())

		tmplName := strings.TrimPrefix(name, "oipProto.templates.")
		detMap[tmplName] = &jr
	}

	if marsh.Indent != "" {
		return json.MarshalIndent(detMap, strings.Repeat(marsh.Indent, 2), marsh.Indent)
	}

	return json.Marshal(detMap)
}

func (m *OipDetails) UnmarshalJSONPB(u *jsonpb.Unmarshaler, b []byte) error {
	var detMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &detMap); err != nil {
		return err
	}

	for k, v := range detMap {
		if len(k) == 13 && strings.HasPrefix(k, "tmpl_") {
			k = "type.googleapis.com/oipProto.templates." + k
		}

		var jsonFields map[string]*json.RawMessage
		if err := json.Unmarshal([]byte(*v), &jsonFields); err != nil {
			return err
		}

		b, err := json.Marshal(k)
		if err != nil {
			return err
		}
		jr := json.RawMessage(b)
		jsonFields["@type"] = &jr

		b, err = json.Marshal(jsonFields)
		if err != nil {
			return err
		}
		a := &any.Any{}
		br := bytes.NewReader(b)
		err = u.Unmarshal(br, a)
		if err != nil {
			return err
		}
		m.Details = append(m.Details, a)
	}

	return nil
}

type o5AnyResolver struct{}

func (r *o5AnyResolver) Resolve(typeUrl string) (proto.Message, error) {
	mname := typeUrl
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}

	// try default behavior first
	mt := proto.MessageType(mname)
	if mt != nil {
		return reflect.New(mt.Elem()).Interface().(proto.Message), nil
	}

	return CreateNewMessage(mname)
}
