package oip5

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/bitspill/flod/flojson"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/oipProto"

	_ "github.com/oipwg/oip/modules/oip"
)

// func TestIntakeRecordNormalize(t *testing.T) {
// 	t.SkipNow()
//
// 	txidDescriptor, err := desc.LoadMessageDescriptorForMessage(&oipProto.Txid{})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	txidMessageType, err := builder.FromMessage(txidDescriptor)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	p := builder.NewMessage("P").
// 		AddField(builder.NewField("i64", builder.FieldTypeInt64()).SetNumber(1)).
// 		AddField(builder.NewField("str", builder.FieldTypeString()).SetNumber(2)).
// 		AddField(builder.NewField("link",
// 			builder.FieldTypeMessage(txidMessageType)).SetNumber(3))
//
// 	f := builder.NewFile("p.proto").SetPackageName("oipProto.templates").AddMessage(p)
//
// 	fd, err := f.Build()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	dsp := desc.ToFileDescriptorSet(fd)
//
// 	b, err := proto.Marshal(dsp)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	bc := &RecordTemplateProto{
// 		Description:        "A",
// 		FriendlyName:       "A",
// 		DescriptorSetProto: b,
// 	}
// 	bctx := &datastore.TransactionData{
// 		Transaction: &flojson.TxRawResult{
// 			Txid: "10000000deadbeef",
// 		},
// 	}
//
// 	cp := builder.NewMessage("P").
// 		AddField(builder.NewField("ci64", builder.FieldTypeInt64()).SetNumber(1)).
// 		AddField(builder.NewField("cstr", builder.FieldTypeString()).SetNumber(2))
//
// 	cf := builder.NewFile("p.proto").SetPackageName("oipProto.templates").AddMessage(cp)
//
// 	cfd, err := cf.Build()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	cdsp := desc.ToFileDescriptorSet(cfd)
//
// 	cb, err := proto.Marshal(cdsp)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	cbr := &RecordTemplateProto{
// 		Description:        "B",
// 		FriendlyName:       "B",
// 		DescriptorSetProto: cb,
// 	}
//
// 	cbtx := &datastore.TransactionData{
// 		Transaction: &flojson.TxRawResult{
// 			Txid: "10000000cafebabe",
// 		},
// 	}
//
// 	err = datastore.Setup(context.Background())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	err = LoadTemplatesFromES(context.Background())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	bir, err := intakeRecordTemplate(cbr, cbtx)
// 	if err != nil {
// 		t.Fatal("failed :(")
// 	}
// 	datastore.AutoBulk.Add(bir)
//
// 	bir, err = intakeRecordTemplate(bc, bctx)
// 	if err != nil {
// 		t.Fatal("failed :(")
// 	}
// 	datastore.AutoBulk.Add(bir)
//
// 	n := &NormalizeRecordProto{
// 		MainTemplate: 0x10000000deadbeef,
// 		Fields: []*NormalField{
// 			{
// 				Name: "field1_i64",
// 				Path: []*Field{
// 					{
// 						Template: 0x10000000deadbeef,
// 					},
// 					{
// 						Tag:  1,
// 						Type: Field_TYPE_INT64,
// 					},
// 				},
// 			},
// 			{
// 				Name: "field2_str",
// 				Path: []*Field{
// 					{
// 						Template: 0x10000000deadbeef,
// 					},
// 					{
// 						Tag:  2,
// 						Type: Field_TYPE_STRING,
// 					},
// 				},
// 			},
// 			{
// 				Name: "field3_ci64",
// 				Path: []*Field{
// 					{
// 						Template: 0x10000000deadbeef,
// 					},
// 					{
// 						Tag:  3,
// 						Type: Field_TYPE_MESSAGE,
// 					},
// 					{
// 						Template: 0x10000000cafebabe,
// 					},
// 					{
// 						Tag:  1,
// 						Type: Field_TYPE_INT64,
// 					},
// 				},
// 			},
// 			{
// 				Name: "field4_cstr",
// 				Path: []*Field{
// 					{
// 						Template: 0x10000000deadbeef,
// 					},
// 					{
// 						Tag:  3,
// 						Type: Field_TYPE_MESSAGE,
// 					},
// 					{
// 						Template: 0x10000000cafebabe,
// 					},
// 					{
// 						Tag:  2,
// 						Type: Field_TYPE_STRING,
// 					},
// 				},
// 			},
// 		},
// 	}
//
// 	ntx := &datastore.TransactionData{
// 		Transaction: &flojson.TxRawResult{
// 			Txid: "10000000beefcafe",
// 		},
// 	}
//
// 	bir, err = intakeNormalize(n, ntx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	datastore.AutoBulk.Add(bir)
//
// 	p.SetName("tmpl_10000000DEADBEEF")
// 	nf, err := f.Build()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	md := nf.FindMessage("oipProto.templates.tmpl_10000000DEADBEEF")
// 	msgDetA := dynamic.NewMessage(md)
// 	msgDetA.SetFieldByNumber(1, 5)
// 	msgDetA.SetFieldByNumber(2, "hello")
// 	msgDetA.SetFieldByNumber(3, oipProto.TxidFromString("0000000000000001000000000000000000000000000000000000000000000000"))
//
// 	anyA, err := ptypes.MarshalAny(msgDetA)
//
// 	recA := &RecordProto{
// 		Details: &OipDetails{
// 			[]*any.Any{anyA},
// 		},
// 	}
// 	txRecA := &datastore.TransactionData{
// 		Transaction: &flojson.TxRawResult{
// 			Txid: "f000000000000000000000000000000000000000000000000000000000000000",
// 		},
// 	}
//
// 	cp.SetName("tmpl_10000000CAFEBABE")
// 	ncf, err := cf.Build()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	mdb := ncf.FindMessage("oipProto.templates.tmpl_10000000CAFEBABE")
// 	msgDetB := dynamic.NewMessage(mdb)
// 	msgDetB.SetFieldByNumber(1, 7)
// 	msgDetB.SetFieldByNumber(2, "world")
//
// 	anyB, err := ptypes.MarshalAny(msgDetB)
//
// 	recB := &RecordProto{
// 		Details: &OipDetails{
// 			[]*any.Any{anyB},
// 		},
// 	}
// 	txRecB := &datastore.TransactionData{
// 		Transaction: &flojson.TxRawResult{
// 			Txid: "0000000000000001000000000000000000000000000000000000000000000000",
// 		},
// 	}
//
// 	bir, err = intakeRecord(recB, txRecB)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	datastore.AutoBulk.Add(bir)
//
// 	bir, err = intakeRecord(recA, txRecA)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	datastore.AutoBulk.Add(bir)
//
// 	err = normalizeRecord(recA, txRecA)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	_, err = datastore.AutoBulk.Do(context.Background())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// }

