package oip5

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/oipProto"
	"gopkg.in/olivere/elastic.v6"
)

func normalizeRecord(r *RecordProto, tx *datastore.TransactionData) error {
	attr := logger.Attrs{"txid": tx.Transaction.Txid}

	var norms []*NormalizeRecordProto
	if r.Details == nil {
		return errors.New("no details")
	}
	for _, dAny := range r.Details.Details {
		name, err := ptypes.AnyMessageName(dAny)
		if err != nil {
			continue
		}
		tmplName := strings.TrimPrefix(name, "oipProto.templates.tmpl_")

		if len(tmplName) != len(name) {
			id, err := strconv.ParseUint(tmplName, 16, 32)
			if err != nil {
				continue
			}
			n, ok := normalizers[uint32(id)]
			if ok {
				norms = append(norms, n...)
			}
		}
	}

	for _, v := range norms {
		err := applyNorm(r, v, tx)
		if err != nil {
			attr["norm"] = "ToDo" // ToDo: store normalizer identifier
			attr["err"] = err
			log.Error("error applying normalizer", attr)
		}
	}

	return nil
}

func applyNorm(rec *RecordProto, n *NormalizeRecordProto, tx *datastore.TransactionData) error {
	// ToDo: currently all normalized fields are required, if a field is missing proceed with other fields
	// ToDo: validate nf.Name is valid ES field name

	var normalRec = make(map[string]interface{})
	msgIter, err := dynamic.AsDynamicMessage(rec)
	if err != nil {
		log.Error("ToDo")
		return err
	}

	for _, nf := range n.Fields {
		val, err := getNormalizedField(msgIter, nf)
		if err != nil {
			log.Error("unable to normalize field", logger.Attrs{"txid": tx.Transaction.Txid, "field": nf.Name, "err": err, "normId": "ToDo"})
			return err
		}
		normalRec[nf.Name] = val
	}

	j, err := json.Marshal(normalRec)
	if err != nil {
		return err
	}

	var el elasticOip5Record
	el.Record = j
	el.Meta = RMeta{
		Block:       tx.Block,
		BlockHash:   tx.BlockHash,
		Deactivated: false,
		Time:        tx.Transaction.Time,
		Tx:          tx,
		Txid:        tx.Transaction.Txid,
		Type:        "normalized",
		Normalizer:  int64(n.MainTemplate),
	}

	bir := elastic.NewBulkIndexRequest().
		Index(datastore.Index("oip5_norm")).
		Type("_doc").
		Doc(el)

	datastore.AutoBulk.Add(bir)

	return nil
}

func getNormalizedField(firstIterator *dynamic.Message, nf *NormalField) (interface{}, error) {
	var normalizedValue []interface{}

	var nextMessageIterator []*dynamic.Message
	currentMessageIterator := []*dynamic.Message{firstIterator}
	for _, nfp := range nf.Path {
		for _, mi := range currentMessageIterator {
			stepValue, err := stepPath(nfp, mi)
			if err != nil {
				return nil, err
			}
			if nextIter, ok := stepValue.(*dynamic.Message); ok {
				nextMessageIterator = append(nextMessageIterator, nextIter)
			} else if interfaceSlice, ok := stepValue.([]interface{}); ok {
				for _, is := range interfaceSlice {
					if nextIter, ok := is.(*dynamic.Message); ok {
						nextMessageIterator = append(nextMessageIterator, nextIter)
					}
				}
			} else {
				normalizedValue = append(normalizedValue, stepValue)
			}
		}
		currentMessageIterator = nextMessageIterator
		nextMessageIterator = []*dynamic.Message{}
	}

	if len(normalizedValue) == 1 {
		return normalizedValue[0], nil
	} else {
		return normalizedValue, nil
	}
}

func stepPath(field *Field, msgIterator *dynamic.Message) (interface{}, error) {
	if field.Tag == 3 {
		fmt.Println("hiya")
	}
	if field.Template != 0 {
		nextIterator, err := enterTemplate(msgIterator, field)
		if err != nil {
			return nil, err
		}
		return nextIterator, nil
	}
	f := msgIterator.FindFieldDescriptor(field.Tag)
	if f == nil {
		log.Error("field with tag not found") // ToDo more info
		return nil, errors.New("field with tag not found")
	}
	if Field_Type(f.GetType()) != field.Type {
		log.Error("field of unexpected type") // ToDo more info
		return nil, errors.New("field of unexpected type")
	}
	fieldValue := msgIterator.GetFieldByNumber(int(field.Tag))
	if sliceFieldValue, ok := fieldValue.([]interface{}); ok {
		var returnSlice []interface{}

		if Field_Type(f.GetType()) == Field_TYPE_MESSAGE {
			for _, fv := range sliceFieldValue {
				nextMessage, err := resolveMessage(fv, f)
				if err != nil {
					return nil, err
				}
				returnSlice = append(returnSlice, nextMessage)
			}
		} else {
			returnSlice = append(returnSlice, sliceFieldValue...)
		}
		return returnSlice, nil
	} else if Field_Type(f.GetType()) == Field_TYPE_MESSAGE {
		nextMessage, err := resolveMessage(fieldValue, f)
		if err != nil {
			return nil, err
		}
		return nextMessage, nil
	}
	return fieldValue, nil
}

