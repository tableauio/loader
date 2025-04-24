package main

import (
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/options"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// generateMessager generates protobuf message wrapped classes
// which inherit from base class Messager.
func generateMessager(gen *protogen.Plugin, file *protogen.File) {
	generateHppFile(gen, file)
	generateCppFile(gen, file)
}

// generateHppFile generates a header file corresponding to a protobuf file.
func generateHppFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + "." + pcExt + ".h"
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, file, g, version)
	g.P()
	generateHppFileContent(gen, file, g)
	return g
}

// generateCppFile generates loader files related to protoconf files.
func generateCppFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + "." + pcExt + ".cc"
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, file, g, version)
	g.P()
	generateCppFileContent(gen, file, g)
	return g
}

// generateHppFileContent generates type definitions.
func generateHppFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P("#pragma once")
	g.P("#include <string>")
	g.P()
	g.P(`#include "`, "messager.", pcExt, `.h"`)
	g.P(`#include "`, file.GeneratedFilenamePrefix, ".", pbExt, `.h"`)
	g.P()

	g.P("namespace ", *namespace, " {")
	var fileMessagers []string
	for _, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			genHppMessage(gen, file, g, message)
			messagerName := string(message.Desc.Name())
			fileMessagers = append(fileMessagers, messagerName)
		}
	}
	g.P("}  // namespace ", *namespace)
	g.P()

	// Generate aliases for all messagers.
	pkg := string(file.Desc.Package())
	pbNamespace := strings.ReplaceAll(pkg, ".", "::")
	g.P("namespace ", pbNamespace, " {")
	g.P("// Here are some type aliases for easy use.")
	for _, messager := range fileMessagers {
		g.P("using ", messager, *messagerSuffix, " = ", *namespace, "::", messager, ";")
	}
	g.P("}  // namespace ", pbNamespace)
}

// genHppMessage generates a message definition.
func genHppMessage(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, message *protogen.Message) {
	pkg := string(file.Desc.Package())
	cppFullName := strings.ReplaceAll(pkg, ".", "::") + "::" + string(message.Desc.Name())
	messagerFullName := string(message.Desc.FullName())
	indexDescriptors := index.ParseIndexDescriptor(gen, message.Desc)

	g.P("class ", message.Desc.Name(), " : public Messager {")
	g.P(" public:")
	g.P("  static const std::string& Name() { return kProtoName; }")
	g.P("  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) override;")
	g.P("  const ", cppFullName, "& Data() const { return data_; }")
	g.P("  const google::protobuf::Message* Message() const override { return &data_; }")
	g.P()

	if options.NeedGenOrderedMap(message.Desc, options.LangCPP) || options.NeedGenIndex(message.Desc, options.LangCPP) {
		g.P(" private:")
		g.P("  virtual bool ProcessAfterLoad() override final;")
		g.P()
	}

	// syntactic sugar for accessing map items
	genHppMapGetters(1, nil, g, message.Desc)
	g.P()
	g.P(" private:")
	g.P("  static const std::string kProtoName;")
	g.P("  ", cppFullName, " data_;")
	if options.NeedGenOrderedMap(message.Desc, options.LangCPP) {
		g.P()
		genHppOrderedMapGetters(g, message.Desc, 1, nil, messagerFullName)
	}
	if options.NeedGenIndex(message.Desc, options.LangCPP) {
		g.P()
		genHppIndexFinders(g, indexDescriptors)
	}
	g.P("};")
	g.P()
}

func genHppMapGetters(depth int, keys []helper.MapKey, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			if depth == 1 {
				g.P(" public:")
			}
			keys = helper.AddMapKey(fd, keys)
			g.P("  const ", helper.ParseCppType(fd.MapValue()), "* Get(", helper.GenGetParams(keys), ") const;")
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genHppMapGetters(depth+1, keys, g, fd.MapValue().Message())
			}
			break
		}
	}
}

