package main

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/index"
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
	g.P(`#include "`, "hub.", pcExt, `.h"`)
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
	messagers = append(messagers, fileMessagers...)
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

	g.P("class ", message.Desc.Name(), " : public Messager {")
	g.P(" public:")
	g.P("  static const std::string& Name() { return kProtoName; };")
	g.P("  virtual bool Load(const std::string& dir, Format fmt) override;")
	g.P("  const ", cppFullName, "& Data() const { return data_; };")
	g.P()

	if helper.NeedGenOrderedMap(message.Desc) || index.NeedGenIndex(message.Desc) {
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
	if helper.NeedGenOrderedMap(message.Desc) {
		g.P()
		genHppOrderedMapGetters(1, nil, g, message.Desc, string(message.Desc.FullName()))
	}
	if index.NeedGenIndex(message.Desc) {
		g.P()
		genHppIndexFinders(1, nil, g, message.Desc, string(message.Desc.FullName()))
	}
	g.P("};")
	g.P()
}

func genHppMapGetters(depth int, params []string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			if depth == 1 {
				g.P(" public:")
			}
			keyParam := helper.ParseCppType(fd.MapKey()) + fmt.Sprintf(" key%d", depth)
			params = append(params, keyParam)
			g.P("  const ", helper.ParseCppType(fd.MapValue()), "* Get(", strings.Join(params, ", "), ") const;")
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genHppMapGetters(depth+1, params, g, fd.MapValue().Message())
			}
			break
		}
	}
}

const mapSuffix = "_Map"
const orderedMapSuffix = "_OrderedMap"
const orderedMapValueSuffix = "_OrderedMapValue"

func genHppOrderedMapGetters(depth int, params []string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor,
	messagerFullName string) {
	if depth == 1 && !helper.NeedGenOrderedMap(md) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			if depth == 1 {
				g.P("  // OrderedMap accessers.")
				g.P(" public:")
			}
			keyType := helper.ParseCppType(fd.MapKey())
			keyParam := keyType + fmt.Sprintf(" key%d", depth)
			nextParams := append(params, keyParam)

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genHppOrderedMapGetters(depth+1, nextParams, g, fd.MapValue().Message(), messagerFullName)
			}

			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix
			orderedMapValue := prefix + orderedMapValueSuffix

			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				nextKeyType, nextValueType := parseMapType(nextMapFD)
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				nextMap := nextPrefix + mapSuffix
				nextOrderedMap := nextPrefix + orderedMapSuffix
				// nextOrderedMapValue := nextPrefix + orderedMapValueSuffix
				g.P("  using ", nextMap, " = ::google::protobuf::Map<", nextKeyType, ", ", nextValueType, ">;")
				g.P("  using ", orderedMapValue, " = std::pair<", nextOrderedMap, ", const ", nextMap, "*>;")
				g.P("  using ", orderedMap, " = std::map<", keyType, ", ", orderedMapValue, ">;")
				g.P("  const ", orderedMap, "* GetOrderedMap(", strings.Join(params, ", "), ") const;")
				g.P()
			} else {
				orderedMapValue := helper.ParseCppType(fd.MapValue())
				constStr := ""
				if fd.MapValue().Kind() == protoreflect.MessageKind {
					orderedMapValue += "*" // If value type is message, should use pointer.
					constStr = "const "
				}
				g.P("  using ", orderedMap, " = std::map<", keyType, ", ", constStr, orderedMapValue, ">;")
				g.P("  const ", orderedMap, "* GetOrderedMap(", strings.Join(params, ", "), ") const;")
				g.P()
			}
			if depth == 1 {
				g.P(" private:")
				g.P("  ", orderedMap, " ordered_map_;")
			}
			break
		}
	}
}

