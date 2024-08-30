package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	helper "github.com/tableauio/loader/internal/helper/go"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const orderedMapSuffix = "_OrderedMap"
const orderedMapValueSuffix = "_OrderedMapValue"

var orderedMapTypeDefMap map[string]bool = make(map[string]bool)

func genOrderedMapTypeDef(gen *protogen.Plugin, depth int, keys []helper.MapKey, messagerName string, g *protogen.GeneratedFile, message *protogen.Message) {
	if *disableOrderedMap {
		return
	}
	if depth == 1 && !helper.NeedGenOrderedMap(message.Desc) {
		return
	}
	for _, field := range message.Fields {
		fd := field.Desc
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
				msg := helper.FindMessage(gen, fd.MapValue().Message())
				if msg != nil {
					genOrderedMapTypeDef(gen, depth+1, nextKeys, messagerName, g, msg)
				}
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
					orderedMapValue := helper.ParseGoType(gen, fd.MapValue())
					if fd.MapValue().Kind() == protoreflect.MessageKind {
						g.P("type ", orderedMap, "= ", treeMapPackage.Ident("TreeMap"), "[", keyType, ", *", helper.FindMessageGoIdent(gen, fd.MapValue().Message()), "]")
					} else {
						g.P("type ", orderedMap, "= ", treeMapPackage.Ident("TreeMap"), "[", keyType, ", ", orderedMapValue, "]")
					}
					g.P()
				}
			}
			return
		}
	}
}

func genOrderedMapField(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	if *disableOrderedMap {
		return
	}
	if !helper.NeedGenOrderedMap(md) {
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

func genOrderedMapLoader(gen *protogen.Plugin, depth int, keys []helper.MapKey, messagerName string, g *protogen.GeneratedFile, message *protogen.Message, lastOrderedMapValue string) {
	if *disableOrderedMap {
		return
	}
	if depth == 1 {
		g.P("// OrderedMap init.")
	}
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
					if fd.MapValue().Kind() == protoreflect.MessageKind {
						g.P("x.orderedMap = ", treeMapPackage.Ident("New"), "[", keyType, ", *", helper.FindMessageGoIdent(gen, fd.MapValue().Message()), "]()")
					} else {
						g.P("x.orderedMap = ", treeMapPackage.Ident("New"), "[", keyType, ", ", helper.ParseGoType(gen, fd.MapValue()), "]()")
					}
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
					if fd.MapValue().Kind() == protoreflect.MessageKind {
						g.P("First: ", treeMapPackage.Ident("New"), "[", keyType, ", *", helper.FindMessageGoIdent(gen, fd.MapValue().Message()), "](),")
					} else {
						g.P("First: ", treeMapPackage.Ident("New"), "[", keyType, ", ", helper.ParseGoType(gen, fd.MapValue()), "](),")
					}
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
				msg := helper.FindMessage(gen, fd.MapValue().Message())
				if msg != nil {
					genOrderedMapLoader(gen, depth+1, nextKeys, messagerName, g, msg, orderedMapValue)
				}
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

func genOrderedMapGetters(gen *protogen.Plugin, depth int, keys []helper.MapKey, messagerName string, file *protogen.File, g *protogen.GeneratedFile, message *protogen.Message) {
	if *disableOrderedMap {
		return
	}
	if depth == 1 && !helper.NeedGenOrderedMap(message.Desc) {
		return
	}
	genGetterName := func(depth int) string {
		getter := "GetOrderedMap"
		if depth > 1 {
			getter = fmt.Sprintf("GetOrderedMap%v", depth-1)
		}
		return getter
	}
	for _, field := range message.Fields {
		fd := field.Desc
		if fd.IsMap() {
			getter := genGetterName(depth)
			prefix := parseOrderedMapPrefix(fd)
			orderedMap := prefix + orderedMapSuffix
			if depth == 1 {
				g.P("// ", getter, " returns the 1-level ordered map.")
				g.P("func (x *", messagerName, ") ", getter, "(", helper.GenGetParams(keys), ") *", orderedMap, "{")
				g.P("  return x.orderedMap ")
			} else {
				g.P("// ", getter, " finds value in the ", depth-1, "-level ordered map. It will return nil if")
				g.P("// the deepest key is not found, otherwise return an error.")
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
				g.P(`return nil, `, errorsPackage.Ident("Errorf"), `(`, codePackage.Ident("NotFound"), `, "`, lastKeyName, `(%v) not found", `, lastKeyName, `)`)
				g.P("} else {")
				g.P(`return val.First, nil`)
				g.P("}")

			}
			g.P("}")
			g.P()

			nextKeys := helper.AddMapKey(gen, fd, keys)
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				msg := helper.FindMessage(gen, fd.MapValue().Message())
				if msg != nil {
					genOrderedMapGetters(gen, depth+1, nextKeys, messagerName, file, g, msg)
				}
			}
			break
		}
	}
}

func parseOrderedMapPrefix(mapFd protoreflect.FieldDescriptor) string {
	return strcase.ToCamel(string(mapFd.FullName()))
}
