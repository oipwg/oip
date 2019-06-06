package oip5

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/oipProto"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	_ = datastore.RegisterMapping("oip5_templates", "oip5_templates.json")
}

func intakeRecordTemplate(rt *RecordTemplateProto, pubKey []byte, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	attr := logger.Attrs{"txid": tx.Transaction.Txid}
	log.Info("oip5 ", attr)

	if len(tx.Transaction.Txid) < 8 {
		log.Error("invalid txid", attr)
		return nil, errors.New("invalid txid")
	}
	strIdent := tx.Transaction.Txid[:8]
	ident, err := strconv.ParseUint(strIdent, 16, 32)
	if err != nil {
		attr["err"] = err
		log.Error("unable to decode txid", attr)
		return nil, errors.New("unable to decode txid")
	}
	rt.Identifier = uint32(ident)

	tmpl := &RecordTemplate{
		FriendlyName:      rt.FriendlyName,
		Description:       rt.Description,
		Identifier:        rt.Identifier,
		Extends:           rt.Extends,
		FileDescriptorSet: base64.StdEncoding.EncodeToString(rt.GetDescriptorSetProto()),
	}

	err = decodeDescriptorSet(tmpl, rt.GetDescriptorSetProto(), tx.Transaction.Txid)
	if err != nil {
		attr["err"] = err
		log.Error("unable to decode descriptor set", attr)
		return nil, errors.New("unable to decode descriptor set")
	}

	elRt := elRecordTemplate{
		Template: templateCache[uint32(ident)],
		Meta: TMeta{
			SignedBy:  string(pubKey),
			Tx:        tx,
			Time:      tx.Transaction.Time,
			Txid:      tx.Transaction.Txid,
			BlockHash: tx.BlockHash,
			Block:     tx.Block,
		},
	}

	bir := elastic.NewBulkIndexRequest().
		Index(datastore.Index("oip5_templates")).
		Type("_doc").
		Id(tx.Transaction.Txid).
		Doc(elRt)

	return bir, nil
}

func decodeDescriptorSet(rt *RecordTemplate, descriptorSetProto []byte, txid string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic within decodeDescriptorSet %s", r)
		}
	}()

	attr := logger.Attrs{"txid": txid}
	var dsp = &descriptor.FileDescriptorSet{}
	err = proto.Unmarshal(descriptorSetProto, dsp)
	if err != nil {
		attr["err"] = err
		log.Error("unable to unmarshal template descriptor", attr)
		return errors.New("unable to unmarshal template descriptor")
	}
	fd, err := desc.CreateFileDescriptorFromSet(dsp)
	if err != nil {
		attr["err"] = err
		log.Error("unable to create file descriptor", attr)
		return errors.New("unable to create file descriptor")
	}
	fileBuilder, err := builder.FromFile(fd)
	if err != nil {
		attr["err"] = err
		log.Error("unable to create builder", attr)
		return errors.New("unable to create builder")
	}
	newName := "tmpl_" + strings.ToUpper(txid[:8])
	err = fileBuilder.TrySetName(newName + ".proto")
	if err != nil {
		attr["err"] = err
		attr["newName"] = newName + ".proto"
		attr["oldName"] = fileBuilder.GetName()
		log.Error("unable to set file name", attr)
		return errors.New("unable to set file name")
	}
	messageBuilder := fileBuilder.GetMessage("P")
	if messageBuilder == nil {
		log.Error("unable to find message oipProto.templates.P", attr)
		return errors.New("unable to find message oipProto.templates.P")
	}
	err = messageBuilder.TrySetName(newName)
	if err != nil {
		attr["err"] = err
		attr["newName"] = newName
		attr["oldName"] = messageBuilder.GetName()
		log.Error("unable to set message name", attr)
		return errors.New("unable to set message name")
	}

	children := messageBuilder.GetChildren()
	for _, child := range children {
		if fb, ok := child.(*builder.FieldBuilder); ok {
			t := fb.GetType()
			if t.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				n := t.GetTypeName()
				if strings.HasPrefix(n, "oipProto.") && strings.HasSuffix(n, ".Txid") {
					txidDescriptor, err := desc.LoadMessageDescriptorForMessage(&oipProto.Txid{})
					if err != nil {
						attr["err"] = err
						log.Error("unable to load txid descriptor", attr)
						return err
					}
					txidMessageType, err := builder.FromMessage(txidDescriptor)
					if err != nil {
						attr["err"] = err
						log.Error("unable to create txid message type", attr)
						return err
					}
					fb.SetType(builder.FieldTypeMessage(txidMessageType))
					ok := messageBuilder.TryRemoveField(fb.GetName())
					if ok {
						err := messageBuilder.TryAddField(fb)
						if err != nil {
							attr["err"] = err
							log.Error("unable to add txid field", attr)
							return err
						}
					} else {
						log.Error("unable to remove txid field", attr)
						return errors.New("unable to remove txid field")
					}
				}
			}
		}
		if mb, ok := child.(*builder.MessageBuilder); ok {
			if mb.GetName() == "Txid" {
				ok := messageBuilder.TryRemoveNestedMessage("Txid")
				if !ok {
					log.Error("unable to remove nested Txid Type", attr)
					return errors.New("unable to remove nested Txid Type")
				}
			}
		}
	}

	message, err := messageBuilder.Build()
	if err != nil {
		attr["err"] = err
		log.Error("unable to build message descriptor", attr)
		return errors.New("unable to build message descriptor")
	}
	file, err := fileBuilder.Build()
	if err != nil {
		attr["err"] = err
		log.Error("unable to build file descriptor", attr)
		return errors.New("unable to build file descriptor")
	}
	for _, fileMsgType := range file.GetMessageTypes() {
		addProtoType(fileMsgType, txid)
	}

	if !strings.HasPrefix(message.GetFullyQualifiedName(), "oipProto.templates.") {
		attr["fqn"] = message.GetFullyQualifiedName()
		log.Error("missing required package", attr)
		return errors.New("missing required package")
	}

	rt.MessageDescriptor = message
	// rt.FileDescriptor = file
	rt.Name = newName

	// rt.MessageType = TemplateMessageFactory.GetKnownTypeRegistry().GetKnownType(message.GetFullyQualifiedName())

	templateCache[rt.Identifier] = rt
	return nil
}