func parseOrderedMapPrefix(mapFd protoreflect.FieldDescriptor, messagerFullName string) string {
	if mapFd.MapValue().Kind() == protoreflect.MessageKind {
		localMsgProtoName := strings.TrimPrefix(string(mapFd.MapValue().Message().FullName()), messagerFullName+".")
		return strings.ReplaceAll(localMsgProtoName, ".", "_")
	}
	return fmt.Sprintf("%s", mapFd.MapValue().Kind())
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

func parseMapType(fd protoreflect.FieldDescriptor) (keyType, valueType string) {
	if fd.IsMap() {
		return helper.ParseCppType(fd.MapKey()), helper.ParseCppType(fd.MapValue())
	}
	return "", ""
}

func genHppIndexFinders(depth int, params []string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor,
	messagerFullName string) {
	g.P("  // Index accessers.")
	indexInfos := index.ParseIndexInfo(md)
	for _, info := range indexInfos {
		g.P(" public:")
		vectorName := fmt.Sprintf("Index_%sVector", info.IndexName)
		mapName := fmt.Sprintf("Index_%sMap", info.IndexName)
		g.P("  using ", vectorName, " = std::vector<const ", info.FullClassName, "*>;")
		g.P("  using ", mapName, " = std::unordered_map<", info.IndexFieldType, ", ", vectorName, ">;")
		g.P("  const ", mapName, "& Find", info.IndexName, "() const;")
		g.P("  const ", vectorName, "* Find", info.IndexName, "(", info.IndexFieldType, " ", info.IndexFieldName,
			") const;")
		g.P("  const ", info.FullClassName, "* FindFirst", info.IndexName, "(", info.IndexFieldType, " ",
			info.IndexFieldName, ") const;")
		g.P()

		g.P(" private:")
		indexContainerName := "index_" + strcase.ToSnake(info.IndexName) + "_map_"
		g.P("  ", mapName, " ", indexContainerName, ";")
		g.P()
	}
}

func genCppIndexFinders(messagerName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	indexInfos := index.ParseIndexInfo(md)
	for _, info := range indexInfos {
		vectorName := "Index_" + info.IndexName + "Vector"
		mapName := "Index_" + info.IndexName + "Map"
		indexContainerName := "index_" + strcase.ToSnake(info.IndexName) + "_map_"

		g.P("const ", messagerName, "::", mapName, "& "+messagerName+"::Find", info.IndexName,
			"() const { return "+indexContainerName+" ;}")
		g.P()

		g.P("const ", messagerName, "::", vectorName, "* "+messagerName+"::Find", info.IndexName, "(",
			info.IndexFieldType, " ", info.IndexFieldName, ") const {")
		g.P("  auto iter = ", indexContainerName, ".find(", info.IndexFieldName, ");")
		g.P("  if (iter == ", indexContainerName, ".end()) {")
		g.P("    return nullptr;")
		g.P("  }")
		g.P("  return &iter->second;")
		g.P("}")
		g.P()

		g.P("const ", info.FullClassName, "* "+messagerName+"::FindFirst", info.IndexName, "(", info.IndexFieldType,
			" ", info.IndexFieldName, ") const {")
		g.P("  auto conf = Find", info.IndexName, "(", info.IndexFieldName, ");")
		g.P("  if (conf == nullptr || conf->size() == 0) {")
		g.P("    return nullptr;")
		g.P("  }")
		g.P("  return (*conf)[0];")
		g.P("}")
		g.P()
	}
}

func genCppIndexLoader(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	g.P("  // Index init.")
	indexInfos := index.ParseIndexInfo(md)
	for _, info := range indexInfos {
		indexContainerName := "index_" + strcase.ToSnake(info.IndexName) + "_map_"
		parentDataName := "data_"
		genOneCppIndexLoader(1, indexContainerName, parentDataName, info.LevelInfo, g)
		g.P()
	}
}

func genOneCppIndexLoader(depth int, indexContainerName string, parentDataName string, info *index.LevelInfo,
	g *protogen.GeneratedFile) {

	// for (auto&& item1 : data_.activity_map()) {
	// 	for (auto&& item2 : item1.second.chapter_map()) {
	// 	  index_chapter_map_[item2.second.chapter_id()].push_back(&item2.second);
	// 	}
	// }
	if info == nil {
		return
	}

	if info.NextLevel != nil {
		itemName := fmt.Sprintf("item%d", depth)
		g.P(strings.Repeat("  ", depth), "for (auto&& "+itemName+" : "+parentDataName+"."+info.FieldName+"()) {")

		parentDataName = itemName
		if info.Type == index.TypeMap {
			parentDataName = itemName + ".second"
		}
		genOneCppIndexLoader(depth+1, indexContainerName, parentDataName, info.NextLevel, g)
		g.P(strings.Repeat("  ", depth), "}")
	} else {
		g.P(strings.Repeat("  ", depth), indexContainerName,
			"["+parentDataName+"."+info.FieldName+"()].push_back(&"+parentDataName+");")
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
	g.P()
	g.P("bool ", message.Desc.Name(), "::Load(const std::string& dir, Format fmt) {")
	g.P("  bool ok = LoadMessage(dir, data_, fmt);")
	g.P("  return ok ? ProcessAfterLoad() : false;")
	g.P("}")
	g.P()

	if helper.NeedGenOrderedMap(message.Desc) || index.NeedGenIndex(message.Desc) {
		g.P("bool ", message.Desc.Name(), "::ProcessAfterLoad() {")
		if helper.NeedGenOrderedMap(message.Desc) {
			genCppOrderedMapLoader(1, string(message.Desc.FullName()), g, message.Desc)
		}
		if index.NeedGenIndex(message.Desc) {
			genCppIndexLoader(g, message.Desc)
		}
		g.P("  return true;")
		g.P("}")
		g.P()
	}

	// syntactic sugar for accessing map items
	genCppMapGetters(1, nil, string(message.Desc.Name()), g, message.Desc)
	genCppOrderedMapGetters(1, nil, string(message.Desc.Name()), string(message.Desc.FullName()), g, message.Desc)
	if index.NeedGenIndex(message.Desc) {
		genCppIndexFinders(string(message.Desc.Name()), g, message.Desc)
		g.P()
	}
}

func genCppMapGetters(depth int, params []string, messagerName string, g *protogen.GeneratedFile,
	md protoreflect.MessageDescriptor) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			keyType := helper.ParseCppType(fd.MapKey())
			keyName := fmt.Sprintf("key%d", depth)

			params = append(params, keyType+" "+keyName)

			g.P("const ", helper.ParseCppType(fd.MapValue()), "* ", messagerName, "::Get(", strings.Join(params, ", "),
				") const {")

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
			break
		}
	}
}

