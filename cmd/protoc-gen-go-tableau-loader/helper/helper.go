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
	return escapeIdentifier(fieldName)
}

// ParseGoType converts a FieldDescriptor to its Go type.
// returns string if fd is scalar type, and protogen.GoIdent if fd is enum or message type.
func ParseGoType(gen *protogen.Plugin, g *protogen.GeneratedFile, fd protoreflect.FieldDescriptor) string {
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
		return g.QualifiedGoIdent(FindEnumGoIdent(gen, fd.Enum()))
	case protoreflect.MessageKind:
		return g.QualifiedGoIdent(FindMessageGoIdent(gen, fd.Message()))
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

// ParseOrderedIndexKeyType converts a FieldDescriptor to its treemap key type.
// fd must be an ordered type, or a message which can be converted to an ordered type.
func ParseOrderedIndexKeyType(gen *protogen.Plugin, g *protogen.GeneratedFile, fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
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
	case protoreflect.EnumKind:
		return g.QualifiedGoIdent(FindEnumGoIdent(gen, fd.Enum()))
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

func ParseMapFieldNameAsKeyStructFieldName(fd protoreflect.FieldDescriptor) string {
	opts := fd.Options().(*descriptorpb.FieldOptions)
	fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
	name := fdOpts.GetKey()
	if fd.MapValue().Kind() == protoreflect.MessageKind {
		valueFd := fd.MapValue().Message().Fields().Get(0)
		name = string(valueFd.Name())
	}
	return strcase.ToCamel(name)
}

func ParseMapFieldNameAsFuncParam(fd protoreflect.FieldDescriptor) string {
	fieldName := ParseMapFieldNameAsKeyStructFieldName(fd)
	if fieldName == "" {
		return fieldName
	}
	return escapeIdentifier(fieldName)
}

func ParseMapValueType(gen *protogen.Plugin, g *protogen.GeneratedFile, fd protoreflect.FieldDescriptor) string {
	valueType := ParseGoType(gen, g, fd.MapValue())
	if fd.MapValue().Kind() == protoreflect.MessageKind {
		return "*" + valueType
	}
	return valueType
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

func ParseLeveledMapPrefix(md protoreflect.MessageDescriptor, mapFd protoreflect.FieldDescriptor) string {
	if mapFd.MapValue().Kind() == protoreflect.MessageKind {
		localMsgProtoName := strings.TrimPrefix(string(mapFd.MapValue().Message().FullName()), string(md.FullName())+".")
		return strings.ReplaceAll(localMsgProtoName, ".", "_")
	}
	return mapFd.MapValue().Kind().String()
}

type MapKey struct {
	Type          string
	Name          string
	FieldName     string                       // multi-column index only (may be deduplicated, e.g., "Id" → "Id3")
	OrigFieldName string                       // original FieldName before deduplication (empty if not renamed)
	Fd            protoreflect.FieldDescriptor // the map field descriptor this key belongs to
}

type MapKeySlice []MapKey

// AddMapKey appends a new map key to the slice, automatically deduplicating
// both Name (used as function parameter names) and FieldName (used as struct
// field names in LevelIndex key structs).
//
// Deduplication is needed because different map levels may share the same key
// name. For example, given the following nested proto maps where country_map
// and item_map both use "ID" as their key name:
//
//	message Fruit4Conf {
//	    map<int32, Fruit> fruit_map = 1;          // key field: "FruitType"
//	    message Fruit {
//	        map<int32, Country> country_map = 2;  // key field: "ID"
//	        message Country {
//	            map<int32, Item> item_map = 3;    // key field: "ID"  ← same name!
//	        }
//	    }
//	}
//
// Without dedup, the generated LevelIndex key struct would have duplicate
// field names, causing a compile error:
//
//	type Fruit4Conf_LevelIndex_Fruit_Country_ItemKey struct {
//	    FruitType int32 // key of protoconf.Fruit4Conf.fruit_map
//	    Id        int32 // key of protoconf.Fruit4Conf.Fruit.country_map
//	    Id        int32 // key of protoconf.Fruit4Conf.Fruit.Country.item_map — COMPILE ERROR!
//	}
//
// With dedup, the conflicting name gets a numeric suffix (the 1-based position
// of the new key in the slice), producing valid Go code:
//
//	type Fruit4Conf_LevelIndex_Fruit_Country_ItemKey struct {
//	    FruitType int32 // key of protoconf.Fruit4Conf.fruit_map
//	    Id        int32 // key of protoconf.Fruit4Conf.Fruit.country_map
//	    Id3       int32 // key of protoconf.Fruit4Conf.Fruit.Country.item_map (renamed from Id)
//	}
func (s MapKeySlice) AddMapKey(newKey MapKey) MapKeySlice {
	if newKey.Name == "" {
		newKey.Name = fmt.Sprintf("key%d", len(s)+1)
	}
	// Deduplicate Name (used as function parameter, e.g., "id" → "id3").
	for _, key := range s {
		if key.Name == newKey.Name {
			newKey.Name = fmt.Sprintf("%s%d", newKey.Name, len(s)+1)
			break
		}
	}
	// Deduplicate FieldName (used as struct field, e.g., "Id" → "Id3").
	// This is only relevant for multi-column indexes that generate LevelIndex
	// key structs; single-column indexes leave FieldName empty.
	if newKey.FieldName != "" {
		for _, key := range s {
			if key.FieldName == newKey.FieldName {
				newKey.OrigFieldName = newKey.FieldName
				newKey.FieldName = fmt.Sprintf("%s%d", newKey.FieldName, len(s)+1)
				break
			}
		}
	}
	return append(s, newKey)
}

// GenGetParams generates function parameters, which are the names listed in the function's definition.
func (s MapKeySlice) GenGetParams() string {
	var params []string
	for _, key := range s {
		params = append(params, key.Name+" "+key.Type)
	}
	return strings.Join(params, ", ")
}

// GenGetArguments generates function arguments, which are the real values passed to the function.
func (s MapKeySlice) GenGetArguments() string {
	var params []string
	for _, key := range s {
		params = append(params, key.Name)
	}
	return strings.Join(params, ", ")
}
