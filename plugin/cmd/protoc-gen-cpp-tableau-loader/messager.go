package main

import (
	"fmt"
	"strings"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// golbal container for record all proto filenames and messager names
var protofiles []string
var messagers []string

// generateMessager generates protobuf message wrapped classes
// which inherit from base class Messager.
func generateMessager(gen *protogen.Plugin, file *protogen.File) {
	protofiles = append(protofiles, file.GeneratedFilenamePrefix)
	generateHppFile(gen, file)
	generateCppFile(gen, file)
}

// generateHppFile generates a header file corresponding to a protobuf file.
func generateHppFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + "." + pcExt + ".h"
	g := gen.NewGeneratedFile(filename, "")
	generateFileHeader(gen, file, g)
	generateHppFileContent(gen, file, g)
	return g
}

// generateCppFile generates loader files related to protoconf files.
func generateCppFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + "." + pcExt + ".cc"
	g := gen.NewGeneratedFile(filename, "")
	generateFileHeader(gen, file, g)
	generateCppFileContent(gen, file, g)
	return g
}

// generateHppFileContent generates type definitions.
func generateHppFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P("#pragma once")
	g.P("#include <string>")
	g.P()
	g.P(`#include "`, "hub.", pcExt, `.h"`)
	g.P(`#include "`, file.GeneratedFilenamePrefix, ".", pbExt, `.h"`)
	g.P()

	g.P("namespace ", *namespace, " {")
	for _, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			genHppMessage(gen, file, g, message)
		}
	}
	g.P("}  // namespace ", *namespace)
}

// genHppMessage generates a message definition.
func genHppMessage(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, message *protogen.Message) {
	pkg := string(file.Desc.Package())
	cppFullName := strings.ReplaceAll(pkg, ".", "::") + "::" + string(message.Desc.Name())

	messagers = append(messagers, string(message.Desc.Name()))

	g.P("class ", message.Desc.Name(), " : public Messager {")
	g.P(" public:")
	g.P("  static const std::string& Name() { return kProtoName; };")
	g.P("  virtual bool Load(const std::string& dir, Format fmt) override;")
	g.P("  const ", cppFullName, "& Data() const { return data_; };")

	// syntactic sugar for accessing map items
	genHppMapGetters(1, nil, g, message.Desc)

	g.P()
	g.P(" private:")
	g.P("  static const std::string kProtoName;")
	g.P("  ", cppFullName, " data_;")
	g.P("};")
	g.P()
}

func genHppMapGetters(depth int, params []string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			keyParam := parseCppType(fd.MapKey()) + fmt.Sprintf(" key%d", depth)
			params = append(params, keyParam)
			g.P("  const ", parseCppType(fd.MapValue()), "* Get(", strings.Join(params, ", "), ") const;")
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genHppMapGetters(depth+1, params, g, fd.MapValue().Message())
			}
			continue
		}
	}
}

// generateCppFileContent generates type implementations.
func generateCppFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P(`#include "`, file.GeneratedFilenamePrefix, ".", pcExt, `.h"`)
	g.P()

	g.P("namespace ", *namespace, " {")
	for _, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			genCppMessage(gen, file, g, message)
		}
	}
	g.P("}  // namespace ", *namespace)
}

// genCppMessage generates a message implementation.
func genCppMessage(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, message *protogen.Message) {
	g.P("const std::string ", message.Desc.Name(), "::kProtoName = ", `"`, message.Desc.Name(), `";`)
	g.P("bool ", message.Desc.Name(), "::Load(const std::string& dir, Format fmt) { return LoadMessage(dir, data_, fmt); }")
	// syntactic sugar for accessing map items
	genCppMapGetters(1, nil, string(message.Desc.Name()), g, message.Desc)
}

func genCppMapGetters(depth int, params []string, messagerName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			keyType := parseCppType(fd.MapKey())
			keyName := fmt.Sprintf("key%d", depth)

			params = append(params, keyType+" "+keyName)

			g.P("const ", parseCppType(fd.MapValue()), "* ", messagerName, "::Get(", strings.Join(params, ", "), ") const {")

			var container string
			if depth == 1 {
				container = "data_." + string(fd.Name()) + "()"
			} else {
				container = "conf->" + string(fd.Name()) + "()"
				var findParams []string
				for i := 1; i < depth; i++ {
					findParams = append(findParams, fmt.Sprintf("key%d", i))
				}
				g.P("  const auto* conf = Get(", strings.Join(findParams, ", "), ");")
				g.P("  if (conf == nullptr) {")
				g.P("    return nullptr;")
				g.P("  }")
			}
			g.P("  auto iter = ", container, ".find(", keyName, ");")
			g.P("  if (iter == ", container, ".end()) {")
			g.P("    return nullptr;")
			g.P("  }")
			g.P("  return &iter->second;")
			g.P("}")
			g.P()

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genCppMapGetters(depth+1, params, messagerName, g, fd.MapValue().Message())
			}
			continue
		}
	}
}

// parseCppType converts a FieldDescriptor to C++ type string.
func parseCppType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		protoFullName := string(fd.Message().FullName())
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
		return "float_t"
	case protoreflect.DoubleKind:
		return "double_t"
	case protoreflect.StringKind, protoreflect.BytesKind:
		return "std::string"
	case protoreflect.MessageKind:
		protoFullName := string(fd.Message().FullName())
		return strings.ReplaceAll(protoFullName, ".", "::")
	// case protoreflect.GroupKind:
	// 	return "group"
	default:
		return fmt.Sprintf("<unknown:%d>", fd.Kind())
	}
}
