package oip5

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/azer/logger"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/oipwg/proto/go/pb_oip5"
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

func intakeRecord(r *pb_oip5.RecordProto, pubKey []byte, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	m := jsonpb.Marshaler{}

	var buf bytes.Buffer
	err := m.Marshal(&buf, r)
	if err != nil {
		return nil, err
	}

	raw, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	raw64 := base64.StdEncoding.EncodeToString(raw)

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
		Original:      tx.Transaction.Txid,
		History:       []string{tx.Transaction.Txid},
		Type:          "oip5",
		RecordRaw:     raw64,
		Latest:        true,
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
			Record: &pb_oip5.RecordProto{},
		}

		raw, err := base64.StdEncoding.DecodeString(eRec.Meta.RecordRaw)
		err = proto.Unmarshal(raw, rec.Record)
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
	Record *pb_oip5.RecordProto `json:"record"`
	Meta   RMeta                `json:"meta"`
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
	Latest        bool                       `json:"latest"`
	Original      string                     `json:"original"`
	History       []string                   `json:"history"`
	RecordRaw     string                     `json:"record_raw"`
}