func TestNormalizeToBlockchainFormat(t *testing.T) {
	// t.SkipNow()

	pubKey := []byte("ofbB67gqjgaYi45u8Qk2U3hGoCmyZcgbN4")
	wif := []byte("cRVa9rNx5N1YKBw8PhavegJPFCiYCfC4n8cYmdc3X1Y6TyFZGG4B")
	_ = wif

	txidDescriptor, err := desc.LoadMessageDescriptorForMessage(&oipProto.Txid{})
	if err != nil {
		t.Fatal(err)
	}

	txidMessageType, err := builder.FromMessage(txidDescriptor)
	if err != nil {
		t.Fatal(err)
	}

	planetMessage := builder.NewMessage("P").
		AddField(builder.NewField("mass", builder.FieldTypeInt64()).SetNumber(1)). // billions of kg
		AddField(builder.NewField("name", builder.FieldTypeString()).SetNumber(2)).
		AddField(builder.NewField("moons",
			builder.FieldTypeMessage(txidMessageType)).SetNumber(3).SetRepeated())

	planetFile := builder.NewFile("p.proto").SetPackageName("oipProto.templates").AddMessage(planetMessage)

	planetFileDescriptor, err := planetFile.Build()
	if err != nil {
		t.Fatal(err)
	}

	planetFileDescriptorSet := desc.ToFileDescriptorSet(planetFileDescriptor)

	planetDescriptorSetProtoBytes, err := proto.Marshal(planetFileDescriptorSet)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("planet descriptor")
	fmt.Println(base64.StdEncoding.EncodeToString(planetDescriptorSetProtoBytes))

	planetOipFive := &OipFive{
		RecordTemplate: &RecordTemplateProto{
			Description:        "A celestial body moving in an elliptical orbit around a star",
			FriendlyName:       "Planet",
			DescriptorSetProto: planetDescriptorSetProtoBytes,
		},
	}

	planetOipFiveBytes, err := proto.Marshal(planetOipFive)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("planet record template to sign")
	fmt.Println(base64.StdEncoding.EncodeToString(planetOipFiveBytes))

	planetSignatureBytes, err := base64.StdEncoding.DecodeString("H6W33G9zsnM4eD8q38qXTi2Y4xBT0JxCj7w2knK3cunfOXCQtQ14d7tmzgqwpDmrrV4QHN2BSpUfnbsz9HWx6JM=")
	if err != nil {
		t.Fatal(err)
	}

	planetSignedMessage := &oipProto.SignedMessage{
		SerializedMessage: planetOipFiveBytes,
		MessageType:       oipProto.MessageTypes_OIP05,
		PubKey:            pubKey,
		SignatureType:     oipProto.SignatureTypes_Flo,
		Signature:         planetSignatureBytes,
	}

	planetSignedMessageBytes, err := proto.Marshal(planetSignedMessage)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("signed planet record template")
	fmt.Println("p64:" + base64.StdEncoding.EncodeToString(planetSignedMessageBytes))

	planetTemplatetTx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "d7309b2a2738a99fe087b032cd245056142374109b83ad2438976c0722ec9c37",
		},
	}

	moonMessage := builder.NewMessage("P").
		AddField(builder.NewField("mass", builder.FieldTypeInt64()).SetNumber(1)). // billions of kg
		AddField(builder.NewField("name", builder.FieldTypeString()).SetNumber(2))

	moonFile := builder.NewFile("p.proto").SetPackageName("oipProto.templates").AddMessage(moonMessage)

	moonFileDescriptor, err := moonFile.Build()
	if err != nil {
		t.Fatal(err)
	}

	moonFileDescriptorSet := desc.ToFileDescriptorSet(moonFileDescriptor)

	moonDescriptorSetProtoBytes, err := proto.Marshal(moonFileDescriptorSet)
	if err != nil {
		t.Fatal(err)
	}

	moonOipFive := &OipFive{
		RecordTemplate: &RecordTemplateProto{
			Description:        "A natural satellite of any planet",
			FriendlyName:       "Moon",
			DescriptorSetProto: moonDescriptorSetProtoBytes,
		},
	}

	moonOipFiveBytes, err := proto.Marshal(moonOipFive)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("moon record template to sign")
	fmt.Println(base64.StdEncoding.EncodeToString(moonOipFiveBytes))

	moonSignatureBytes, err := base64.StdEncoding.DecodeString("H4uL4EhLcVM/pT9XAMcMcDx4Oz3/CmTFpC6FRxGysnHbOJaIZ522Gu676QnXjzcx4kyAgwqXgMKlwLCHvr0qdkA=")
	if err != nil {
		t.Fatal(err)
	}

	moonSignedMessage := &oipProto.SignedMessage{
		SerializedMessage: moonOipFiveBytes,
		MessageType:       oipProto.MessageTypes_OIP05,
		PubKey:            pubKey,
		SignatureType:     oipProto.SignatureTypes_Flo,
		Signature:         moonSignatureBytes,
	}

	moonSignedMessageBytes, err := proto.Marshal(moonSignedMessage)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("signed moon record template")
	fmt.Println("p64:" + base64.StdEncoding.EncodeToString(moonSignedMessageBytes))

	moonTemplatetTx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "370840eecb3c27ca834a199f6d456849eaa187489d008173427772215725bf3e",
		},
	}

	moonMessage.SetName("tmpl_370840EE")
	moonf, err := moonFile.Build()
	if err != nil {
		t.Fatal(err)
	}
	moonMd := moonf.FindMessage("oipProto.templates.tmpl_370840EE")
	moonDet := dynamic.NewMessage(moonMd)
	moonDet.SetFieldByNumber(1, int64(73420000000000)) // billions of kg
	moonDet.SetFieldByNumber(2, "Luna")

	moonAny, err := ptypes.MarshalAny(moonDet)

	moonRecord := &OipFive{
		Record: &RecordProto{
			Details: &OipDetails{
				[]*any.Any{moonAny},
			},
		},
	}

	moonRecordBytes, err := proto.Marshal(moonRecord)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("moon record to sign")
	fmt.Println(base64.StdEncoding.EncodeToString(moonRecordBytes))

	moonRecordSignatureBytes, err := base64.StdEncoding.DecodeString("ICGHFFS+Q+r1CYblmMp4D9l9fdyLTZ7baDRntOTwhaaFJebEB2WFgQiLoimRqu52f21NeDMgCsT3X6IGTg1f5ls=")
	if err != nil {
		t.Fatal(err)
	}

	moonRecordSignedMessage := &oipProto.SignedMessage{
		SerializedMessage: moonRecordBytes,
		MessageType:       oipProto.MessageTypes_OIP05,
		PubKey:            pubKey,
		SignatureType:     oipProto.SignatureTypes_Flo,
		Signature:         moonRecordSignatureBytes,
	}

	moonRecordSignedMessageBytes, err := proto.Marshal(moonRecordSignedMessage)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("signed moon record")
	fmt.Println("p64:" + base64.StdEncoding.EncodeToString(moonRecordSignedMessageBytes))

	moonRecordTx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "37a65ffa5e85e3bce6eb1f80568efc0757ba10726da0b0f9f283df1f50368d68",
		},
	}

	planetMessage.SetName("tmpl_D7309B2A")
	nf, err := planetFile.Build()
	if err != nil {
		t.Fatal(err)
	}
	planetMd := nf.FindMessage("oipProto.templates.tmpl_D7309B2A")
	planetDet := dynamic.NewMessage(planetMd)
	planetDet.SetFieldByNumber(1, int64(5972000000000000)) // billions of kg
	planetDet.SetFieldByNumber(2, "Earth")
	lunaRef := oipProto.TxidFromString("37a65ffa5e85e3bce6eb1f80568efc0757ba10726da0b0f9f283df1f50368d68")
	planetDet.AddRepeatedFieldByNumber(3, lunaRef)

	planetAny, err := ptypes.MarshalAny(planetDet)

	planetRecord := &OipFive{
		Record: &RecordProto{
			Details: &OipDetails{
				[]*any.Any{planetAny},
			},
		},
	}

	planetRecordBytes, err := proto.Marshal(planetRecord)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("planet record to sign")
	fmt.Println(base64.StdEncoding.EncodeToString(planetRecordBytes))

	planetRecordSignatureBytes, err := base64.StdEncoding.DecodeString("H5g7AE5nYplhf4uPVYWl/oLeKJWYgz7iFu5UIpEFLmvIEeBvcGDM+4THa7N/iBjU6GPdE9dd/1wvw4KpyaJdCbM=")
	if err != nil {
		t.Fatal(err)
	}

	planetRecordSignedMessage := &oipProto.SignedMessage{
		SerializedMessage: planetRecordBytes,
		MessageType:       oipProto.MessageTypes_OIP05,
		PubKey:            pubKey,
		SignatureType:     oipProto.SignatureTypes_Flo,
		Signature:         planetRecordSignatureBytes,
	}

	planetRecordSignedMessageBytes, err := proto.Marshal(planetRecordSignedMessage)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("signed planet record")
	fmt.Println("p64:" + base64.StdEncoding.EncodeToString(planetRecordSignedMessageBytes))

	planetRecordTx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "ddd130e03ed49bb07d6286a1b96d3e9fe37aa05fe1daa404fbeff967f5635293",
		},
	}

	planetNormalize := &OipFive{
		Normalize: &NormalizeRecordProto{
			MainTemplate: 0xd7309b2a,
			Fields: []*NormalField{
				{
					Name: "planet_mass",
					Path: []*Field{
						{
							Template: 0xd7309b2a,
						},
						{
							Tag:  1,
							Type: Field_TYPE_INT64,
						},
					},
				},
				{
					Name: "planet_name",
					Path: []*Field{
						{
							Template: 0xd7309b2a,
						},
						{
							Tag:  2,
							Type: Field_TYPE_STRING,
						},
					},
				},
				{
					Name: "moon_name",
					Path: []*Field{
						{
							Template: 0xd7309b2a,
						},
						{
							Tag:  3,
							Type: Field_TYPE_MESSAGE,
						},
						{
							Template: 0x370840ee,
						},
						{
							Tag:  2,
							Type: Field_TYPE_STRING,
						},
					},
				},
				{
					Name: "moon_mass",
					Path: []*Field{
						{
							Template: 0xd7309b2a,
						},
						{
							Tag:  3,
							Type: Field_TYPE_MESSAGE,
						},
						{
							Template: 0x370840ee,
						},
						{
							Tag:  1,
							Type: Field_TYPE_INT64,
						},
					},
				},
			},
		},
	}

	planetNormalizeBytes, err := proto.Marshal(planetNormalize)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("planet normalize to sign")
	fmt.Println(base64.StdEncoding.EncodeToString(planetNormalizeBytes))

	planetNormalizeSignatureBytes, err := base64.StdEncoding.DecodeString("Hx/sNuprdJXddYKrh5ei/zm3mRsvQiCp7CDONS5bj3uLc6xjXaw7lTKglNO5p2aOiUv57RnrluPg2OCGR/Z1oLw=")
	if err != nil {
		t.Fatal(err)
	}

	planetNormalizeSignedMessage := &oipProto.SignedMessage{
		SerializedMessage: planetNormalizeBytes,
		MessageType:       oipProto.MessageTypes_OIP05,
		PubKey:            pubKey,
		SignatureType:     oipProto.SignatureTypes_Flo,
		Signature:         planetNormalizeSignatureBytes,
	}

	planetNormalizeSignedMessageBytes, err := proto.Marshal(planetNormalizeSignedMessage)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("signed planet normalize")
	fmt.Println("p64:" + base64.StdEncoding.EncodeToString(planetNormalizeSignedMessageBytes))

	planetNormalizeTx := &datastore.TransactionData{
		Transaction: &flojson.TxRawResult{
			Txid: "ddd130e03ed49bb07d6286a1b96d3e9fe37aa05fe1daa404fbeff967f5635293",
		},
	}

	_ = planetNormalizeTx
	_ = planetRecordTx
	_ = planetTemplatetTx
	_ = moonRecordTx
	_ = moonTemplatetTx
}
