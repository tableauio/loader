package main

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/options"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const orderedMapSuffix = "_OrderedMap"
const orderedMapValueSuffix = "_OrderedMapValue"

func genOrderedMapTypeDef(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerFullName string) {
	if depth == 1 && !options.NeedGenOrderedMap(md, options.LangCS) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			if depth == 1 {
				g.P(helper.Indent(2), "// OrderedMap types.")
			}
			nextKeys := helper.AddMapKey(fd, keys)
			keyType := nextKeys[len(nextKeys)-1].Type

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genOrderedMapTypeDef(g, fd.MapValue().Message(), depth+1, nextKeys, messagerFullName)
			}

			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix
			orderedMapValue := prefix + orderedMapValueSuffix

			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				currValueType := helper.ParseCsharpType(fd.MapValue())
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				nextOrderedMap := nextPrefix + orderedMapSuffix
				g.P(helper.Indent(2), "public class ", orderedMapValue, "(", nextOrderedMap, " item1, ", currValueType, "? item2) : Tuple<", nextOrderedMap, ", ", currValueType, "?>(item1, item2) { }")
				g.P(helper.Indent(2), "public class ", orderedMap, " : SortedDictionary<", keyType, ", ", orderedMapValue, "> { }")
				g.P()
			} else {
				g.P(helper.Indent(2), "public class ", orderedMap, " : SortedDictionary<", keyType, ", ", parseMapValueType(fd), "> { }")
				g.P()
			}
			if depth == 1 {
				g.P(helper.Indent(2), "private readonly ", orderedMap, " OrderedMap = [];")
				g.P()
			}
			break
		}
	}
}

func genOrderedMapLoader(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, messagerFullName string) {
	if depth == 1 {
		g.P(helper.Indent(3), "// OrderedMap init.")
		g.P(helper.Indent(3), "OrderedMap.Clear();")
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
			g.P(helper.Indent(depth+2), "foreach (var (", keyName, ", ", valueName, ") in ", prevContainer, ".", strcase.ToCamel(string(fd.Name())), ")")
			g.P(helper.Indent(depth+2), "{")
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				nextPrefix := parseOrderedMapPrefix(nextMapFD, messagerFullName)
				nextOrderedMap := nextPrefix + orderedMapSuffix
				g.P(helper.Indent(depth+3), "var ", tmpOrderedMapName, " = new ", nextOrderedMap, "();")
			}
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genOrderedMapLoader(g, fd.MapValue().Message(), depth+1, messagerFullName)
			}

			if nextMapFD != nil {
				g.P(helper.Indent(depth+3), prevTmpOrderedMapName, "[", keyName, "] = new ", orderedMapValue, "(", tmpOrderedMapName, ", ", valueName, ");")
			} else {
				g.P(helper.Indent(depth+3), prevTmpOrderedMapName, "[", keyName, "] = ", valueName, ";")
			}
			g.P(helper.Indent(depth+2), "}")
			break
		}
	}
}

func genOrderedMapGetters(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerFullName string) {
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
				g.P(helper.Indent(2), "// OrderedMap accessors.")
			}
			getter := genGetterName(depth)
			prefix := parseOrderedMapPrefix(fd, messagerFullName)
			orderedMap := prefix + orderedMapSuffix

			if depth == 1 {
				g.P(helper.Indent(2), "public ref readonly ", orderedMap, " ", getter, "() => ref OrderedMap;")
			} else {
				lastKeyName := keys[len(keys)-1].Name
				if depth == 2 {
					g.P(helper.Indent(2), "public ", orderedMap, "? ", getter, "(", helper.GenGetParams(keys), ") => ",
						"OrderedMap.TryGetValue(", lastKeyName, ", out var value) ? value.Item1 : null;")
				} else {
					prevKeys := keys[:len(keys)-1]
					prevGetter := genGetterName(depth - 1)
					g.P(helper.Indent(2), "public ", orderedMap, "? ", getter, "(", helper.GenGetParams(keys), ") => ",
						prevGetter, "(", helper.GenGetArguments(prevKeys), ")?.TryGetValue(", lastKeyName, ", out var value) == true ? value.Item1 : null;")
				}
			}

			keys = helper.AddMapKey(fd, keys)
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genOrderedMapGetters(g, fd.MapValue().Message(), depth+1, keys, messagerFullName)
			}
			break
		}
	}
}

var caser = cases.Title(language.Und, cases.NoLower)

func parseOrderedMapPrefix(mapFd protoreflect.FieldDescriptor, messagerFullName string) string {
	if mapFd.MapValue().Kind() == protoreflect.MessageKind {
		localMsgProtoName := strings.TrimPrefix(string(mapFd.MapValue().Message().FullName()), messagerFullName+".")
		return caser.String(strings.ReplaceAll(localMsgProtoName, ".", "_"))
	}
	return caser.String(mapFd.MapValue().Kind().String())
}
