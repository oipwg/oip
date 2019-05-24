package main

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/oipwg/oip/oipProto"
)

func main() {
	mp := &oipProto.MultiPart{
		RawData:     []byte("hello world"),
		CountParts:  2,
		CurrentPart: 1,
		Reference: &oipProto.Txid{
			Raw: []byte{0xde, 0xad, 0xbe, 0xef},
		},
	}
	fmt.Println(mp)

	dynMp, err := dynamic.AsDynamicMessage(mp)
	if err != nil {
		panic(err)
	}
	fmt.Println(dynMp)

	fields := dynMp.GetKnownFields()
	for _, f := range fields {
		// fmt.Println(f)
		if f.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			fmt.Println(f.String())
			fmt.Println(f.AsFieldDescriptorProto().GetTypeName())
			fmt.Println(f.GetName())
		}
	}

}
