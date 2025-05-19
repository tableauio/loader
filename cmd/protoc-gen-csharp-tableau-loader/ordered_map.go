package main

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const orderedMapSuffix = "_OrderedMap"
const orderedMapValueSuffix = "_OrderedMapValue"

func genOrderedMapTypeDef(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerFullName string) {
	if depth == 1 && !options.NeedGenOrderedMap(md, options.LangCS) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			if depth == 1 {
				g.P("        // OrderedMap types.")
			}
			nextKeys := helper.AddMapKey(gen, fd, keys)
			keyType := nextKeys[len(nextKeys)-1].Type

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genOrderedMapTypeDef(gen, g, fd.MapValue().Message(), depth+1, nextKeys, messagerFullName)
			}

			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix
			orderedMapValue := prefix + orderedMapValueSuffix

			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				currValueType := helper.ParseCsharpType(fd.MapValue())
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				nextOrderedMap := nextPrefix + orderedMapSuffix
				g.P("        public class ", orderedMapValue, " : Tuple<", nextOrderedMap, ", ", currValueType, "?>")
				g.P("        {")
				g.P("            public ", orderedMapValue, "(", nextOrderedMap, " item1, ", currValueType, "? item2) : base(item1, item2) { }")
				g.P("        }")
				g.P("        public class ", orderedMap, " : SortedDictionary<", keyType, ", ", orderedMapValue, "> { }")
				g.P()
			} else {
				g.P("        public class ", orderedMap, " : SortedDictionary<", keyType, ", ", parseMapValueType(fd), "> { }")
				g.P()
			}
			if depth == 1 {
				g.P("        private ", orderedMap, " OrderedMap = new ", orderedMap, "();")
				g.P()
			}
			break
		}
	}
}

func genOrderedMapLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, messagerFullName string) {
	if depth == 1 {
		g.P("            // OrderedMap init.")
		g.P("            OrderedMap.Clear();")
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			// orderedMap := prefix + orderedMapSuffix
			orderedMapValue := prefix + orderedMapValueSuffix
			keyName := fmt.Sprintf("key%d", depth)
			valueName := fmt.Sprintf("value%d", depth)

			tmpOrderedMapName := fmt.Sprintf("ordered_map%d", depth)

			prevContainer := fmt.Sprintf("value%d", depth-1)
			prevTmpOrderedMapName := fmt.Sprintf("ordered_map%d", depth-1)
			if depth == 1 {
				prevContainer = "Data_"
				prevTmpOrderedMapName = "OrderedMap"
			}
			g.P(strings.Repeat("    ", depth+2), "foreach (var (", keyName, ", ", valueName, ") in ", prevContainer, ".", strcase.ToCamel(string(fd.Name())), ")")
			g.P(strings.Repeat("    ", depth+2), "{")
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				nextOrderedMap := nextPrefix + orderedMapSuffix
				g.P(strings.Repeat("    ", depth+3), "var ", tmpOrderedMapName, " = new ", nextOrderedMap, "();")
			}
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genOrderedMapLoader(gen, g, fd.MapValue().Message(), depth+1, messagerFullName)
			}

			if nextMapFD != nil {
				g.P(strings.Repeat("    ", depth+3), prevTmpOrderedMapName, "[", keyName, "] = new ", orderedMapValue, "(", tmpOrderedMapName, ", ", valueName, ");")
			} else {
				g.P(strings.Repeat("    ", depth+3), prevTmpOrderedMapName, "[", keyName, "] = ", valueName, ";")
			}
			g.P(strings.Repeat("    ", depth+2), "}")
			break
		}
	}
}

func genOrderedMapGetters(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerFullName string) {
	if depth == 1 && !options.NeedGenOrderedMap(md, options.LangCS) {
		return
	}
	genGetterName := func(depth int) string {
		getter := "GetOrderedMap"
		if depth > 1 {
			getter = fmt.Sprintf("GetOrderedMap%v", depth-1)
		}
		return getter
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			g.P()
			if depth == 1 {
				g.P("        // OrderedMap accessors.")
			}
			getter := genGetterName(depth)
			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix

			if depth == 1 {
				g.P("        public ref readonly ", orderedMap, " ", getter, "()")
				g.P("        {")
				g.P("            return ref OrderedMap;")
			} else {
				g.P("        public ", orderedMap, "? ", getter, "(", helper.GenGetParams(keys), ")")
				g.P("        {")
				lastKeyName := keys[len(keys)-1].Name
				if depth == 2 {
					g.P("            if (OrderedMap.TryGetValue(", lastKeyName, ", out var value))")
				} else {
					prevKeys := keys[:len(keys)-1]
					prevGetter := genGetterName(depth - 1)
					g.P("            var conf = ", prevGetter, "(", helper.GenGetArguments(prevKeys), ");")
					g.P("            if (conf != null && conf.TryGetValue(", lastKeyName, ", out var value))")
				}
				g.P("            {")
				g.P("                return value.Item1;")
				g.P("            }")
				g.P("            return null;")
			}
			g.P("        }")

			keys = helper.AddMapKey(gen, fd, keys)
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genOrderedMapGetters(gen, g, fd.MapValue().Message(), depth+1, keys, messagerFullName)
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
