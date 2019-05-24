package oip5

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/bitspill/flod/flojson"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/oipProto"

	_ "github.com/oipwg/oip/modules/oip"
)

func TestIntakeRecordTemplate(t *testing.T) {
	t.SkipNow()

	b, err := base64.StdEncoding.DecodeString("CvMBCh5nb29nbGUvcHJvdG9idWYvZHVyYXRpb24ucHJvdG8SD2dvb2dsZS5wcm90b2J1ZiI6CghEdXJhdGlvbhIYCgdzZWNvbmRzGAEgASgDUgdzZWNvbmRzEhQKBW5hbm9zGAIgASgFUgVuYW5vc0J8ChNjb20uZ29vZ2xlLnByb3RvYnVmQg1EdXJhdGlvblByb3RvUAFaKmdpdGh1Yi5jb20vZ29sYW5nL3Byb3RvYnVmL3B0eXBlcy9kdXJhdGlvbvgBAaICA0dQQqoCHkdvb2dsZS5Qcm90b2J1Zi5XZWxsS25vd25UeXBlc2IGcHJvdG8zCvIKCgdwLnByb3RvEhVvaXA1LnJlY29yZC50ZW1wbGF0ZXMaHmdvb2dsZS9wcm90b2J1Zi9kdXJhdGlvbi5wcm90byLYAQoBUBIQCgNwaWQYASABKAlSA3BpZBISCgRuYW1lGAIgASgJUgRuYW1lEiAKC2Rlc2NyaXB0aW9uGAMgASgJUgtkZXNjcmlwdGlvbhIQCgNsYWIYBCADKAlSA2xhYhIgCgtpbnN0aXR1dGlvbhgFIAMoCVILaW5zdGl0dXRpb24SIAoLZGV2ZWxvcGVkQnkYBiADKAlSC2RldmVsb3BlZEJ5EjUKCGR1cmF0aW9uGAcgASgLMhkuZ29vZ2xlLnByb3RvYnVmLkR1cmF0aW9uUghkdXJhdGlvbkILWgl0ZW1wbGF0ZXNKvwgKCAoBDBIDAAASCggKAQISAwIAHgoJCgIDABIDBgAoCgkKAggLEgMEACAKCgoCBAASBAgAJQEKCgoDBAABEgMICAkKNQoEBAACABIDCwQTGiggSW50ZXJuYWwgUHJvdG9jb2wgSUQNCiBFeGFtcGxlOiBOUy0wMDENCgwKBQQAAgABEgMLCw4KDAoFBAACAAUSAwsECgoMCgUEAAIAAxIDCxESCjgKBAQAAgESAw8EFBorIFByb3RvY29sJ3MgbmFtZQ0KIEV4YW1wbGU6IG5lZ2F0aXZlIHN0YWluDQoMCgUEAAIBARIDDwsPCgwKBQQAAgEFEgMPBAoKDAoFBAACAQMSAw8SEwrWAQoEBAACAhIDFQQbGsgBIEJyaWVmIGRlc2NyaXB0aW9uIG9mIHRoZSBtZXRob2QNCiBFeGFtcGxlOg0KIDIgbWljcm8gbGl0ZXJzIG9mIHNhbXBsZSwgd2FpdCBmb3IgNjAgc2Vjb25kcywgYmxvdCB3aXRoIHBhcGVyIDMgdGltZXMsDQogMiBtaWNybyBsaXRlcnMgb2YgdXJhbnlsIGFjZXRhdGUsIHdhaXQgZm9yIDYwIHNlY29uZHMsIGJsb3Qgd2l0aCBwYXBlciAzIHRpbWVzLg0KDAoFBAACAgESAxULFgoMCgUEAAICBRIDFQQKCgwKBQQAAgIDEgMVGRoKXAoEBAACAxIDGQQcGk8gTGlzdCBvZiBsYWJzIGFzc29jaWF0ZWQgd2l0aCB0aGUgc2FtcGxlIGNvbGxlY3Rpb24NCiBFeGFtcGxlOiBbIERleHRlciBMYWJzIF0NCgwKBQQAAgMBEgMZFBcKDAoFBAACAwUSAxkNEwoMCgUEAAIDBBIDGQQMCgwKBQQAAgMDEgMZGhsKeQoEBAACBBIDHQQkGmwgTGlzdCBvZiBuYW1lIG9mIHRoZSBpbnN0aXR1dGlvbiBmcm9tIHRoZSBsYWJzIGludm9sdmVkIGluIHNhbXBsZSBjb2xsZWN0aW9uDQogRXhhbXBsZTogWyBDYXJ0b29uIE5ldHdvcmsgXQ0KDAoFBAACBAESAx0UHwoMCgUEAAIEBRIDHQ0TCgwKBQQAAgQEEgMdBAwKDAoFBAACBAMSAx0iIwpVCgQEAAIFEgMhBCQaSCBMaXN0IG9mIHBlb3BsZSB3aG8gZGV2ZWxvcGVkIHRoZSBwcm90b2NvbA0KIEV4YW1wbGU6IFsgQ2hhcmxpZSwgRG91ZyBdDQoMCgUEAAIFARIDIRQfCgwKBQQAAgUFEgMhDRMKDAoFBAACBQQSAyEEDAoMCgUEAAIFAxIDISIjCjEKBAQAAgYSAyQEKhokIEV4YW1wbGUgb2YgdXNpbmcgYSBzdGFuZGFyZCBpbXBvcnQNCgwKBQQAAgYBEgMkHSUKDAoFBAACBgUSAyQEHAoMCgUEAAIGAxIDJCgpYgZwcm90bzM=")
	if err != nil {
		t.Fatal(err)
	}
	bc := &RecordTemplateProto{
		Description:        "a description",
		FriendlyName:       "Research Protocol BC",
		DescriptorSetProto: b,
		// Required:           []int64{},
		// Recommended:        []int64{0xcafebabe, 0xdeadbeef},
	}

	bctx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "00000000deadbeef",
		},
	}
	cb := &RecordTemplateProto{
		Description:        "a description",
		FriendlyName:       "Research Protocol CB",
		DescriptorSetProto: b,
		// Required:           []int64{},
		// Recommended:        []int64{0xdeadbeef},
	}

	cbtx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "00000000cafebabe",
		},
	}

	err = datastore.Setup(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	bir, err := intakeRecordTemplate(cb, cbtx)
	if err != nil {
		t.Fatal("failed :(")
	}
	datastore.AutoBulk.Add(bir)

	bir, err = intakeRecordTemplate(bc, bctx)
	if err != nil {
		t.Fatal("failed :(")
	}
	datastore.AutoBulk.Add(bir)

	_, err = datastore.AutoBulk.Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_ = bir
}

