package oip5

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/bitspill/flod/flojson"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/oipwg/oip/datastore"
)

func TestCreateRecord(t *testing.T) {
	researchProtocol := &Tmpl_00000000Deadbeef{
		Pid:         "NS-001",
		Name:        "negative stain",
		Lab:         []string{"Dexter Labs"},
		Institution: []string{"Cartoon Network"},
		DevelopedBy: []string{"Charlie", "Doug"},
		Description: "2 micro liters of sample, wait for 60 seconds, blot with paper 3 times,\n2 micro liters of uranyl acetate, wait for 60 seconds, blot with paper 3 times.",
	}

	rpAny, err := ptypes.MarshalAny(researchProtocol)
	if err != nil {
		t.Fatal(err)
	}

	o5 := &OipFive{
		Record: &RecordProto{
			Title:       "title",
			Description: "description",
			Tags:        []string{"test", "record"},
			Year:        2015,
			Payment:     nil,
			Storage: &Storage{
				Network:  Network_IPFS,
				Location: "Qm...",
				Files: []*File{
					{
						DisplayName: "file one",
						FilePath:    "file1.txt",
					},
				},
			},
			Details: &OipDetails{
				Details: []*any.Any{
					rpAny,
				},
			},
		},
	}

	_ = o5

	marsh := jsonpb.Marshaler{Indent: "  "}
	fmt.Println(marsh.MarshalToString(o5))
}

func TestUnmarshalOipDetails(t *testing.T) {
	o5 := &OipFive{}
	r := strings.NewReader(oipDetailsTestJSON)
	err := jsonpb.Unmarshal(r, o5)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(proto.MarshalTextString(o5))
	fmt.Println((&jsonpb.Marshaler{Indent: " "}).MarshalToString(o5))
}

func TestIntakeRecord(t *testing.T) {
	researchProtocol := &Tmpl_00000000Deadbeef{
		Pid:         "NS-001",
		Name:        "negative stain",
		Lab:         []string{"Dexter Labs"},
		Institution: []string{"Cartoon Network"},
		DevelopedBy: []string{"Charlie", "Doug"},
		Description: "2 micro liters of sample, wait for 60 seconds, blot with paper 3 times,\n2 micro liters of uranyl acetate, wait for 60 seconds, blot with paper 3 times.",
	}

	rpAny, err := ptypes.MarshalAny(researchProtocol)
	if err != nil {
		t.Fatal(err)
	}

	o5 := &OipFive{
		Record: &RecordProto{
			Title:       "title",
			Description: "description",
			Tags:        []string{"test", "record"},
			Year:        2015,
			Payment:     nil,
			Storage: &Storage{
				Network:  Network_IPFS,
				Location: "Qm...",
				Files: []*File{
					{
						DisplayName: "file one",
						FilePath:    "file1.txt",
					},
				},
			},
			Details: &OipDetails{
				Details: []*any.Any{
					rpAny,
				},
			},
		},
	}

	// tx.Block,
	// 		BlockHash:   tx.BlockHash,
	// 		Deactivated: false,
	// 		Time:        tx.Transaction.Time,
	// 		Tx:          tx,
	// 		Txid:        tx.Transaction.Txid,
	tx := &datastore.TransactionData{
		BlockHash: "hashDatBlock",
		Transaction: &flojson.TxRawResult{
			Txid: "12345567890",
			Time: 123,
		},
	}
	bir, err := intakeRecord(o5.Record, tx)

	_ = bir
	_ = err
	err = datastore.Setup(context.Background())
	if err != nil {
		panic(err)
	}
	datastore.AutoBulk.Add(bir)
	_, err = datastore.AutoBulk.Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
}

func (m *OipDetails) MarshalJSONPB(marsh *jsonpb.Marshaler) ([]byte, error) {
	var detMap = make(map[string]*json.RawMessage)

	// "@type": "type.googleapis.com/oip5.record.templates.tmpl_00000000deadbeef",
	// oip5.record.templates.tmpl_00000000deadbeef
	for _, detAny := range m.Details {
		name, err := ptypes.AnyMessageName(detAny)
		if err != nil {
			return nil, err
		}

		tmplName := strings.TrimPrefix(name, "oip5.record.templates.")
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

		tmplName = strings.Replace(tmplName, "deadbeef", "cafebabe", -1)
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
		if len(k) == 21 && strings.HasPrefix(k, "tmpl_") {
			k = "type.googleapis.com/oip5.record.templates." + k
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

const oipDetailsTestJSON = `{
  "record": {
    "title": "title",
    "description": "description",
    "year": "2015",
    "tags": [
      "test",
      "record"
    ],
    "storage": {
      "network": "IPFS",
      "location": "Qm...",
      "files": [
        {
          "displayName": "file one",
          "filePath": "file1.txt"
        }
      ]
    },
    "details": {
      "tmpl_00000000deadbeef": {
        "pid": "NS-001",
        "name": "negative stain",
        "description": "2 micro liters of sample, wait for 60 seconds, blot with paper 3 times,\n2 micro liters of uranyl acetate, wait for 60 seconds, blot with paper 3 times.",
        "lab": [
          "Dexter Labs"
        ],
        "institution": [
          "Cartoon Network"
        ],
        "developedBy": [
          "Charlie",
          "Doug"
        ]
      }
    }
  }
}`
