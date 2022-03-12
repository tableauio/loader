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
	g.P()
	generateHppFileContent(gen, file, g)
	return g
}

// generateCppFile generates loader files related to protoconf files.
func generateCppFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + "." + pcExt + ".cc"
	g := gen.NewGeneratedFile(filename, "")
	generateFileHeader(gen, file, g)
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

	if needGenOrderedMap(message.Desc) {
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
	if needGenOrderedMap(message.Desc) {
		g.P()
		genHppOrderedMapGetters(1, nil, g, message.Desc, string(message.Desc.FullName()))
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
			keyParam := parseCppType(fd.MapKey()) + fmt.Sprintf(" key%d", depth)
			params = append(params, keyParam)
			g.P("  const ", parseCppType(fd.MapValue()), "* Get(", strings.Join(params, ", "), ") const;")
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

func genHppOrderedMapGetters(depth int, params []string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, messagerFullName string) {
	if depth == 1 && !needGenOrderedMap(md) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			if depth == 1 {
				g.P(" public:")
			}
			keyType := parseCppType(fd.MapKey())
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
				orderedMapValue = parseCppType(fd.MapValue())
				g.P("  using ", orderedMap, " = std::map<", keyType, ", ", orderedMapValue, ">;")
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
		return parseCppType(fd.MapKey()), parseCppType(fd.MapValue())
	}
	return "", ""
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

	if needGenOrderedMap(message.Desc) {
		g.P("bool ", message.Desc.Name(), "::ProcessAfterLoad() {")
		genCppOrderedMapLoader(1, string(message.Desc.FullName()), g, message.Desc)
		g.P("  return true;")
		g.P("}")
		g.P()
	}

	// syntactic sugar for accessing map items
	genCppMapGetters(1, nil, string(message.Desc.Name()), g, message.Desc)
	genCppOrderedMapGetters(1, nil, string(message.Desc.Name()), string(message.Desc.FullName()), g, message.Desc)
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
			break
		}
	}
}

func genCppOrderedMapGetters(depth int, params []string, messagerName, messagerFullName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	if depth == 1 && !needGenOrderedMap(md) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix

			g.P("const ", messagerName, "::", orderedMap, "* ", messagerName, "::GetOrderedMap(", strings.Join(params, ", "), ") const {")
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

			keyType := parseCppType(fd.MapKey())
			keyParam := keyType + fmt.Sprintf(" key%d", depth)
			nextParams := append(params, keyParam)
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genCppOrderedMapGetters(depth+1, nextParams, messagerName, messagerFullName, g, fd.MapValue().Message())
			}
			break
		}
	}
}

func genCppOrderedMapLoader(depth int, messagerFullName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
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
				g.P(strings.Repeat("  ", depth+1), prevTmpOrderedMapName, "[", itemName, ".first] = ", orderedMapValue, "(", nextOrderedMap, "(), &", itemName, ".second.", string(nextMapFD.Name()), "());")
				g.P(strings.Repeat("  ", depth+1), "auto&& ", tmpOrderedMapName, " = ", prevTmpOrderedMapName, "[", itemName, ".first].first;")
			} else {
				g.P(strings.Repeat("  ", depth+1), prevTmpOrderedMapName, "[", itemName, ".first] = ", itemName, ".second;")
			}
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genCppOrderedMapLoader(depth+1, messagerFullName, g, fd.MapValue().Message())
			}
			g.P(strings.Repeat("  ", depth), "}")
			break

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
		return "float"
	case protoreflect.DoubleKind:
		return "double"
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