func TestLoadTemplatesFromES(t *testing.T) {
	t.SkipNow()
	err := datastore.Setup(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	err = LoadTemplatesFromES(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestDescriptorFromProtobufJs(t *testing.T) {
	t.SkipNow()
	// dsc := []byte{10, 113, 10, 27, 111, 105, 112, 53, 95, 114, 101, 99, 111, 114, 100, 95,
	// 	116, 101, 109, 112, 108, 97, 116, 101, 115, 46, 112, 114, 111, 116, 111, 18,
	// 	21, 111, 105, 112, 53, 46, 114, 101, 99, 111, 114, 100, 46, 116, 101, 109,
	// 	112, 108, 97, 116, 101, 115, 34, 51, 10, 1, 80, 18, 14, 10, 6, 102, 114,
	// 	117, 105, 116, 115, 24, 1, 32, 3, 40, 9, 18, 17, 10, 9, 102, 105, 114,
	// 	115, 116, 78, 97, 109, 101, 24, 2, 32, 1, 40, 9, 18, 11, 10, 3, 97,
	// 	103, 101, 24, 3, 32, 1, 40, 5, 98, 6, 112, 114, 111, 116, 111, 51,
	// }

	dsc := []byte{10, 85, 10, 27, 111, 105, 112, 53, 95, 114, 101, 99, 111, 114, 100, 95, 116, 101, 109, 112, 108, 97, 116, 101, 115, 46, 112, 114, 111, 116, 111, 18, 21, 111, 105, 112, 53, 46, 114, 101, 99, 111, 114, 100, 46, 116, 101, 109, 112, 108, 97, 116, 101, 115, 34, 23, 10, 1, 80, 18, 18, 10, 10, 102, 114, 117, 105, 116, 115, 32, 114, 114, 114, 24, 1, 32, 3, 40, 9, 98, 6, 112, 114, 111, 116, 111, 51}

	bc := &RecordTemplateProto{
		Description:        "Test generated from protobuf.js",
		FriendlyName:       "Protobuf.js test",
		DescriptorSetProto: dsc,
	}

	bctx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "000000000badbabe",
		},
	}

	err := datastore.Setup(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	bir, err := intakeRecordTemplate(bc, bctx)
	if err != nil {
		t.Fatal("failed :(")
	}
	datastore.AutoBulk.Add(bir)

	_, err = datastore.AutoBulk.Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_ = bir
}

func TestEncodeRecordTemplate(t *testing.T) {
	t.SkipNow()
	fd := []byte{10, 79, 10, 27, 111, 105, 112, 53, 95, 114, 101, 99, 111, 114, 100, 95, 116, 101, 109, 112, 108, 97, 116, 101, 115, 46, 112, 114, 111, 116, 111, 18, 21, 111, 105, 112, 53, 46, 114, 101, 99, 111, 114, 100, 46, 116, 101, 109, 112, 108, 97, 116, 101, 115, 34, 17, 10, 1, 80, 18, 12, 10, 4, 116, 101, 115, 116, 24, 1, 32, 1, 40, 9, 98, 6, 112, 114, 111, 116, 111, 51}

	rt := &RecordTemplateProto{
		Description:        "description for test template",
		DescriptorSetProto: fd,
		FriendlyName:       "Test Template",
	}

	o5 := &OipFive{
		RecordTemplate: rt,
	}

	b, err := proto.Marshal(o5)
	if err != nil {
		panic(err)
	}

	o5b64 := base64.StdEncoding.EncodeToString(b)
	pubKey := "ofbB67gqjgaYi45u8Qk2U3hGoCmyZcgbN4"
	wif := "cRVa9rNx5N1YKBw8PhavegJPFCiYCfC4n8cYmdc3X1Y6TyFZGG4B"
	_ = wif
	fmt.Println(o5b64)
	// copy/paste ^^ to flo-qt to sign, copy/pasted result back here
	signatureB64 := "HwNyg/TsW2nDhkYfZlicrXrD29J2kgNpyKZMGP6b8GDaA9uTpSYyWK80ULVoxyDHhMSN9ogQj3jTnTQV0r9NYnw="

	sigBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		panic(err)
	}

	sm := &oipProto.SignedMessage{
		SerializedMessage: b,
		MessageType:       oipProto.MessageTypes_OIP05,
		PubKey:            []byte(pubKey),
		Signature:         sigBytes,
		SignatureType:     oipProto.SignatureTypes_Flo,
	}

	smBytes, err := proto.Marshal(sm)
	if err != nil {
		panic(err)
	}

	fmt.Println(base64.StdEncoding.EncodeToString(smBytes))
}

func TestP64(t *testing.T) {
	t.SkipNow()
	// 	p64 := "Cl0KBHJ5YW4SAm9rIlEKTwobb2lwNV9yZWNvcmRfdGVtcGxhdGVzLnByb3RvEhVvaXA1LnJlY29yZC50ZW1wbGF0ZXMiEQoBUBIMCgR3aGF0GAEgASgJYgZwcm90bzMQARgBIiJvZmJCNjdncWpnYVlpNDV1OFFrMlUzaEdvQ215WmNnYk40KkEgR2nZ8Qz3anwls8iQeIOTqIDlJdIQ91Zif6UaQN3lsccZoXo0jDvWQPgblgMSBME4FQJJm3dxgto0lXP1Im3HGQ=="
	p64 := "CoQBCoEBCg1UZXN0IFRlbXBsYXRlEh1kZXNjcmlwdGlvbiBmb3IgdGVzdCB0ZW1wbGF0ZSJRCk8KG29pcDVfcmVjb3JkX3RlbXBsYXRlcy5wcm90bxIVb2lwNS5yZWNvcmQudGVtcGxhdGVzIhEKAVASDAoEdGVzdBgBIAEoCWIGcHJvdG8zEAEYASIib2ZiQjY3Z3FqZ2FZaTQ1dThRazJVM2hHb0NteVpjZ2JONCpBHwNyg/TsW2nDhkYfZlicrXrD29J2kgNpyKZMGP6b8GDaA9uTpSYyWK80ULVoxyDHhMSN9ogQj3jTnTQV0r9NYnw="

	tx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "00000000deadbeef",
		},
	}

	events.Publish("sync:floData:p64", p64, tx)

	ch := make(chan struct{})
	<-ch
}

