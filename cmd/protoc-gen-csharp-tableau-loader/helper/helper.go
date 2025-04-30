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

// ParseCsharpType converts a FieldDescriptor to C# type string.
func ParseCsharpType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		fullname := string(fd.Enum().FullName())
		seps := strings.Split(fullname, ".")
		seps[0] = strcase.ToCamel(seps[0])
		for i := 2; i < len(seps); i++ {
			seps[i] = "Types." + seps[i]
		}
		return strings.Join(seps, ".")
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "long"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "ulong"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.StringKind, protoreflect.BytesKind:
		return "string"
	case protoreflect.MessageKind:
		fullname := string(fd.Message().FullName())
		seps := strings.Split(fullname, ".")
		seps[0] = strcase.ToCamel(seps[0])
		for i := 2; i < len(seps); i++ {
			seps[i] = "Types." + seps[i]
		}
		return strings.Join(seps, ".")
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
	// name = escapeIdentifier(name)
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
	keys = append(keys, MapKey{ParseCsharpType(fd.MapKey()), strcase.ToLowerCamel(name)})
	return keys
}

// GenGetParams generates function parameters, which are the names listed in the function's definition.
func GenGetParams(keys []MapKey) string {
	var params string
	for i, key := range keys {
		params += key.Type + " " + key.Name
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
