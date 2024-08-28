package helper

import (
	"fmt"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func NeedGenOrderedMap(md protoreflect.MessageDescriptor) bool {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	if wsOpts == nil || !wsOpts.OrderedMap {
		// Not an ordered map.
		return false
	}
	return true
}

// ParseGoType converts a FieldDescriptor to Go type string.
func ParseGoType(gen *protogen.Plugin, fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	// case protoreflect.EnumKind:
	// 	protoFullName := string(fd.Message().FullName())
	// 	return strings.ReplaceAll(protoFullName, ".", "_")
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
	case protoreflect.MessageKind:
		if file, ok := gen.FilesByPath[fd.Message().ParentFile().Path()]; ok {
			message := FindMessageByDescriptor(file.Messages, fd.Message())
			if message != nil {
				return message.GoIdent.GoName
			}
		}
		return fmt.Sprintf("<not found:%s>", fd.Message().FullName())
	// case protoreflect.GroupKind:
	// 	return "group"
	default:
		return fmt.Sprintf("<unknown:%d>", fd.Kind())
	}
}

func FindMessageByDescriptor(messages []*protogen.Message, desc protoreflect.MessageDescriptor) *protogen.Message {
	for _, message := range messages {
		if message.Desc.FullName() == desc.FullName() {
			return message
		}
		// Recursively search nested messages
		if nestedMessage := FindMessageByDescriptor(message.Messages, desc); nestedMessage != nil {
			return nestedMessage
		}
	}
	return nil
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
		return ""
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
	keys = append(keys, MapKey{ParseGoType(gen, fd.MapKey()), name})
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