func addProtoType(fileMsgType *desc.MessageDescriptor, txid string) {
	ktr := TemplateMessageFactory.GetKnownTypeRegistry()
	fqn := fileMsgType.GetFullyQualifiedName()
	ktrMsgType := ktr.GetKnownType(fqn)
	if ktrMsgType != nil {
		log.Info("message type already known", logger.Attrs{"fqn": fqn, "txid": txid})
		return
	}
	ktr.AddKnownType(dynamic.NewMessageWithMessageFactory(fileMsgType, TemplateMessageFactory))
}

var templateCache = make(map[uint32]*RecordTemplate)
var TemplateMessageFactory = dynamic.NewMessageFactoryWithDefaults()

// var TemplateAnyResolver = anyResolver{upstreamAny: dynamic.AnyResolver(TemplateMessageFactory)}

type RecordTemplate struct {
	// Human readable name to quickly identify type (non-unique)
	FriendlyName string `json:"friendly_name,omitempty"`
	// Generated name
	Name string `json:"name"`
	// Description of the purpose behind this new type
	Description string `json:"description,omitempty"`
	// Message
	MessageDescriptor *desc.MessageDescriptor `json:"-"`
	// File
	// FileDescriptor *desc.FileDescriptor `json:"-"`
	FileDescriptorSet string `json:"file_descriptor_set"`
	// Message type
	// MessageType reflect.Type `json:"-"`
	// Populated by oipd with the unique identifier for this type
	Identifier uint32 `json:"identifier"`
	// List of unique template identifiers recommended for use with this template
	Extends []uint32 `json:"extends,omitempty"`
}

type elRecordTemplate struct {
	Template *RecordTemplate `json:"template"`
	Meta     TMeta           `json:"meta"`
}

type TMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	SignedBy  string                     `json:"signed_by"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"-"`
	Txid      string                     `json:"txid"`
}

// func (rt *RecordTemplate) CreateNewMessage() proto.Message {
// // 	if rt.MessageType == nil {
// // 		log.Error("nil message type", logger.Attrs{"rt.ident": uint64(rt.Identifier), "rt.sIdent": rt.Identifier})
// // 		return nil
// // 	}
// //
// // 	if rt.MessageType.Kind() == reflect.Ptr {
// // 		return reflect.New(rt.MessageType.Elem()).Interface().(proto.Message)
// // 	} else {
// // 		return reflect.New(rt.MessageType).Elem().Interface().(proto.Message)
// // 	}
// // }

type anyResolver struct {
	upstreamAny jsonpb.AnyResolver
}

func (r anyResolver) Resolve(typeUrl string) (proto.Message, error) {
	m, err := CreateNewMessage(typeUrl)
	if err != nil {
		return m, nil
	}

	return r.upstreamAny.Resolve(typeUrl)
}

func CreateNewMessage(id string) (proto.Message, error) {
	hexId := strings.TrimPrefix(id, "oipProto.templates.tmpl_")
	if len(hexId) == 8 {
		ident, err := strconv.ParseUint(hexId, 16, 32)
		if err == nil {
			if t, ok := templateCache[uint32(ident)]; ok {
				msg := dynamic.NewMessageWithMessageFactory(t.MessageDescriptor, TemplateMessageFactory)
				return msg, nil
			}
		}
	}

	m := TemplateMessageFactory.GetKnownTypeRegistry().CreateIfKnown(id)
	if m == nil {
		return nil, fmt.Errorf("unknown message type %q", id)
	}

	if dm, ok := m.(*dynamic.Message); ok {
		if dm.GetMessageDescriptor() == nil {
			return nil, fmt.Errorf("unable to create dynamic type %s", id)
		}
	}

	return m, nil
}

func LoadTemplatesFromES(ctx context.Context) error {
	searchService := datastore.Client().Search(datastore.Index("oip5_templates")).Type("_doc").Size(10000)

	res, err := searchService.Do(ctx)
	if err != nil {
		return err
	}

	templates := res.Each(reflect.TypeOf(elRecordTemplate{}))

	for _, value := range templates {
		tmpl := value.(elRecordTemplate)

		b, err := base64.StdEncoding.DecodeString(tmpl.Template.FileDescriptorSet)
		err = decodeDescriptorSet(tmpl.Template, b, tmpl.Meta.Txid)
		if err != nil {
			return err
		}
	}

	return nil
}
