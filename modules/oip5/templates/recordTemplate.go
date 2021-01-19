package templates

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/oipwg/proto/go/pb_oip"
	"github.com/oipwg/proto/go/pb_oip5/pb_templates"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/datastore"
)

func IntakeRecordTemplate(rt *pb_templates.RecordTemplateProto, pubKey []byte, tx *datastore.TransactionData) (*elastic.BulkIndexRequest, error) {
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
		Txid:              tx.Transaction.Txid,
		SignedBy:          string(pubKey),
		FriendlyName:      rt.FriendlyName,
		Description:       rt.Description,
		Identifier:        rt.Identifier,
		Extends:           rt.Extends,
		FileDescriptorSet: base64.StdEncoding.EncodeToString(rt.DescriptorSetProto),
	}

	err = DecodeDescriptorSet(tmpl, rt.DescriptorSetProto)
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

func DecodeDescriptorSet(rt *RecordTemplate, descriptorSetProto []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic within DecodeDescriptorSet %s", r)
		}
	}()

	attr := logger.Attrs{"txid": rt.Txid}
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
	newName := "tmpl_" + strings.ToUpper(rt.Txid[:8])
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
					txidDescriptor, err := desc.LoadMessageDescriptorForMessage(&pb_oip.Txid{})
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
		addProtoType(fileMsgType, rt.Txid)
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
	// ToDo: test if dynamic templates override compiled templates
	// fqn := fileMsgType.GetFullyQualifiedName()
	// ktrMsgType := ktr.GetKnownType(fqn)
	// if ktrMsgType != nil {
	// 	log.Info("message type already known", logger.Attrs{"fqn": fqn, "txid": txid})
	// 	return
	// }
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

	Txid     string `json:"-"`
	SignedBy string `json:"-"`
}

type elRecordTemplate struct {
	Template *RecordTemplate `json:"template"`
	Meta     TMeta           `json:"meta"`
}

type TMeta struct {
	Block     int64                      `json:"block,omitempty"`
	BlockHash string                     `json:"block_hash,omitempty"`
	SignedBy  string                     `json:"signed_by,omitempty"`
	Time      int64                      `json:"time,omitempty"`
	Tx        *datastore.TransactionData `json:"-"`
	Txid      string                     `json:"txid,omitempty"`
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
		tmpl.Template.SignedBy = tmpl.Meta.SignedBy
		tmpl.Template.Txid = tmpl.Meta.Txid

		b, err := base64.StdEncoding.DecodeString(tmpl.Template.FileDescriptorSet)
		if err != nil {
			return err
		}
		err = DecodeDescriptorSet(tmpl.Template, b)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetTemplate(txid string) (*RecordTemplate, error) {
	if len(txid) < 8 {
		log.Error("invalid txid", logger.Attrs{"txid": txid})
		return nil, errors.New("invalid txid")
	}
	strIdent := txid[:8]
	ident, err := strconv.ParseUint(strIdent, 16, 32)
	if err != nil {
		log.Error("invalid txid", logger.Attrs{"txid": txid})
		return nil, errors.New("invalid txid")
	}
	tmpl := templateCache[uint32(ident)]
	return tmpl, nil
}

func EditTemplate(tmpl *RecordTemplate, newRaw string, editTxid string) error {
	b, err := base64.StdEncoding.DecodeString(newRaw)
	if err != nil {
		return errors.New("unable to decode raw template")
	}

	newVal := &pb_templates.RecordTemplateProto{}
	err = proto.Unmarshal(b, newVal)
	if err != nil {
		return errors.New("unable to decode template proto for edit")
	}

	if newVal.FriendlyName != "" {
		tmpl.FriendlyName = newVal.FriendlyName
	}
	if newVal.Description != "" {
		tmpl.Description = newVal.Description
	}
	if newVal.Extends != nil {
		tmpl.Extends = newVal.Extends
	}

	tmpl.FileDescriptorSet = base64.StdEncoding.EncodeToString(newVal.DescriptorSetProto)

	err = DecodeDescriptorSet(tmpl, newVal.DescriptorSetProto)
	if err != nil {
		log.Error("unable to decode descriptor set", tmpl.Name)
		return errors.New("unable to decode descriptor set")
	}

	elRt := elRecordTemplate{
		Template: tmpl,
	}

	bir := elastic.NewBulkUpdateRequest().
		Index(datastore.Index("oip5_templates")).
		Type("_doc").
		Id(tmpl.Txid).
		Doc(elRt)

	datastore.AutoBulk.Add(bir)

	bur := elastic.NewBulkUpdateRequest().
		Index(datastore.Index("oip5_edit")).
		Type("_doc").
		Id(editTxid).
		Doc(MetaApplied{Applied{true}})

	datastore.AutoBulk.Add(bur)

	return nil
}

type Applied struct {
	Applied bool `json:"applied"`
}
type MetaApplied struct {
	Meta Applied `json:"meta"`
}
