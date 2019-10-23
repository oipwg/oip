package oip5

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

func CreateOipDetails(details ...proto.Message) (*OipDetails, error) {
	ret := &OipDetails{}

	for _, v := range details {
		vAny, err := ptypes.MarshalAny(v)
		if err != nil {
			return nil, err
		}
		ret.Details = append(ret.Details, vAny)
	}
	return ret, nil
}

func (m *OipDetails) MarshalJSONPB(marsh *jsonpb.Marshaler) ([]byte, error) {
	var detMap = make(map[string]*json.RawMessage)

	// "@type": "type.googleapis.com/oipProto.templates.tmpl_deadbeef",
	// oipProto.templates.tmpl_deadbeef
	for _, detAny := range m.Details {
		name, err := ptypes.AnyMessageName(detAny)
		if err != nil {
			return nil, err
		}

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

		tmplName := strings.TrimPrefix(name, "oipProto.templates.")
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
		if len(k) == 13 && strings.HasPrefix(k, "tmpl_") {
			k = "type.googleapis.com/oipProto.templates." + k
		}

		var jsonFields map[string]*json.RawMessage
		if err := json.Unmarshal(*v, &jsonFields); err != nil {
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

type o5AnyResolver struct{}

func (r *o5AnyResolver) Resolve(typeUrl string) (proto.Message, error) {
	mname := typeUrl
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}

	// try default behavior first
	mt := proto.MessageType(mname)
	if mt != nil {
		return reflect.New(mt.Elem()).Interface().(proto.Message), nil
	}

	return CreateNewMessage(mname)
}