// generateCppFileContent generates type implementations.
func generateCppFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P(`#include "`, file.GeneratedFilenamePrefix, ".", pcExt, `.h"`)
	g.P()
	g.P(`#include "hub.pc.h"`)
	g.P()

	g.P("namespace ", *namespace, " {")
	for _, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			genCppMessage(gen, g, message)
		}
	}
	g.P("}  // namespace ", *namespace)
}

// genCppMessage generates a message implementation.
func genCppMessage(gen *protogen.Plugin, g *protogen.GeneratedFile, message *protogen.Message) {
	messagerName := string(message.Desc.Name())
	messagerFullName := string(message.Desc.FullName())
	indexDescriptors := index.ParseIndexDescriptor(gen, message.Desc)

	g.P("const std::string ", messagerName, "::kProtoName = ", `"`, messagerName, `";`)
	g.P()
	g.P("bool ", messagerName, "::Load(const std::string& dir, Format fmt, const LoadOptions* options /* = nullptr */) {")
	g.P("  tableau::util::TimeProfiler profiler;")
	g.P("  bool loaded = LoadMessage(data_, dir, fmt, options);")
	g.P("  bool ok = loaded ? ProcessAfterLoad() : false;")
	g.P("  stats_.duration = profiler.Elapse();")
	g.P("  return ok;")
	g.P("}")
	g.P()

	if options.NeedGenOrderedMap(message.Desc, options.LangCPP) || options.NeedGenIndex(message.Desc, options.LangCPP) {
		g.P("bool ", messagerName, "::ProcessAfterLoad() {")
		if options.NeedGenOrderedMap(message.Desc, options.LangCPP) {
			genCppOrderedMapLoader(g, message.Desc, 1, messagerFullName)
		}
		if options.NeedGenIndex(message.Desc, options.LangCPP) {
			genCppIndexLoader(g, indexDescriptors)
		}
		g.P("  return true;")
		g.P("}")
		g.P()
	}

	// syntactic sugar for accessing map items
	genCppMapGetters(g, message.Desc, 1, nil, messagerName)
	genCppOrderedMapGetters(g, message.Desc, 1, nil, messagerName, messagerFullName)
	if options.NeedGenIndex(message.Desc, options.LangCPP) {
		genCppIndexFinders(g, indexDescriptors, messagerName)
		g.P()
	}
}

func genCppMapGetters(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerName string) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			keys = helper.AddMapKey(fd, keys)
			g.P("const ", helper.ParseCppType(fd.MapValue()), "* ", messagerName, "::Get(", helper.GenGetParams(keys), ") const {")

			var container string
			if depth == 1 {
				container = "data_." + string(fd.Name()) + "()"
			} else {
				container = "conf->" + string(fd.Name()) + "()"
				prevKeys := keys[:len(keys)-1]
				g.P("  const auto* conf = Get(", helper.GenGetArguments(prevKeys), ");")
				g.P("  if (conf == nullptr) {")
				g.P("    return nullptr;")
				g.P("  }")
			}
			lastKeyName := keys[len(keys)-1].Name
			g.P("  auto iter = ", container, ".find(", lastKeyName, ");")
			g.P("  if (iter == ", container, ".end()) {")
			g.P("    return nullptr;")
			g.P("  }")
			g.P("  return &iter->second;")
			g.P("}")
			g.P()

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genCppMapGetters(g, fd.MapValue().Message(), depth+1, keys, messagerName)
			}
			break
		}
	}
}

func getNextLevelMapFD(fd protoreflect.FieldDescriptor) protoreflect.FieldDescriptor {
	if fd.Kind() == protoreflect.MessageKind {
		md := fd.Message()
		for i := 0; i < md.Fields().Len(); i++ {
			fd := md.Fields().Get(i)
			if fd.IsMap() {
				return fd
			}
		}
	}
	return nil
}

func parseMapValueType(fd protoreflect.FieldDescriptor) string {
	valueType := helper.ParseCppType(fd.MapValue())
	if fd.MapValue().Kind() == protoreflect.MessageKind {
		return "const " + valueType + "*"
	}
	return valueType
}
