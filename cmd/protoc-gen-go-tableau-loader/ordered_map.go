package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const orderedMapSuffix = "_OrderedMap"
const orderedMapValueSuffix = "_OrderedMapValue"

var orderedMapTypeDefMap map[string]bool = make(map[string]bool)

func genOrderedMapTypeDef(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerName string) {
	if depth == 1 && !options.NeedGenOrderedMap(md, options.LangGO) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			if depth == 1 {
				g.P("// OrderedMap types.")
			}
			nextKeys := helper.AddMapKey(gen, fd, keys)
			keyType := nextKeys[len(nextKeys)-1].Type
			if keyType == "bool" {
				keyType = "int"
			}
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genOrderedMapTypeDef(gen, g, fd.MapValue().Message(), depth+1, nextKeys, messagerName)
			}
			prefix := parseOrderedMapPrefix(fd)
			orderedMap := prefix + orderedMapSuffix
			orderedMapValue := prefix + orderedMapValueSuffix
			_, ok := orderedMapTypeDefMap[orderedMap]
			if !ok {
				orderedMapTypeDefMap[orderedMap] = true
				nextMapFD := getNextLevelMapFD(fd.MapValue())
				if nextMapFD != nil {
					currValueType := helper.FindMessageGoIdent(gen, fd.MapValue().Message())
					nextPrefix := parseOrderedMapPrefix(nextMapFD)
					nextOrderedMap := nextPrefix + orderedMapSuffix
					g.P("type ", orderedMapValue, "= ", pairPackage.Ident("Pair"), "[*", nextOrderedMap, ", *", currValueType, "];")
					g.P("type ", orderedMap, "= ", treeMapPackage.Ident("TreeMap"), "[", keyType, ", *", orderedMapValue, "]")
					g.P()
				} else {
					g.P("type ", orderedMap, "= ", treeMapPackage.Ident("TreeMap"), "[", keyType, ", ", parseMapValueType(gen, g, fd), "]")
					g.P()
				}
			}
			return
		}
	}
}

func genOrderedMapField(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	if !options.NeedGenOrderedMap(md, options.LangGO) {
		return
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			g.P("orderedMap *", parseOrderedMapPrefix(fd), orderedMapSuffix)
			return
		}
	}
}

func genOrderedMapLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerName string, lastOrderedMapValue string) {
	if depth == 1 {
		g.P("// OrderedMap init.")
	}
	message := helper.FindMessage(gen, md)
	for _, field := range message.Fields {
		fd := field.Desc
		if fd.IsMap() {
			needConvertBool := false
			if len(keys) > 0 && keys[len(keys)-1].Type == "bool" {
				needConvertBool = true
			}
			nextKeys := helper.AddMapKey(gen, fd, keys)
			keyType := nextKeys[len(nextKeys)-1].Type
			needConvertBoolNext := false
			if keyType == "bool" {
				keyType = "int"
				needConvertBoolNext = true
			}
			prefix := parseOrderedMapPrefix(fd)
			orderedMapValue := prefix + orderedMapValueSuffix
			mapName := fmt.Sprintf("x.Data().Get%s()", field.GoName)
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if depth == 1 {
				if nextMapFD == nil {
					g.P("x.orderedMap = ", treeMapPackage.Ident("New"), "[", keyType, ", ", parseMapValueType(gen, g, fd), "]()")
				} else {
					g.P("x.orderedMap = ", treeMapPackage.Ident("New"), "[", keyType, ", *", orderedMapValue, "]()")
				}
			}
			if depth != 1 {
				mapName = fmt.Sprintf("v%d.Get%s()", depth-1, field.GoName)
				keyName := fmt.Sprintf("k%d", depth-1)
				if needConvertBool {
					keyName = fmt.Sprintf("BoolToInt(%s)", keyName)
				}
				g.P("k", depth-1, "v := &", lastOrderedMapValue, "{")
				if nextMapFD == nil {
					g.P("First: ", treeMapPackage.Ident("New"), "[", keyType, ", ", parseMapValueType(gen, g, fd), "](),")
				} else {
					g.P("First: ", treeMapPackage.Ident("New"), "[", keyType, ", *", orderedMapValue, "](),")
				}
				g.P("Second: v", depth-1, ",")
				g.P("}")
				g.P("map", depth-1, ".Put(", keyName, ", k", depth-1, "v)")
			}
			g.P("for k", depth, ", v", depth, " := range ", mapName, "{")
			if depth == 1 {
				g.P("map", depth, " := x.orderedMap")
			} else {
				g.P("map", depth, " := k", depth-1, "v.First")
			}
			if nextMapFD != nil {
				genOrderedMapLoader(gen, g, fd.MapValue().Message(), depth+1, nextKeys, messagerName, orderedMapValue)
			} else {
				keyName := fmt.Sprintf("k%d", depth)
				if needConvertBoolNext {
					keyName = fmt.Sprintf("BoolToInt(%s)", keyName)
				}
				g.P("map", depth, ".Put(", keyName, ", v", depth, ")")
			}
			g.P("}")
			break
		}
	}
}

func genOrderedMapGetters(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerName string) {
	if depth == 1 && !options.NeedGenOrderedMap(md, options.LangGO) {
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
			getter := genGetterName(depth)
			prefix := parseOrderedMapPrefix(fd)
			orderedMap := prefix + orderedMapSuffix
			if depth == 1 {
				g.P("// ", getter, " returns the 1-level ordered map.")
				g.P("func (x *", messagerName, ") ", getter, "(", helper.GenGetParams(keys), ") *", orderedMap, "{")
				g.P("return x.orderedMap ")
			} else {
				g.P("// ", getter, " finds value in the ", depth-1, "-level ordered map. It will return")
				g.P("// NotFound error if the key is not found.")
				g.P("func (x *", messagerName, ") ", getter, "(", helper.GenGetParams(keys), ") (*", orderedMap, ", error) {")
				if depth == 2 {
					g.P("conf := x.orderedMap")
				} else {
					prevKeys := keys[:len(keys)-1]
					prevGetter := genGetterName(depth - 1)
					g.P("conf, err := x.", prevGetter, "(", helper.GenGetArguments(prevKeys), ")")
					g.P("if err != nil {")
					g.P(`return nil, err`)
					g.P("}")
				}
				lastKeyName := keys[len(keys)-1].Name
				lastKeyType := keys[len(keys)-1].Type
				keyName := lastKeyName
				if lastKeyType == "bool" {
					keyName = fmt.Sprintf("BoolToInt(%s)", keyName)
				}
				g.P("if val, ok := conf.Get(", keyName, "); !ok {")
				g.P(`return nil, `, fmtPackage.Ident("Errorf"), `("`, lastKeyName, `(%v) %w", `, lastKeyName, `, ErrNotFound)`)
				g.P("} else {")
				g.P(`return val.First, nil`)
				g.P("}")

			}
			g.P("}")
			g.P()

			nextKeys := helper.AddMapKey(gen, fd, keys)
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genOrderedMapGetters(gen, g, fd.MapValue().Message(), depth+1, nextKeys, messagerName)
			}
			break
		}
	}
}

func parseOrderedMapPrefix(mapFd protoreflect.FieldDescriptor) string {
	return strcase.ToCamel(string(mapFd.FullName()))
}
