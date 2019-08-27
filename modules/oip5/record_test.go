package oip5

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/oipProto"
)

// import (
// 	"bytes"
// 	"compress/gzip"
// 	"context"
// 	"encoding/base64"
// 	"fmt"
// 	"strings"
// 	"testing"
//
// 	"github.com/bitspill/flod/flojson"
// 	"github.com/golang/protobuf/jsonpb"
// 	"github.com/golang/protobuf/proto"
// 	"github.com/golang/protobuf/ptypes"
// 	"github.com/golang/protobuf/ptypes/any"
// 	"github.com/oipwg/oip/datastore"
// 	"github.com/oipwg/oip/oipProto"
// )
//
// func TestCreateRecord(t *testing.T) {
// 	hero := &Tmpl_8D66C6AFF9BDD8EE{
// 		Powers: []string{"flight", "invisibility"},
// 	}
// 	basic := &Tmpl_00000000000BA51C{
// 		Title:       "Sintel",
// 		Description: "Sintel, a free, Creative Commons movie",
// 		Year:        2010,
// 	}
//
// 	heroAny, err := ptypes.MarshalAny(hero)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	basicAny, err := ptypes.MarshalAny(basic)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	file := &Tmpl_000000000000F113{
// 		Location:    "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent",
// 		Network:     Network_WEB_TORRENT,
// 		ContentType: "video/mp4",
// 		DisplayName: "Sintel.mp4",
// 		FilePath:    "Sintel/Sintel.mp4",
// 		Size:        129241752,
// 	}
// 	fileAny, err := ptypes.MarshalAny(file)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	o5 := &OipFive{
// 		Record: &RecordProto{
// 			Payment:     nil,
// 			Permissions: nil,
// 			Tags:        nil,
// 			Details: &OipDetails{
// 				Details: []*any.Any{
// 					heroAny,
// 					basicAny,
// 					fileAny,
// 				},
// 			},
// 		},
// 	}
//
// 	_ = o5
//
// 	marsh := jsonpb.Marshaler{Indent: "  "}
// 	fmt.Println(marsh.MarshalToString(o5))
//
// 	b, err := proto.Marshal(o5)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	o5P64 := base64.StdEncoding.EncodeToString(b)
// 	fmt.Println(o5P64)
//
// 	sig := "II5q/gIgTi1w6MJYrbn3fIQsQ6SjCUsRvOBNo2bbzI8qFMSX98bAiRyvE69eY1PgGwLRMNMNeIeOFRAe4nR3qwk="
// 	sigBytes, err := base64.StdEncoding.DecodeString(sig)
// 	sm := &oipProto.SignedMessage{
// 		SerializedMessage: b,
// 		MessageType:       oipProto.MessageTypes_OIP05,
// 		SignatureType:     oipProto.SignatureTypes_Flo,
// 		PubKey:            []byte("ofbB67gqjgaYi45u8Qk2U3hGoCmyZcgbN4"),
// 		Signature:         sigBytes,
// 	}
//
// 	smBytes, err := proto.Marshal(sm)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	p64 := base64.StdEncoding.EncodeToString(smBytes)
// 	fmt.Println(len(p64)+4, "p64:", p64)
//
// 	var wcBuf = new(bytes.Buffer)
// 	var gzw = gzip.NewWriter(wcBuf)
// 	_, err = gzw.Write(smBytes)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	_ = gzw.Close()
// 	gp64 := base64.StdEncoding.EncodeToString(wcBuf.Bytes())
// 	fmt.Println(len(gp64)+5, "gp64:", gp64)
// }
//
// func TestUnmarshalOipDetails(t *testing.T) {
// 	o5 := &OipFive{}
// 	r := strings.NewReader(oipDetailsTestJSON)
// 	err := jsonpb.Unmarshal(r, o5)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	fmt.Println(proto.MarshalTextString(o5))
// 	fmt.Println((&jsonpb.Marshaler{Indent: " "}).MarshalToString(o5))
// }
//
// func TestIntakeRecord(t *testing.T) {
// 	hero := &Tmpl_8D66C6AFF9BDD8EE{
// 		Powers: []string{"flight", "invisibility"},
// 	}
// 	basic := &Tmpl_00000000000BA51C{
// 		Title:       "The first hero",
// 		Description: "They have both flight and invisibility",
// 		Year:        2019,
// 	}
//
// 	heroAny, err := ptypes.MarshalAny(hero)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	basicAny, err := ptypes.MarshalAny(basic)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	o5 := &OipFive{
// 		Record: &RecordProto{
// 			Payment:     nil,
// 			Permissions: nil,
// 			Tags:        nil,
// 			Details: &OipDetails{
// 				Details: []*any.Any{
// 					heroAny,
// 					basicAny,
// 				},
// 			},
// 		},
// 	}
//
// 	tx := &datastore.TransactionData{
// 		BlockHash: "hashDatBlock",
// 		Transaction: &flojson.TxRawResult{
// 			Txid: "12345567890",
// 			Time: 123,
// 		},
// 	}
// 	bir, err := intakeRecord(o5.Record, tx)
//
// 	_ = bir
// 	_ = err
// 	err = datastore.Setup(context.Background())
// 	if err != nil {
// 		panic(err)
// 	}
// 	datastore.AutoBulk.Add(bir)
// 	_, err = datastore.AutoBulk.Do(context.Background())
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }
//
// const oipDetailsTestJSON = `{
//   "record": {
//     "title": "title",
//     "description": "description",
//     "year": "2015",
//     "tags": [
//       "test",
//       "record"
//     ],
//     "storage": {
//       "network": "IPFS",
//       "location": "Qm...",
//       "files": [
//         {
//           "displayName": "file one",
//           "filePath": "file1.txt"
//         }
//       ]
//     },
//     "details": {
//       "tmpl_00000000deadbeef": {
//         "pid": "NS-001",
//         "name": "negative stain",
//         "description": "2 micro liters of sample, wait for 60 seconds, blot with paper 3 times,\n2 micro liters of uranyl acetate, wait for 60 seconds, blot with paper 3 times.",
//         "lab": [
//           "Dexter Labs"
//         ],
//         "institution": [
//           "Cartoon Network"
//         ],
//         "developedBy": [
//           "Charlie",
//           "Doug"
//         ]
//       }
//     }
//   }
// }`

