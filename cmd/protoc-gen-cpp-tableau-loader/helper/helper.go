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

func GenerateFileHeader(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, version string) {
	GenerateCommonHeader(gen, g, version)
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
}

func GenerateCommonHeader(gen *protogen.Plugin, g *protogen.GeneratedFile, version string) {
	g.P("// Code generated by protoc-gen-cpp-tableau-loader. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-cpp-tableau-loader v", version)
	g.P("// - protoc                        ", protocVersion(gen))
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

func NeedGenOrderedMap(md protoreflect.MessageDescriptor) bool {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	if wsOpts == nil || !wsOpts.OrderedMap {
		// Not an ordered map.
		return false
	}
	return true
}

func ParseIndexFieldName(fd protoreflect.FieldDescriptor) string {
	return escapeIdentifier(string(fd.Name()))
}

func ParseIndexFieldNameAsKeyStructFieldName(fd protoreflect.FieldDescriptor) string {
	if fd.IsList() {
		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		return escapeIdentifier(strcase.ToSnake(fdOpts.GetName()))
	}
	return ParseIndexFieldName(fd)
}

func ParseIndexFieldNameAsFuncParam(fd protoreflect.FieldDescriptor) string {
	return ParseIndexFieldNameAsKeyStructFieldName(fd)
}

// ParseCppType converts a FieldDescriptor to C++ type string.
func ParseCppType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		protoFullName := string(fd.Enum().FullName())
		return strings.ReplaceAll(protoFullName, ".", "::")
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32_t"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32_t"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64_t"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64_t"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.StringKind, protoreflect.BytesKind:
		return "std::string"
	case protoreflect.MessageKind:
		return ParseCppClassType(fd.Message())
	// case protoreflect.GroupKind:
	// 	return "group"
	default:
		return fmt.Sprintf("<unknown:%d>", fd.Kind())
	}
}

func ToConstRefType(cpptype string) string {
	if cpptype == "std::string" {
		return "const std::string&"
	}
	return cpptype
}

func ParseCppClassType(md protoreflect.MessageDescriptor) string {
	protoFullName := string(md.FullName())
	return strings.ReplaceAll(protoFullName, ".", "::")
}

type MapKey struct {
	Type string
	Name string
}

func AddMapKey(fd protoreflect.FieldDescriptor, keys []MapKey) []MapKey {
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
	keys = append(keys, MapKey{ParseCppType(fd.MapKey()), name})
	return keys
}

// GenGetParams generates function parameters, which are the names listed in the function's definition.
func GenGetParams(keys []MapKey) string {
	var params string
	for i, key := range keys {
		params += ToConstRefType(key.Type) + " " + key.Name
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
