package oip5_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/bitspill/flod/chaincfg"
	"github.com/bitspill/flod/floec"
	"github.com/bitspill/flosig"
	"github.com/bitspill/floutil"
	patch "github.com/bitspill/protoPatch"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/oipwg/proto/go/pb_oip"
	"github.com/oipwg/proto/go/pb_oip5"
	"github.com/oipwg/proto/go/pb_oip5/pb_templates"
)

func TestCreatePublisherRegistration(t *testing.T) {
	pk, err := floec.NewPrivateKey(floec.S256())
	if err != nil {
		t.Fatal(err)
	}

	wif, err := floutil.NewWIF(pk, &chaincfg.TestNet3Params, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(wif.String())

	addr, err := floutil.NewAddressPubKeyHash(floutil.Hash160(wif.SerializePubKey()), &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(addr.EncodeAddress())

	regPubDet := &pb_templates.Tmpl_433C2783{
		Name:         "test publisher",
		FloBip44XPub: "",
	}

	det, err := pb_oip5.CreateOipDetails(regPubDet)
	if err != nil {
		t.Fatal(err)
	}

	rec := &pb_oip5.RecordProto{
		Tags:        nil,
		Payment:     nil,
		Details:     det,
		Permissions: nil,
	}

	o5 := &pb_oip5.OipFive{
		Record: rec,
	}

	b, err := proto.Marshal(o5)
	if err != nil {
		t.Fatal(err)
	}

	b64 := base64.StdEncoding.EncodeToString(b)
	fmt.Println(b64)

	sig64, err := flosig.SignMessagePk(b64, "Florincoin", pk, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sig64)

	sig, err := base64.StdEncoding.DecodeString(sig64)
	if err != nil {
		t.Fatal(err)
	}

	sm := &pb_oip.SignedMessage{
		SerializedMessage: b,
		MessageType:       pb_oip.MessageTypes_OIP05,
		SignatureType:     pb_oip.SignatureTypes_Flo,
		PubKey:            []byte(addr.EncodeAddress()),
		Signature:         sig,
	}

	pb, err := proto.Marshal(sm)
	if err != nil {
		t.Fatal(err)
	}

	p64 := base64.StdEncoding.EncodeToString(pb)
	fmt.Println(p64)

	// Broken Publish (forgot OipFive)
	// cUgks5vY6GMvbUt6TudpvVCpv3xqYW6M41sQR8sUwxBLVrEWf1Gs
	// oJwZ5NLxJyw9gSGx7eXQPBwiFQQD3vJZ5o
	// OkoKSAo0dHlwZS5nb29nbGVhcGlzLmNvbS9vaXBQcm90by50ZW1wbGF0ZXMudG1wbF80MzNDMjc4MxIQCg50ZXN0IHB1Ymxpc2hlcg==
	// IA2Sok+X71RiRS5CUW+JCKcFBWr2WGqydGJO4lG8Q5uyHq30avPF7dC09sjm131GulCch6HQ9zuX+afeLmwgGyI=
	// Ckw6SgpICjR0eXBlLmdvb2dsZWFwaXMuY29tL29pcFByb3RvLnRlbXBsYXRlcy50bXBsXzQzM0MyNzgzEhAKDnRlc3QgcHVibGlzaGVyEAEYASIib0p3WjVOTHhKeXc5Z1NHeDdlWFFQQndpRlFRRDN2Slo1bypBIA2Sok+X71RiRS5CUW+JCKcFBWr2WGqydGJO4lG8Q5uyHq30avPF7dC09sjm131GulCch6HQ9zuX+afeLmwgGyI=
	// 1a33713c4a3d4922c5b767a4372bce5b0b7f20b373dc41e21bb5be3db9e9c56a

	// Registered as "test publisher"
	// wif: cVka3cstn95iDafAWoLyKFm7wvrRoY6b5kVXPH1B8gVqYb28x9YR
	// addr: oRnb8WFGejy6KffxrtT24moF9fn2HAY7yk
	// o5: Ekw6SgpICjR0eXBlLmdvb2dsZWFwaXMuY29tL29pcFByb3RvLnRlbXBsYXRlcy50bXBsXzQzM0MyNzgzEhAKDnRlc3QgcHVibGlzaGVy
	// sig: ILUGMReJiTXJYKe+UmtG8mAZDLDzZPQKQ82emjyD1vctPMjLNuezUIogUt0Yeg5aj6mrmWg/XTIqaZSfS9obQ9A=
	// p64:Ck4STDpKCkgKNHR5cGUuZ29vZ2xlYXBpcy5jb20vb2lwUHJvdG8udGVtcGxhdGVzLnRtcGxfNDMzQzI3ODMSEAoOdGVzdCBwdWJsaXNoZXIQARgBIiJvUm5iOFdGR2VqeTZLZmZ4cnRUMjRtb0Y5Zm4ySEFZN3lrKkEgtQYxF4mJNclgp75Sa0byYBkMsPNk9ApDzZ6aPIPW9y08yMs257NQiiBS3Rh6DlqPqauZaD9dMipplJ9L2htD0A==
	// txid: 4a059effa20389f2be9bfad9308f4a46b4c2bfaf02dd65e68f113db1669fba81
}

func TestCreateEditTestPublisher(t *testing.T) {
	wifStr := "cVka3cstn95iDafAWoLyKFm7wvrRoY6b5kVXPH1B8gVqYb28x9YR"
	addrStr := "oRnb8WFGejy6KffxrtT24moF9fn2HAY7yk"

	wif, err := floutil.DecodeWIF(wifStr)
	if err != nil {
		t.Fatal(err)
	}

	regPubDet := &pb_templates.Tmpl_433C2783{
		Name: "edited publisher",
	}

	det, err := pb_oip5.CreateOipDetails(regPubDet)
	if err != nil {
		t.Fatal(err)
	}

	newValues := &pb_oip5.RecordProto{
		Tags:        nil,
		Payment:     nil,
		Details:     det,
		Permissions: nil,
	}

	p := &patch.Patch{
		NewValues: newValues,
		Ops: []patch.Op{
			{
				[]patch.Step{
					{
						Tag:      7, // details
						Action:   patch.ActionStepInto,
						SrcIndex: 1,
						DstIndex: 1,
					},
					{
						Tag:    1, // publisher name
						Action: patch.ActionReplace,
					},
				},
			},
		},
	}

	pp, err := patch.ToProto(p)
	if err != nil {
		t.Fatal(err)
	}

	edit := &pb_oip5.EditProto{
		Reference: pb_oip.TxidFromString("4a059effa20389f2be9bfad9308f4a46b4c2bfaf02dd65e68f113db1669fba81"),
		Patch:     pp.(*patch.ProtoPatch),
	}

	o5 := &pb_oip5.OipFive{
		RecordTemplate: nil,
		Record:         nil,
		Normalize:      nil,
		Transfer:       nil,
		Deactivate:     nil,
		Edit:           edit,
	}

	b, err := proto.Marshal(o5)
	if err != nil {
		t.Fatal(err)
	}

	b64 := base64.StdEncoding.EncodeToString(b)
	fmt.Println(b64)

	sig64, err := flosig.SignMessagePk(b64, "Florincoin", wif.PrivKey, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sig64)

	sig, err := base64.StdEncoding.DecodeString(sig64)
	if err != nil {
		t.Fatal(err)
	}

	sm := &pb_oip.SignedMessage{
		SerializedMessage: b,
		MessageType:       pb_oip.MessageTypes_OIP05,
		SignatureType:     pb_oip.SignatureTypes_Flo,
		PubKey:            []byte(addrStr),
		Signature:         sig,
	}

	pb, err := proto.Marshal(sm)
	if err != nil {
		t.Fatal(err)
	}

	p64 := base64.StdEncoding.EncodeToString(pb)
	fmt.Println(p64)
}

func TestEditTest(t *testing.T) {
	regPubDet := &pb_templates.Tmpl_433C2783{
		Name:         "test publisher",
		FloBip44XPub: "",
	}

	det, err := pb_oip5.CreateOipDetails(regPubDet)
	if err != nil {
		t.Fatal(err)
	}

	rec := &pb_oip5.RecordProto{
		Tags:        nil,
		Payment:     nil,
		Details:     det,
		Permissions: nil,
	}

	newRegPubDet := &pb_templates.Tmpl_433C2783{
		Name: "edited publisher",
	}

	newDet, err := pb_oip5.CreateOipDetails(newRegPubDet)
	if err != nil {
		t.Fatal(err)
	}

	newRec := &pb_oip5.RecordProto{
		Tags:        nil,
		Payment:     nil,
		Details:     newDet,
		Permissions: nil,
	}

	p := patch.Patch{
		NewValues: newRec,
		Ops: []patch.Op{
			{
				[]patch.Step{
					{
						Tag:      7, // details
						Action:   patch.ActionStepInto,
						SrcIndex: 1,
						DstIndex: 1,
					},
					{
						Tag:    1, // publisher name
						Action: patch.ActionReplace,
					},
				},
			},
		},
	}

	res, err := patch.ApplyPatch(p, rec)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println((&jsonpb.Marshaler{Indent: " "}).MarshalToString(res))
}