func TestDecodeRecord(t *testing.T) {
	t.SkipNow()

	b, err := base64.StdEncoding.DecodeString("CpcBEpQBOpEBCkMKNHR5cGUuZ29vZ2xlYXBpcy5jb20vb2lwUHJvdG8udGVtcGxhdGVzLnRtcGxfMkYyOUQ4QzASCwoDZmx5EgRlbW1hCkoKNHR5cGUuZ29vZ2xlYXBpcy5jb20vb2lwUHJvdG8udGVtcGxhdGVzLnRtcGxfNUQ4REI4NUISEgoEcnlsbxIFZWFydGgaA3JlZBABGAEiIm9ScG1lWXZqZ2Zoa1NwUFdHTDhlUDVlUHVweW9wM2h6OWoqQR8cQQI9PEBYKuv15qK4aJ1BDg+pdLnuFSRMlNKtUg1zSRv3QTPefPerz8MVTqd5o77mIh4klLFuMzeEt5j/uUiz")
	if err != nil {
		t.Fatal(err)
	}

	sm := &oipProto.SignedMessage{}

	err = proto.Unmarshal(b, sm)
	if err != nil {
		t.Fatal(err)
	}

	o5 := &OipFive{}

	err = proto.Unmarshal(sm.SerializedMessage, o5)
	if err != nil {
		t.Fatal(err)
	}

	_ = datastore.Setup(context.Background())
	_ = LoadTemplatesFromES(context.Background())

	fmt.Println((&jsonpb.Marshaler{}).MarshalToString(o5))
}
