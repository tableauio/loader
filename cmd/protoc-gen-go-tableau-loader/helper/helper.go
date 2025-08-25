package helper

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ParseIndexFieldName(gen *protogen.Plugin, fd protoreflect.FieldDescriptor) string {
	md := fd.ContainingMessage()
	msg := FindMessage(gen, md)
	for _, field := range msg.Fields {
		if field.Desc == fd {
			return field.GoName
		}
	}
	panic(fmt.Sprintf("unknown fd: %v", fd))
}

func ParseIndexFieldNameAsKeyStructFieldName(gen *protogen.Plugin, fd protoreflect.FieldDescriptor) string {
	if fd.IsList() {
		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		return strcase.ToCamel(fdOpts.GetName())
	}
	return ParseIndexFieldName(gen, fd)
}

func ParseIndexFieldNameAsFuncParam(gen *protogen.Plugin, fd protoreflect.FieldDescriptor) string {
	fieldName := ParseIndexFieldNameAsKeyStructFieldName(gen, fd)
	if fieldName == "" {
		return fieldName
	}
	return escapeIdentifier(strings.ToLower(fieldName[:1]) + fieldName[1:])
}

// ParseGoType converts a FieldDescriptor to its Go type.
// returns string if fd is scalar type, and protogen.GoIdent if fd is enum or message type.
func ParseGoType(gen *protogen.Plugin, fd protoreflect.FieldDescriptor) any {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.EnumKind:
		return FindEnumGoIdent(gen, fd.Enum())
	case protoreflect.MessageKind:
		return FindMessageGoIdent(gen, fd.Message())
	// case protoreflect.GroupKind:
	// 	return "group"
	default:
		panic(fmt.Sprintf("unknown kind: %d", fd.Kind()))
	}
}

// ParseMapKeyType converts a FieldDescriptor to its map key type.
// fd must be an comparable type.
func ParseMapKeyType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.EnumKind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.StringKind:
		return "string"
	default:
		panic(fmt.Sprintf("unsupported kind: %d", fd.Kind()))
	}
}

// ParseOrderedMapKeyType converts a FieldDescriptor to its treemap key type.
// fd must be an ordered type, or a message which can be converted to an ordered type.
func ParseOrderedMapKeyType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.EnumKind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.MessageKind:
		switch fd.Message().FullName() {
		case "google.protobuf.Timestamp", "google.protobuf.Duration":
			return "int64"
		default:
		}
		fallthrough
	default:
		panic(fmt.Sprintf("unsupported kind: %d", fd.Kind()))
	}
}

func FindMessage(gen *protogen.Plugin, md protoreflect.MessageDescriptor) *protogen.Message {
	if file, ok := gen.FilesByPath[md.ParentFile().Path()]; ok {
		return FindMessageByDescriptor(file.Messages, md)
	}
	return nil
}

func FindMessageByDescriptor(messages []*protogen.Message, md protoreflect.MessageDescriptor) *protogen.Message {
	for _, message := range messages {
		if message.Desc.FullName() == md.FullName() {
			return message
		}
		// Recursively search nested messages
		if nestedMessage := FindMessageByDescriptor(message.Messages, md); nestedMessage != nil {
			return nestedMessage
		}
	}
	return nil
}

func FindMessageGoIdent(gen *protogen.Plugin, md protoreflect.MessageDescriptor) protogen.GoIdent {
	msg := FindMessage(gen, md)
	if msg == nil {
		panic(fmt.Sprintf("unknown message: %s", md.FullName()))
	}
	return msg.GoIdent
}

func FindEnum(gen *protogen.Plugin, ed protoreflect.EnumDescriptor) *protogen.Enum {
	if file, ok := gen.FilesByPath[ed.ParentFile().Path()]; ok {
		if enum := FindEnumByDescriptor(file.Enums, ed); enum != nil {
			return enum
		}
		return FindEnumFromMessageByDescriptor(file.Messages, ed)
	}
	return nil
}

func FindEnumByDescriptor(enums []*protogen.Enum, ed protoreflect.EnumDescriptor) *protogen.Enum {
	for _, enum := range enums {
		if enum.Desc.FullName() == ed.FullName() {
			return enum
		}
	}
	return nil
}

func FindEnumFromMessageByDescriptor(messages []*protogen.Message, ed protoreflect.EnumDescriptor) *protogen.Enum {
	for _, message := range messages {
		if enum := FindEnumByDescriptor(message.Enums, ed); enum != nil {
			return enum
		}
		// Recursively search nested messages
		if nestedEnum := FindEnumFromMessageByDescriptor(message.Messages, ed); nestedEnum != nil {
			return nestedEnum
		}
	}
	return nil
}

func FindEnumGoIdent(gen *protogen.Plugin, ed protoreflect.EnumDescriptor) protogen.GoIdent {
	enum := FindEnum(gen, ed)
	if enum == nil {
		panic(fmt.Sprintf("unknown enum: %s", ed.FullName()))
	}
	return enum.GoIdent
}

func GetTypeEmptyValue(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "false"
	// case protoreflect.EnumKind:
	// 	protoFullName := string(fd.Message().FullName())
	// 	return strings.ReplaceAll(protoFullName, ".", "_")
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "0"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "0.0"
	case protoreflect.StringKind:
		return `""`
	case protoreflect.BytesKind, protoreflect.MessageKind:
		return "nil"
	// case protoreflect.GroupKind:
	// 	return "group"
	default:
		return fmt.Sprintf("<unknown:%d>", fd.Kind())
	}
}

type MapKey struct {
	Type string
	Name string
}

func AddMapKey(gen *protogen.Plugin, fd protoreflect.FieldDescriptor, keys []MapKey) []MapKey {
	opts := fd.Options().(*descriptorpb.FieldOptions)
	fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
	name := fdOpts.GetKey()
	if fd.MapValue().Kind() == protoreflect.MessageKind {
		valueFd := fd.MapValue().Message().Fields().Get(0)
		name = string(valueFd.Name())
	}
	name = escapeIdentifier(name)
	if name == "" {
		name = fmt.Sprintf("key%d", len(keys)+1)
	} else {
		for _, key := range keys {
			if key.Name == name {
				// rewrite to avoid name confict
				name = fmt.Sprintf("%s%d", name, len(keys)+1)
				break
			}
		}
	}
	keys = append(keys, MapKey{ParseMapKeyType(fd.MapKey()), name})
	return keys
}

// GenGetParams generates function parameters, which are the names listed in the function's definition.
func GenGetParams(keys []MapKey) string {
	var params string
	for i, key := range keys {
		params += key.Name + " " + key.Type
		if i != len(keys)-1 {
			params += ", "
		}
	}
	return params
}

// GenGetArguments generates function arguments, which are the real values passed to the function.
func GenGetArguments(keys []MapKey) string {
	var params string
	for i, key := range keys {
		params += key.Name
		if i != len(keys)-1 {
			params += ", "
		}
	}
	return params
}