func resolveMessage(fieldValue interface{}, f *desc.FieldDescriptor) (*dynamic.Message, error) {
	if txid, ok := fieldValue.(*oipProto.Txid); ok {
		ref, err := resolveTxidReference(txid)
		if err != nil {
			return nil, err
		}
		return ref, nil
	} else if pm, ok := fieldValue.(proto.Message); ok {
		r, err := dynamic.AsDynamicMessage(pm)
		if err != nil {
			log.Error("ToDo")
			return nil, err
		}

		if f.AsFieldDescriptorProto().GetTypeName() == ".oipProto.Txid" {
			if link, ok := fieldValue.(*oipProto.Txid); ok {
				ref, err := resolveTxidReference(link)
				if err != nil {
					return nil, err
				}
				return ref, nil
			}
		}
		return r, nil
	}
	return nil, errors.New("fieldValue not a message")
}

func resolveTxidReference(txid *oipProto.Txid) (*dynamic.Message, error) {
	r, err := GetRecord(oipProto.TxidToString(txid))
	if err != nil {
		log.Error("unable to get linked record")
		return nil, err
	}
	dr, err := dynamic.AsDynamicMessage(r.Record)
	if err != nil {
		log.Error("unable to dynamic record")
		return nil, err
	}
	return dr, nil
}

func enterTemplate(recMsg *dynamic.Message, field *Field) (*dynamic.Message, error) {
	name := recMsg.GetMessageDescriptor().GetFullyQualifiedName()
	if name != "oipProto.RecordProto" {
		log.Error("cannot enter details of on record type (%s)", name)
		return nil, fmt.Errorf("cannot enter details of on record type (%s)", name)
	}
	d, err := recMsg.TryGetFieldByNumber(7) // record.details tag number
	if err != nil {
		log.Error("unable to get details")
		return nil, err
	}
	if details, ok := d.(*OipDetails); ok {
		for _, detAny := range details.Details {
			dMsg, ok := detailsMessageFromTemplate(detAny, field.Template)
			if ok {
				return dMsg, nil
			}
		}
	}
	return nil, errors.New("details template not found")
}

func detailsMessageFromTemplate(detAny *any.Any, id uint32) (*dynamic.Message, bool) {
	name, err := ptypes.AnyMessageName(detAny)
	if err != nil {
		return nil, false
	}
	tmplName := strings.TrimPrefix(name, "oipProto.templates.tmpl_")
	tmplId, err := strconv.ParseUint(tmplName, 16, 32)
	if err != nil {
		return nil, false
	}
	if id != uint32(tmplId) {
		return nil, false
	}

	msg, err := CreateNewMessage(name)
	if err != nil {
		log.Error("error creating new message", logger.Attrs{"err": err, "detAny": detAny, "id": id})
		return nil, false
	}

	err = ptypes.UnmarshalAny(detAny, msg)
	if err != nil {
		return nil, false
	}

	dMsg, err := dynamic.AsDynamicMessage(msg)
	if err != nil {
		return nil, false
	}

	return dMsg, true
}

func intakeNormalize(n *NormalizeRecordProto, pubKey []byte, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	m := jsonpb.Marshaler{}

	var buf bytes.Buffer
	err := m.Marshal(&buf, n)
	if err != nil {
		return nil, err
	}

	fmt.Println(buf.String())

	var el elasticOip5Normalize
	el.Normalize = buf.Bytes()
	el.Meta = NMeta{
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		SignedBy:  string(pubKey),
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
	}

	bir := elastic.NewBulkIndexRequest().
		Index(datastore.Index("oip5_normalize")).
		Type("_doc").
		Id(tx.Transaction.Txid).
		Doc(el)

	norms := normalizers[n.MainTemplate]
	normalizers[n.MainTemplate] = append(norms, n)

	return bir, nil
}

type NMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	SignedBy  string                     `json:"signed_by"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"-"`
	Txid      string                     `json:"txid"`
}

type elasticOip5Normalize struct {
	Normalize json.RawMessage `json:"normalize"`
	Meta      NMeta           `json:"meta"`
}
