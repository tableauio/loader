package main

import (
	"fmt"
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const orderedMapSuffix = "_OrderedMap"
const orderedMapValueSuffix = "_OrderedMapValue"

func genHppOrderedMapGetters(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerFullName string) {
	if depth == 1 && !options.NeedGenOrderedMap(md, options.LangCPP) {
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
				genHppOrderedMapGetters(g, fd.MapValue().Message(), depth+1, nextKeys, messagerFullName)
			}

			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix
			orderedMapValue := prefix + orderedMapValueSuffix

			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				currValueType := helper.ParseCppType(fd.MapValue())
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				nextOrderedMap := nextPrefix + orderedMapSuffix
				// nextOrderedMapValue := nextPrefix + orderedMapValueSuffix
				g.P("  using ", orderedMapValue, " = std::pair<", nextOrderedMap, ", const ", currValueType, "*>;")
				g.P("  using ", orderedMap, " = std::map<", keyType, ", ", orderedMapValue, ">;")
				g.P("  const ", orderedMap, "* GetOrderedMap(", helper.GenGetParams(keys), ") const;")
				g.P()
			} else {
				g.P("  using ", orderedMap, " = std::map<", keyType, ", ", parseMapValueType(fd), ">;")
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

func genCppOrderedMapLoader(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, messagerFullName string) {
	if depth == 1 {
		g.P("  // OrderedMap init.")
		g.P("  ordered_map_.clear();")
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
				genCppOrderedMapLoader(g, fd.MapValue().Message(), depth+1, messagerFullName)
			}
			g.P(strings.Repeat("  ", depth), "}")
			break
		}
	}
}

func genCppOrderedMapGetters(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerName, messagerFullName string) {
	if depth == 1 && !options.NeedGenOrderedMap(md, options.LangCPP) {
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
				genCppOrderedMapGetters(g, fd.MapValue().Message(), depth+1, keys, messagerName, messagerFullName)
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
	return mapFd.MapValue().Kind().String()
}