func TestUnmarshalSignedMessage(t *testing.T) {
	t.SkipNow()
	p64 := "Cg1UZXN0IFRlbXBsYXRlEh1kZXNjcmlwdGlvbiBmb3IgdGVzdCB0ZW1wbGF0ZSJRCk8KG29pcDVfcmVjb3JkX3RlbXBsYXRlcy5wcm90bxIVb2lwNS5yZWNvcmQudGVtcGxhdGVzIhEKAVASDAoEdGVzdBgBIAEoCWIGcHJvdG8z"

	b, err := base64.StdEncoding.DecodeString(p64)
	if err != nil {
		panic(err)
	}
	rtp := &RecordTemplateProto{}
	err = proto.Unmarshal(b, rtp)
	if err != nil {
		panic(err)
	}

	_ = rtp
}

func TestDecodeRecordTemplate(t *testing.T) {
	b, err := base64.StdEncoding.DecodeString("CskDCsYDChJTdGFyX1dhcnNfUHJvZmlsZXMSNVBlcnNvbmFsIHByb2ZpbGVzIG9mIHBlb3BsZSBpbiB0aGUgU3RhciBXYXJzIHVuaXZlcnNlIvgCCvUCChhvaXBQcm90b190ZW1wbGF0ZXMucHJvdG8SEm9pcFByb3RvLnRlbXBsYXRlcyK8AgoBUBIMCgRuYW1lGAEgASgJEhkKC2JpcnRoUGxhbmV0GAQgASgLMgRUeGlkEhEKCWJpcnRoZGF0ZRgFIAEoBBoTCgRUeGlkEgsKA3JhdxgBIAIoDCJZCgdGYWN0aW9uEg0KCVVOREVGSU5FRBAAEhsKF0ZhY3Rpb25fR0FMQUNUSUNfRU1QSVJFEAESEAoMRmFjdGlvbl9KRURJEAISEAoMRmFjdGlvbl9TSVRIEAMiigEKD0xpZ2h0c2FiZXJDb2xvchINCglVTkRFRklORUQQABIYChRMaWdodHNhYmVyQ29sb3JfQkxVRRABEhkKFUxpZ2h0c2FiZXJDb2xvcl9HUkVFThACEhcKE0xpZ2h0c2FiZXJDb2xvcl9SRUQQAxIaChZMaWdodHNhYmVyQ29sb3JfUFVSUExFEARiBnByb3RvMxABGAEiIkZUZFFKSkN0RVA3Wkp5cFhuMlJHeWRlYnpjRkxWZ0RLWFIqQSAaPYpQOROOWLRfV8ZfcHSGrnxFTx8MuIqJxPxJP6KQFxB4ZDfNiEB9DedmYNQtCeLU476geKZuUWeFg2zSrOFS")
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

	rt := &RecordTemplate{}
	err = decodeDescriptorSet(rt, o5.RecordTemplate.DescriptorSetProto, "8910cbc1923e6b64d4012b88b85703237630bab2083410a74fa1ff8e7ffca439")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(*rt)
}
