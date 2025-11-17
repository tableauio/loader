package main

import (
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	idx "github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/index"
	orderedindex "github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/ordered_index"
	orderedmap "github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/ordered_map"
	"github.com/tableauio/loader/internal/extensions"
	"github.com/tableauio/loader/internal/index"
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
	filename := file.GeneratedFilenamePrefix + "." + extensions.PC + ".h"
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, file, g, version)
	g.P()
	generateHppFileContent(file, g)
	return g
}

// generateCppFile generates loader files related to protoconf files.
func generateCppFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + "." + extensions.PC + ".cc"
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, file, g, version)
	g.P()
	generateCppFileContent(file, g)
	return g
}

// generateHppFileContent generates type definitions.
func generateHppFileContent(file *protogen.File, g *protogen.GeneratedFile) {
	g.P("#pragma once")
	g.P("#include <filesystem>")
	g.P("#include <string>")
	g.P()
	g.P(`#include "`, "load.", extensions.PC, `.h"`)
	g.P(`#include "`, "util.", extensions.PC, `.h"`)
	g.P(`#include "`, file.GeneratedFilenamePrefix, ".", extensions.PB, `.h"`)
	g.P()

	g.P("namespace ", *namespace, " {")
	var fileMessagers []string
	for _, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			genHppMessage(g, message)
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
func genHppMessage(g *protogen.GeneratedFile, message *protogen.Message) {
	cppFullName := helper.ParseCppClassType(message.Desc)
	indexDescriptor := index.ParseIndexDescriptor(message.Desc)

	orderedMapGenerator := orderedmap.NewGenerator(g, message)
	indexGenerator := idx.NewGenerator(g, indexDescriptor, message)
	orderedIndexGenerator := orderedindex.NewGenerator(g, indexDescriptor, message)

	g.P("class ", message.Desc.Name(), " : public Messager {")
	g.P(" public:")
	g.P(helper.Indent(1), "static const std::string& Name() { return kProtoName; }")
	g.P(helper.Indent(1), "virtual bool Load(const std::filesystem::path& dir, Format fmt, std::shared_ptr<const load::MessagerOptions> options = nullptr) override;")
	g.P(helper.Indent(1), "const ", cppFullName, "& Data() const { return data_; }")
	g.P(helper.Indent(1), "const google::protobuf::Message* Message() const override { return &data_; }")
	g.P()

	if orderedMapGenerator.NeedGenerate() || indexGenerator.NeedGenerate() || orderedIndexGenerator.NeedGenerate() {
		g.P(" private:")
		g.P(helper.Indent(1), "virtual bool ProcessAfterLoad() override final;")
		g.P()
	}

	// syntactic sugar for accessing map items
	genHppMapGetters(1, nil, g, message.Desc)
	g.P()
	g.P(" private:")
	g.P(helper.Indent(1), "static const std::string kProtoName;")
	g.P(helper.Indent(1), cppFullName, " data_;")
	orderedMapGenerator.GenHppOrderedMapGetters()
	indexGenerator.GenHppIndexFinders()
	orderedIndexGenerator.GenHppOrderedIndexFinders()
	g.P("};")
	g.P()
}

func genHppMapGetters(depth int, keys helper.MapKeys, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			if depth == 1 {
				g.P(" public:")
			}
			keys = keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			g.P(helper.Indent(1), "const ", helper.ParseCppType(fd.MapValue()), "* Get(", keys.GenGetParams(), ") const;")
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genHppMapGetters(depth+1, keys, g, fd.MapValue().Message())
			}
			break
		}
	}
}

// generateCppFileContent generates type implementations.
func generateCppFileContent(file *protogen.File, g *protogen.GeneratedFile) {
	g.P(`#include "`, file.GeneratedFilenamePrefix, ".", extensions.PC, `.h"`)
	g.P()
	g.P(`#include "hub.pc.h"`)
	g.P(`#include "util.pc.h"`)
	g.P()

	g.P("namespace ", *namespace, " {")
	for _, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			genCppMessage(g, message)
		}
	}
	g.P("}  // namespace ", *namespace)
}

// genCppMessage generates a message implementation.
func genCppMessage(g *protogen.GeneratedFile, message *protogen.Message) {
	messagerName := string(message.Desc.Name())
	cppFullName := helper.ParseCppClassType(message.Desc)
	indexDescriptor := index.ParseIndexDescriptor(message.Desc)

	orderedMapGenerator := orderedmap.NewGenerator(g, message)
	indexGenerator := idx.NewGenerator(g, indexDescriptor, message)
	orderedIndexGenerator := orderedindex.NewGenerator(g, indexDescriptor, message)

	g.P("const std::string ", messagerName, "::kProtoName = ", cppFullName, `::GetDescriptor()->name();`)
	g.P()
	g.P("bool ", messagerName, "::Load(const std::filesystem::path& dir, Format fmt, std::shared_ptr<const load::MessagerOptions> options /* = nullptr */) {")
	g.P(helper.Indent(1), "tableau::util::TimeProfiler profiler;")
	g.P(helper.Indent(1), "bool loaded = LoadMessagerInDir(data_, dir, fmt, options);")
	g.P(helper.Indent(1), "bool ok = loaded ? ProcessAfterLoad() : false;")
	g.P(helper.Indent(1), "stats_.duration = profiler.Elapse();")
	g.P(helper.Indent(1), "return ok;")
	g.P("}")
	g.P()

	if orderedMapGenerator.NeedGenerate() || indexGenerator.NeedGenerate() || orderedIndexGenerator.NeedGenerate() {
		g.P("bool ", messagerName, "::ProcessAfterLoad() {")
		orderedMapGenerator.GenOrderedMapLoader()
		indexGenerator.GenCppIndexLoader()
		orderedIndexGenerator.GenCppOrderedIndexLoader()
		g.P(helper.Indent(1), "return true;")
		g.P("}")
		g.P()
	}

	// syntactic sugar for accessing map items
	genCppMapGetters(g, message.Desc, 1, nil, messagerName)
	orderedMapGenerator.GenOrderedMapGetters()
	indexGenerator.GenCppIndexFinders()
	orderedIndexGenerator.GenCppOrderedIndexFinders()
}

func genCppMapGetters(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys helper.MapKeys, messagerName string) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			keys = keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			g.P("const ", helper.ParseCppType(fd.MapValue()), "* ", messagerName, "::Get(", keys.GenGetParams(), ") const {")

			var container string
			if depth == 1 {
				container = "data_." + string(fd.Name()) + "()"
			} else {
				container = "conf->" + string(fd.Name()) + "()"
				prevKeys := keys[:len(keys)-1]
				g.P(helper.Indent(1), "const auto* conf = Get(", prevKeys.GenGetArguments(), ");")
				g.P(helper.Indent(1), "if (conf == nullptr) {")
				g.P(helper.Indent(2), "return nullptr;")
				g.P(helper.Indent(1), "}")
			}
			lastKeyName := keys[len(keys)-1].Name
			g.P(helper.Indent(1), "auto iter = ", container, ".find(", lastKeyName, ");")
			g.P(helper.Indent(1), "if (iter == ", container, ".end()) {")
			g.P(helper.Indent(2), "return nullptr;")
			g.P(helper.Indent(1), "}")
			g.P(helper.Indent(1), "return &iter->second;")
			g.P("}")
			g.P()

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genCppMapGetters(g, fd.MapValue().Message(), depth+1, keys, messagerName)
			}
			break
		}
	}
}
