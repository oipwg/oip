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
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

func intakeRecordTemplate(rt *RecordTemplateProto, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
	attr := logger.Attrs{"txid": tx.Transaction.Txid}
	log.Info("oip5 ", attr)

	if len(tx.Transaction.Txid) < 16 {
		log.Error("invalid txid", attr)
		return nil, errors.New("invalid txid")
	}
	strIdent := tx.Transaction.Txid[:16]
	ident, err := strconv.ParseUint(strIdent, 16, 64)
	if err != nil {
		attr["err"] = err
		log.Error("unable to decode txid", attr)
		return nil, errors.New("unable to decode txid")
	}
	rt.Identifier = ident

	tmpl := &RecordTemplate{
		FriendlyName: rt.FriendlyName,
		Description:  rt.Description,
		Identifier:   rt.Identifier,
		Recommended:  rt.Recommended,
		Required:     rt.Required,
	}

	err = decodeDescriptorSet(tmpl, rt.GetDescriptorSetProto(), tx.Transaction.Txid)
	if err != nil {
		attr["err"] = err
		log.Error("unable to decode descriptor set", attr)
		return nil, errors.New("unable to decode descriptor set")
	}

	elRt := elRecordTemplate{
		Template:          templateCache[ident],
		FileDescriptorSet: base64.StdEncoding.EncodeToString(rt.GetDescriptorSetProto()),
		Meta: TMeta{
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

func decodeDescriptorSet(rt *RecordTemplate, descriptorSetProto []byte, txid string) error {
	defer func() {
		if r := recover(); r != nil {

		}
	}()

	attr := logger.Attrs{"txid": txid}
	var dsp = &descriptor.FileDescriptorSet{}
	err := proto.Unmarshal(descriptorSetProto, dsp)
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
	newName := "tmpl_" + txid[:16]
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
		log.Error("unable to find message oip5.record.templates.P", attr)
		return errors.New("unable to find message oip5.record.templates.P")
	}
	err = messageBuilder.TrySetName(newName)
	if err != nil {
		attr["err"] = err
		attr["newName"] = newName
		attr["oldName"] = messageBuilder.GetName()
		log.Error("unable to set message name", attr)
		return errors.New("unable to set message name")
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

	rt.MessageDescriptor = message
	rt.FileDescriptor = file
	rt.Name = newName

	rt.MessageType = TemplateMessageFactory.GetKnownTypeRegistry().GetKnownType(newName)

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

var templateCache = make(map[uint64]*RecordTemplate)
var TemplateMessageFactory = dynamic.NewMessageFactoryWithDefaults()
var TemplateAnyResolver = anyResolver{upstreamAny: dynamic.AnyResolver(TemplateMessageFactory)}

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
	FileDescriptor *desc.FileDescriptor `json:"-"`
	// Message type
	MessageType reflect.Type `json:"-"`
	// Populated by oipd with the unique identifier for this type
	Identifier uint64 `json:"identifier"`
	// List of unique template identifiers recommended for use with this template
	Recommended []uint64 `json:"recommended,omitempty"`
	// List of unique template identifiers required for use with this template
	Required []uint64 `json:"required,omitempty"`
}

type elRecordTemplate struct {
	Template          *RecordTemplate `json:"template"`
	FileDescriptorSet string          `json:"file_descriptor_set"`
	Meta              TMeta           `json:"meta"`
}

type TMeta struct {
	Block       int64                      `json:"block"`
	BlockHash   string                     `json:"block_hash"`
	Deactivated bool                       `json:"deactivated"`
	Time        int64                      `json:"time"`
	Tx          *datastore.TransactionData `json:"tx"`
	Txid        string                     `json:"txid"`
	Type        string                     `json:"type"`
}

func (rt *RecordTemplate) CreateNewMessage() proto.Message {
	if rt.MessageType.Kind() == reflect.Ptr {
		return reflect.New(rt.MessageType.Elem()).Interface().(proto.Message)
	} else {
		return reflect.New(rt.MessageType).Elem().Interface().(proto.Message)
	}
}

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
	hexId := strings.TrimPrefix(id, "oip5.record.templates.tmpl_")
	if len(hexId) == 16 {
		ident, err := strconv.ParseUint(hexId, 16, 64)
		if err == nil {
			if t, ok := templateCache[ident]; ok {
				return t.CreateNewMessage(), nil
			}
		}
	}

	m := TemplateMessageFactory.GetKnownTypeRegistry().CreateIfKnown(id)
	if m == nil {
		return nil, fmt.Errorf("unknown message type %q", id)
	}

	return m, nil
}

func LoadTemplatesFromES(ctx context.Context) error {
	searchService := datastore.Client().Search(datastore.Index("oip5_templates")).Type("_doc")

	res, err := searchService.Do(ctx)
	if err != nil {
		return err
	}

	templates := res.Each(reflect.TypeOf(elRecordTemplate{}))

	for _, value := range templates {
		tmpl := value.(elRecordTemplate)

		b, err := base64.StdEncoding.DecodeString(tmpl.FileDescriptorSet)
		err = decodeDescriptorSet(tmpl.Template, b, tmpl.Meta.Txid)
		if err != nil {
			return err
		}
	}

	return nil
}