func genCppOrderedMapGetters(depth int, params []string, messagerName, messagerFullName string,
	g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	if depth == 1 && !helper.NeedGenOrderedMap(md) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix

			g.P("const ", messagerName, "::", orderedMap, "* ", messagerName, "::GetOrderedMap(",
				strings.Join(params, ", "), ") const {")
			if depth == 1 {
				g.P("  return &ordered_map_; ")
			} else {
				keyName := fmt.Sprintf("key%d", depth-1)
				var findParams []string
				for i := 1; i < depth-1; i++ {
					findParams = append(findParams, fmt.Sprintf("key%d", i))
				}
				g.P("  const auto* conf = GetOrderedMap(", strings.Join(findParams, ", "), ");")
				g.P("  if (conf == nullptr) {")
				g.P("    return nullptr;")
				g.P("  }")
				g.P()
				g.P("  auto iter = conf->find(", keyName, ");")
				g.P("  if (iter == conf->end()) {")
				g.P("    return nullptr;")
				g.P("  }")
				g.P("  return &iter->second.first;")

			}
			g.P("}")
			g.P()

			keyType := helper.ParseCppType(fd.MapKey())
			keyParam := keyType + fmt.Sprintf(" key%d", depth)
			nextParams := append(params, keyParam)
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genCppOrderedMapGetters(depth+1, nextParams, messagerName, messagerFullName, g, fd.MapValue().Message())
			}
			break
		}
	}
}

func genCppOrderedMapLoader(depth int, messagerFullName string, g *protogen.GeneratedFile,
	md protoreflect.MessageDescriptor) {
	if depth == 1 {
		g.P("  // OrderedMap init.")
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			// orderedMap := prefix + orderedMapSuffix
			orderedMapValue := prefix + orderedMapValueSuffix
			itemName := fmt.Sprintf("item%d", depth)

			tmpOrderedMapName := fmt.Sprintf("ordered_map%d", depth)

			prevItemName := fmt.Sprintf("item%d", depth-1)
			prevContainer := prevItemName + ".second"
			prevTmpOrderedMapName := fmt.Sprintf("ordered_map%d", depth-1)
			if depth == 1 {
				prevContainer = "data_"
				prevTmpOrderedMapName = "ordered_map_"
			}
			g.P(strings.Repeat("  ", depth), "for (auto&& ", itemName, " : ", prevContainer, ".", string(fd.Name()),
				"()) {")
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				// nextMap := nextPrefix + mapSuffix
				nextOrderedMap := nextPrefix + orderedMapSuffix
				g.P(strings.Repeat("  ", depth+1), prevTmpOrderedMapName, "[", itemName, ".first] = ", orderedMapValue,
					"(", nextOrderedMap, "(), &", itemName, ".second.", string(nextMapFD.Name()), "());")
				g.P(strings.Repeat("  ", depth+1), "auto&& ", tmpOrderedMapName, " = ", prevTmpOrderedMapName, "[",
					itemName, ".first].first;")
			} else {
				ref := "&"
				if fd.MapValue().Kind() != protoreflect.MessageKind {
					ref = "" // scalar value type just do value copy.
				}
				g.P(strings.Repeat("  ", depth+1), prevTmpOrderedMapName, "[", itemName, ".first] = ", ref, itemName,
					".second;")
			}
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genCppOrderedMapLoader(depth+1, messagerFullName, g, fd.MapValue().Message())
			}
			g.P(strings.Repeat("  ", depth), "}")
			break
		}
	}
	if depth == 1 {
		g.P("")
	}
}

func appendMessager(gen *protogen.Plugin, file *protogen.File) {
	var fileMessagers []string
	for _, message := range file.Messages {
		opts, ok := message.Desc.Options().(*descriptorpb.MessageOptions)
		if !ok {
			gen.Error(errors.New("get message options failed"))
		}
		worksheet, ok := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if !ok {
			gen.Error(errors.New("get worksheet extension failed"))
		}
		if worksheet != nil {
			messagerName := string(message.Desc.Name())
			fileMessagers = append(fileMessagers, messagerName)
		}
	}
	messagers = append(messagers, fileMessagers...)
}
