package oip5

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/azer/logger"
	"github.com/golang/protobuf/jsonpb"
	lru "github.com/hashicorp/golang-lru"
	"github.com/spf13/viper"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/config"
	"github.com/oipwg/oip/datastore"
)

var recordCacheDepth = 10000
var recordCache *lru.Cache

func init() {
	recordCache, _ = lru.New(recordCacheDepth)

	config.OnPostConfig(func(ctx context.Context) {
		rcd := viper.GetInt("oip.oip5.recordCacheDepth")
		if rcd != recordCacheDepth && rcd > 0 {
			recordCacheDepth = rcd
			recordCache.Resize(recordCacheDepth)
		}
	})
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