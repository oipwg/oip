package oip5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/oipProto"
	"gopkg.in/olivere/elastic.v6"
)

func normalizeRecord(r *RecordProto, tx *datastore.TransactionData) error {
	attr := logger.Attrs{"txid": tx.Transaction.Txid}

	var norms []*NormalizeRecordProto
	for _, dAny := range r.Details.Details {
		name, err := ptypes.AnyMessageName(dAny)
		if err != nil {
			continue
		}
		tmplName := strings.TrimPrefix(name, "oipProto.templates.tmpl_")

		if len(tmplName) != len(name) {
			id, err := strconv.ParseUint(tmplName, 16, 64)
			if err != nil {
				continue
			}
			n, ok := normalizers[id]
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

func applyNorm(r *RecordProto, n *NormalizeRecordProto, tx *datastore.TransactionData) error {
	// ToDo: currently all normalized fields are required, if a field is missing proceed with other fields
	// ToDo: validate nf.Name is valid ES field name

	var nRec = make(map[string]interface{})

	for _, nf := range n.Fields {
		rec, err := dynamic.AsDynamicMessage(r)
		if err != nil {
			log.Error("ToDo")
			return err
		}
		var val interface{} = nil
		for _, nfp := range nf.Path {
			if nfp.Tag == 3 {
				fmt.Println("hiya")
			}
			if nfp.Template != 0 {
				name := rec.GetMessageDescriptor().GetFullyQualifiedName()
				if name != "oipProto.RecordProto" {
					log.Error("expected record got %s", name)
					// return errors.New("did not receive expected record type")
				}
				d, err := rec.TryGetFieldByNumber(7) // record.details tag number
				if err != nil {
					log.Error("unable to get details")
					// return err
				}
				if details, ok := d.(*OipDetails); ok {
					for _, detAny := range details.Details {
						name, err := ptypes.AnyMessageName(detAny)
						if err != nil {
							// return err
						}

						tmplName := strings.TrimPrefix(name, "oipProto.templates.tmpl_")
						id, err := strconv.ParseUint(tmplName, 16, 64)
						if err != nil {
							// continue
						}

						if id != nfp.Template {
							// continue
						}

						msg, err := CreateNewMessage(name)
						if err != nil {
							// return err
						}
						err = ptypes.UnmarshalAny(detAny, msg)
						if err != nil {
							// return err
						}

						r, err := dynamic.AsDynamicMessage(msg)
						if err != nil {
							// return err
						}
						rec = r
					}
				}
				continue
			}

			f := rec.FindFieldDescriptor(nfp.Tag)
			if f == nil {
				log.Error("field with tag not found") // ToDo more info
				// return errors.New("field with tag not found")
			}
			if Field_Type(f.GetType()) != nfp.Type {
				log.Error("field of unexpected type") // ToDo more info
				// return errors.New("field of unexpected type")
			}
			val = rec.GetFieldByNumber(int(nfp.Tag))
			if Field_Type(f.GetType()) == Field_TYPE_MESSAGE {
				spew.Dump(val)

				// ToDo: append/[:-1] make a basic queue of rec/val to process arrays
				//  break mega function to multiple methods

				if sliceInt, ok := val.([]interface{}); ok {
					for _, si := range sliceInt {
						_ = si
						// ToDo: extract below txid/proto.message parsing to function and call
					}
				} else if txid, ok := val.(*oipProto.Txid); ok {
					r, err := GetRecord(oipProto.TxidToString(txid))
					if err != nil {
						log.Error("unable to get linked record")
						// return err
					}
					dr, err := dynamic.AsDynamicMessage(r.Record)
					if err != nil {
						log.Error("unable to dynamic record")
						// return err
					}
					rec = dr
				} else if pm, ok := val.(proto.Message); ok {
					r, err := dynamic.AsDynamicMessage(pm)
					if err != nil {
						log.Error("ToDo")
						// return err
					}
					rec = r

					if f.AsFieldDescriptorProto().GetTypeName() == ".oipProto.Txid" {
						if link, ok := val.(*oipProto.Txid); ok {
							r, err := GetRecord(oipProto.TxidToString(link))
							if err != nil {
								log.Error("unable to get linked record")
								// return err
							}
							dr, err := dynamic.AsDynamicMessage(r.Record)
							if err != nil {
								log.Error("unable to dynamic record")
								// return err
							}
							rec = dr
						} else {
							fmt.Println("not ok")
						}
					} else {
						fmt.Println("not txid")
					}
				} else {
					fmt.Println("not proto.message")
				}
			} else {
				fmt.Println("not message")
			}
		}
		nRec[nf.Name] = val
	}

	j, err := json.Marshal(nRec)
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

func intakeNormalize(n *NormalizeRecordProto, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
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
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
	}

	bir := elastic.NewBulkIndexRequest().
		Index(datastore.Index("oip5_normalize")).
		Type("_doc").
		Id(tx.Transaction.Txid).
		Doc(el)

	norms, _ := normalizers[n.MainTemplate]
	normalizers[n.MainTemplate] = append(norms, n)

	return bir, nil
}

type NMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"-"`
	Txid      string                     `json:"txid"`
}

type elasticOip5Normalize struct {
	Normalize json.RawMessage `json:"normalize"`
	Meta      NMeta           `json:"meta"`
}
