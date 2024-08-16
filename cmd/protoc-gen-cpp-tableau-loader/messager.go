package main

import (
	"fmt"
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
	g.P("  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) override;")
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

const mapSuffix = "_Map"
const orderedMapSuffix = "_OrderedMap"
const orderedMapValueSuffix = "_OrderedMapValue"

func genHppOrderedMapGetters(depth int, keys []helper.MapKey, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, messagerFullName string) {
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
			nextKeys := helper.AddMapKey(fd, keys)
			keyType := nextKeys[len(nextKeys)-1].Type

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genHppOrderedMapGetters(depth+1, nextKeys, g, fd.MapValue().Message(), messagerFullName)
			}

			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix
			orderedMapValue := prefix + orderedMapValueSuffix

			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				_, currValueType := parseMapType(fd)
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				nextOrderedMap := nextPrefix + orderedMapSuffix
				// nextOrderedMapValue := nextPrefix + orderedMapValueSuffix
				g.P("  using ", orderedMapValue, " = std::pair<", nextOrderedMap, ", const ", currValueType, "*>;")
				g.P("  using ", orderedMap, " = std::map<", keyType, ", ", orderedMapValue, ">;")
				g.P("  const ", orderedMap, "* GetOrderedMap(", helper.GenGetParams(keys), ") const;")
				g.P()
			} else {
				orderedMapValue := helper.ParseCppType(fd.MapValue())
				constStr := ""
				if fd.MapValue().Kind() == protoreflect.MessageKind {
					orderedMapValue += "*" // If value type is message, should use pointer.
					constStr = "const "
				}
				g.P("  using ", orderedMap, " = std::map<", keyType, ", ", constStr, orderedMapValue, ">;")
				g.P("  const ", orderedMap, "* GetOrderedMap(", helper.GenGetParams(keys), ") const;")
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

func genHppIndexFinders(depth int, params []string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, messagerFullName string) {
	g.P("  // Index accessers.")
	descriptors := index.ParseIndexDescriptor(md)
	for _, descriptor := range descriptors {
		if len(descriptor.Fields) == 1 {
			// single-column index
			field := descriptor.Fields[0] // just take first field
			g.P("  // Index: ", descriptor.Index)
			g.P(" public:")
			vectorType := fmt.Sprintf("Index_%sVector", descriptor.Name)
			mapType := fmt.Sprintf("Index_%sMap", descriptor.Name)
			g.P("  using ", vectorType, " = std::vector<const ", descriptor.FullClassName, "*>;")
			keyType := field.TypeStr
			if field.Type == index.TypeEnum {
				// treat enum as integer
				keyType = "int"
			}
			g.P("  using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, ">;")
			g.P("  const ", mapType, "& Find", descriptor.Name, "() const;")
			g.P("  const ", vectorType, "* Find", descriptor.Name, "(", helper.ToConstRefType(field.TypeStr), " ", field.ScalarName, ") const;")
			g.P("  const ", descriptor.FullClassName, "* FindFirst", descriptor.Name, "(", helper.ToConstRefType(field.TypeStr), " ", field.ScalarName, ") const;")
			g.P()

			g.P(" private:")
			indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"
			g.P("  ", mapType, " ", indexContainerName, ";")
			g.P()
		} else {
			// multi-column index
			g.P("  // Index: ", descriptor.Index)
			g.P(" public:")
			keyType := fmt.Sprintf("Index_%sKey", descriptor.Name)
			keyHasherType := fmt.Sprintf("Index_%sKeyHasher", descriptor.Name)
			vectorType := fmt.Sprintf("Index_%sVector", descriptor.Name)
			mapType := fmt.Sprintf("Index_%sMap", descriptor.Name)

			// generate key struct
			g.P("  struct ", keyType, " {")
			equality := ""
			for i, field := range descriptor.Fields {
				g.P("    ", field.TypeStr, " ", field.ScalarName, ";")
				equality += field.ScalarName + " == other." + field.ScalarName
				if i != len(descriptor.Fields)-1 {
					equality += " && "
				}
			}
			g.P("    bool operator==(const ", keyType, "& other) const {")
			g.P("      return ", equality, ";")
			g.P("    }")
			g.P("  };")

			// generate key hasher struct
			g.P("  struct ", keyHasherType, " {")
			combinedKeys := ""
			for i, field := range descriptor.Fields {
				key := "key." + field.ScalarName
				if field.Type == index.TypeEnum {
					key = "static_cast<int>(" + key + ")"
				}
				combinedKeys += key
				if i != len(descriptor.Fields)-1 {
					combinedKeys += ", "
				}
			}
			g.P("    std::size_t operator()(const ", keyType, "& key) const {")
			g.P("      return util::SugaredHashCombine(", combinedKeys, ");")
			g.P("    }")
			g.P("  };")

			g.P("  using ", vectorType, " = std::vector<const ", descriptor.FullClassName, "*>;")
			g.P("  using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, ", ", keyHasherType, ">;")
			g.P("  const ", mapType, "& Find", descriptor.Name, "() const;")
			g.P("  const ", vectorType, "* Find", descriptor.Name, "(const ", keyType, "& key) const;")
			g.P("  const ", descriptor.FullClassName, "* FindFirst", descriptor.Name, "(const ", keyType, "& key) const;")
			g.P()

			g.P(" private:")
			indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"
			g.P("  ", mapType, " ", indexContainerName, ";")
			g.P()
		}
	}
}

func genCppIndexFinders(messagerName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	descriptors := index.ParseIndexDescriptor(md)
	for _, descriptor := range descriptors {
		vectorType := "Index_" + descriptor.Name + "Vector"
		mapType := "Index_" + descriptor.Name + "Map"
		indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"

		g.P("// Index: ", descriptor.Index)
		g.P("const ", messagerName, "::", mapType, "& "+messagerName+"::Find", descriptor.Name, "() const { return "+indexContainerName+" ;}")
		g.P()

		var keyType, keyName, keyVariable string
		if len(descriptor.Fields) == 1 {
			// single-column index
			field := descriptor.Fields[0] // just take first field
			keyType = field.TypeStr
			keyName = field.ScalarName
			keyVariable = keyName
			if field.Type == index.TypeEnum {
				keyVariable = "static_cast<int>(" + keyVariable + ")"
			}
		} else {
			// multi-column index
			keyType = fmt.Sprintf("const Index_%sKey&", descriptor.Name)
			keyName = "key"
			keyVariable = keyName
		}

		g.P("const ", messagerName, "::", vectorType, "* "+messagerName+"::Find", descriptor.Name, "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
		g.P("  auto iter = ", indexContainerName, ".find(", keyVariable, ");")
		g.P("  if (iter == ", indexContainerName, ".end()) {")
		g.P("    return nullptr;")
		g.P("  }")
		g.P("  return &iter->second;")
		g.P("}")
		g.P()

		g.P("const ", descriptor.FullClassName, "* "+messagerName+"::FindFirst", descriptor.Name, "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
		g.P("  auto conf = Find", descriptor.Name, "(", keyName, ");")
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
	descriptors := index.ParseIndexDescriptor(md)
	for _, descriptor := range descriptors {
		parentDataName := "data_"
		g.P("  // Index: ", descriptor.Index)
		genOneCppIndexLoader(1, descriptor, parentDataName, descriptor.LevelMessage, g)
		g.P()
	}
}

func genOneCppIndexLoader(depth int, descriptor *index.IndexDescriptor, parentDataName string, levelMessage *index.LevelMessage, g *protogen.GeneratedFile) {
	if levelMessage == nil {
		return
	}
	indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"
	if levelMessage.NextLevel != nil {
		itemName := fmt.Sprintf("item%d", depth)
		g.P(strings.Repeat("  ", depth), "for (auto&& "+itemName+" : "+parentDataName+"."+levelMessage.FieldName+"()) {")

		parentDataName = itemName
		if levelMessage.FieldType == index.TypeMap {
			parentDataName = itemName + ".second"
		}
		genOneCppIndexLoader(depth+1, descriptor, parentDataName, levelMessage.NextLevel, g)
		g.P(strings.Repeat("  ", depth), "}")
	} else {
		if len(levelMessage.Fields) == 1 {
			// single-column index
			field := levelMessage.Fields[0] // just take the first field
			if field.Card == index.CardList {
				itemName := fmt.Sprintf("item%d", depth)
				g.P(strings.Repeat("  ", depth), "for (auto&& "+itemName+" : "+parentDataName+"."+field.Name+"()) {")
				key := itemName
				if field.Type == index.TypeEnum {
					// convert enum to integer, which is used as unorderd map key that need hash and comparator
					key = "static_cast<int>(" + itemName + ")"
				}
				g.P(strings.Repeat("  ", depth+1), indexContainerName, "["+key+"].push_back(&"+parentDataName+");")
				g.P(strings.Repeat("  ", depth), "}")
			} else {
				key := parentDataName + "." + field.Name + "()"
				if field.Type == index.TypeEnum {
					// convert enum to integer, which is used as unorderd map key that need hash and comparator
					key = "static_cast<int>(" + key + ")"
				}
				g.P(strings.Repeat("  ", depth), indexContainerName, "["+key+"].push_back(&"+parentDataName+");")
			}
		} else {
			// multi-column index
			var keys []string
			generateOneCppMulticolumnIndex(depth, parentDataName, descriptor, &keys, g)
		}
	}
}

func generateOneCppMulticolumnIndex(depth int, parentDataName string, descriptor *index.IndexDescriptor, keys *[]string, g *protogen.GeneratedFile) {
	cursor := len(*keys)
	if cursor >= len(descriptor.Fields) {
		var keyParams string
		for i, key := range *keys {
			field := descriptor.Fields[i]
			if field.Type == index.TypeEnum {
				// convert enum to integer, which is used as unorderd map key that need hash and comparator
				key = "static_cast<" + field.TypeStr + ">(" + key + ")"
			}
			keyParams += key
			if i != len(*keys)-1 {
				keyParams += ", "
			}
		}
		keyType := fmt.Sprintf("Index_%sKey", descriptor.Name)
		indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"
		g.P(strings.Repeat("  ", depth), keyType, " key{", keyParams, "};")
		g.P(strings.Repeat("  ", depth), indexContainerName, "[key].push_back(&"+parentDataName+");")
		return
	}
	field := descriptor.Fields[cursor]
	if field.Card == index.CardList {
		itemName := fmt.Sprintf("index_item%d", cursor)
		g.P(strings.Repeat("  ", depth), "for (auto&& "+itemName+" : "+parentDataName+"."+field.Name+"()) {")
		*keys = append(*keys, itemName)
		generateOneCppMulticolumnIndex(depth+1, parentDataName, descriptor, keys, g)
		g.P(strings.Repeat("  ", depth), "}")
	} else {
		key := parentDataName + "." + field.Name + "()"
		*keys = append(*keys, key)
		generateOneCppMulticolumnIndex(depth, parentDataName, descriptor, keys, g)
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
	g.P("bool ", message.Desc.Name(), "::Load(const std::string& dir, Format fmt, const LoadOptions* options /* = nullptr */) {")
	g.P("  bool ok = LoadMessage(data_, dir, fmt, options);")
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

func genCppMapGetters(depth int, keys []helper.MapKey, messagerName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
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
				genCppMapGetters(depth+1, keys, messagerName, g, fd.MapValue().Message())
			}
			break
		}
	}
}

func genCppOrderedMapGetters(depth int, keys []helper.MapKey, messagerName, messagerFullName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	if depth == 1 && !helper.NeedGenOrderedMap(md) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix

			g.P("const ", messagerName, "::", orderedMap, "* ", messagerName, "::GetOrderedMap(", helper.GenGetParams(keys), ") const {")
			if depth == 1 {
				g.P("  return &ordered_map_; ")
			} else {
				lastKeyName := keys[len(keys)-1].Name
				prevKeys := keys[:len(keys)-1]
				g.P("  const auto* conf = GetOrderedMap(", helper.GenGetArguments(prevKeys), ");")
				g.P("  if (conf == nullptr) {")
				g.P("    return nullptr;")
				g.P("  }")
				g.P()
				g.P("  auto iter = conf->find(", lastKeyName, ");")
				g.P("  if (iter == conf->end()) {")
				g.P("    return nullptr;")
				g.P("  }")
				g.P("  return &iter->second.first;")

			}
			g.P("}")
			g.P()

			keys = helper.AddMapKey(fd, keys)
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genCppOrderedMapGetters(depth+1, keys, messagerName, messagerFullName, g, fd.MapValue().Message())
			}
			break
		}
	}
}

func genCppOrderedMapLoader(depth int, messagerFullName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
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
			g.P(strings.Repeat("  ", depth), "for (auto&& ", itemName, " : ", prevContainer, ".", string(fd.Name()), "()) {")
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				// nextMap := nextPrefix + mapSuffix
				nextOrderedMap := nextPrefix + orderedMapSuffix
				g.P(strings.Repeat("  ", depth+1), prevTmpOrderedMapName, "[", itemName, ".first] = ", orderedMapValue, "(", nextOrderedMap, "(), &", itemName, ".second);")
				g.P(strings.Repeat("  ", depth+1), "auto&& ", tmpOrderedMapName, " = ", prevTmpOrderedMapName, "[", itemName, ".first].first;")
			} else {
				ref := "&"
				if fd.MapValue().Kind() != protoreflect.MessageKind {
					ref = "" // scalar value type just do value copy.
				}
				g.P(strings.Repeat("  ", depth+1), prevTmpOrderedMapName, "[", itemName, ".first] = ", ref, itemName, ".second;")
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
