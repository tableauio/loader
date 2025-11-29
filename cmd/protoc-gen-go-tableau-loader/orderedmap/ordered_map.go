package orderedmap

import (
	"fmt"
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Generator struct {
	gen     *protogen.Plugin
	g       *protogen.GeneratedFile
	message *protogen.Message
}

func NewGenerator(gen *protogen.Plugin, g *protogen.GeneratedFile, message *protogen.Message) *Generator {
	return &Generator{
		gen:     gen,
		g:       g,
		message: message,
	}
}

func (x *Generator) needGenerate() bool {
	return options.NeedGenOrderedMap(x.message.Desc, options.LangGO)
}

func (x *Generator) messagerName() string {
	return string(x.message.Desc.Name())
}

func (x *Generator) orderedMapPrefix(mapFd protoreflect.FieldDescriptor) string {
	if mapFd.MapValue().Kind() == protoreflect.MessageKind {
		localMsgProtoName := strings.TrimPrefix(string(mapFd.MapValue().Message().FullName()), string(x.message.Desc.FullName())+".")
		return strings.ReplaceAll(localMsgProtoName, ".", "_")
	}
	return mapFd.MapValue().Kind().String()
}

func (x *Generator) mapType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("%s_OrderedMap_%sMap", x.messagerName(), x.orderedMapPrefix(mapFd))
}

func (x *Generator) mapValueType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("%s_OrderedMap_%sValue", x.messagerName(), x.orderedMapPrefix(mapFd))
}

func (x *Generator) mapValueFieldType(fd protoreflect.FieldDescriptor) string {
	nextMapFD := getNextLevelMapFD(fd.MapValue())
	if nextMapFD != nil {
		return "*" + x.mapValueType(fd)
	}
	return helper.ParseMapValueType(x.gen, x.g, fd)
}

func (x *Generator) GenOrderedMapTypeDef() {
	if !x.needGenerate() {
		return
	}
	x.g.P("// OrderedMap types.")
	x.genOrderedMapTypeDef(x.message.Desc, 1, nil)
}

func (x *Generator) genOrderedMapTypeDef(md protoreflect.MessageDescriptor, depth int, keys helper.MapKeys) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			nextKeys := keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			keyType := nextKeys[len(nextKeys)-1].Type
			if keyType == "bool" {
				keyType = "int"
			}
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				x.genOrderedMapTypeDef(fd.MapValue().Message(), depth+1, nextKeys)
			}
			orderedMap := x.mapType(fd)
			orderedMapValue := x.mapValueType(fd)
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				currValueType := helper.FindMessageGoIdent(x.gen, fd.MapValue().Message())
				nextOrderedMap := x.mapType(nextMapFD)
				x.g.P("type ", orderedMapValue, "= ", helper.PairPackage.Ident("Pair"), "[*", nextOrderedMap, ", *", currValueType, "];")
			}
			x.g.P("type ", orderedMap, "= ", helper.TreeMapPackage.Ident("TreeMap"), "[", keyType, ", ", x.mapValueFieldType(fd), "]")
			x.g.P()
			return
		}
	}
}

func (x *Generator) GenOrderedMapField() {
	if !x.needGenerate() {
		return
	}
	md := x.message.Desc
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			x.g.P("orderedMap *", x.mapType(fd))
			return
		}
	}
}

func (x *Generator) GenOrderedMapLoader() {
	if !x.needGenerate() {
		return
	}
	x.g.P("// OrderedMap init.")
	x.genOrderedMapLoader(x.message.Desc, 1, nil, "")
}

func (x *Generator) genOrderedMapLoader(md protoreflect.MessageDescriptor, depth int, keys helper.MapKeys, lastOrderedMapValue string) {
	message := helper.FindMessage(x.gen, md)
	for _, field := range message.Fields {
		fd := field.Desc
		if fd.IsMap() {
			needConvertBool := len(keys) > 0 && keys[len(keys)-1].Type == "int"
			nextKeys := keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			keyType := nextKeys[len(nextKeys)-1].Type
			needConvertBoolNext := keyType == "bool"
			if keyType == "bool" {
				keyType = "int"
			}
			orderedMapValue := x.mapValueType(fd)
			mapName := fmt.Sprintf("x.Data().Get%s()", field.GoName)
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if depth == 1 {
				x.g.P("x.orderedMap = ", helper.TreeMapPackage.Ident("New"), "[", keyType, ", ", x.mapValueFieldType(fd), "]()")
			}
			if depth != 1 {
				mapName = fmt.Sprintf("v%d.Get%s()", depth-1, field.GoName)
				keyName := fmt.Sprintf("k%d", depth-1)
				if needConvertBool {
					keyName = fmt.Sprintf("boolToInt(%s)", keyName)
				}
				x.g.P("k", depth-1, "v := &", lastOrderedMapValue, "{")
				x.g.P("First: ", helper.TreeMapPackage.Ident("New"), "[", keyType, ", ", x.mapValueFieldType(fd), "](),")
				x.g.P("Second: v", depth-1, ",")
				x.g.P("}")
				x.g.P("map", depth-1, ".Put(", keyName, ", k", depth-1, "v)")
			}
			x.g.P("for k", depth, ", v", depth, " := range ", mapName, "{")
			if depth == 1 {
				x.g.P("map", depth, " := x.orderedMap")
			} else {
				x.g.P("map", depth, " := k", depth-1, "v.First")
			}
			if nextMapFD != nil {
				x.genOrderedMapLoader(fd.MapValue().Message(), depth+1, nextKeys, orderedMapValue)
			} else {
				keyName := fmt.Sprintf("k%d", depth)
				if needConvertBoolNext {
					keyName = fmt.Sprintf("boolToInt(%s)", keyName)
				}
				x.g.P("map", depth, ".Put(", keyName, ", v", depth, ")")
			}
			x.g.P("}")
			break
		}
	}
}

func (x *Generator) GenOrderedMapGetters() {
	if !x.needGenerate() {
		return
	}
	x.genOrderedMapGetters(x.message.Desc, 1, nil)
}

func (x *Generator) genOrderedMapGetters(md protoreflect.MessageDescriptor, depth int, keys helper.MapKeys) {
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
			orderedMap := x.mapType(fd)
			if depth == 1 {
				x.g.P("// ", getter, " returns the 1-level ordered map.")
				x.g.P("func (x *", x.messagerName(), ") ", getter, "(", keys.GenGetParams(), ") *", orderedMap, "{")
				x.g.P("return x.orderedMap ")
			} else {
				x.g.P("// ", getter, " finds value in the ", depth-1, "-level ordered map. It will return")
				x.g.P("// NotFound error if the key is not found.")
				x.g.P("func (x *", x.messagerName(), ") ", getter, "(", keys.GenGetParams(), ") (*", orderedMap, ", error) {")
				if depth == 2 {
					x.g.P("conf := x.orderedMap")
				} else {
					prevKeys := keys[:len(keys)-1]
					prevGetter := genGetterName(depth - 1)
					x.g.P("conf, err := x.", prevGetter, "(", prevKeys.GenGetArguments(), ")")
					x.g.P("if err != nil {")
					x.g.P(`return nil, err`)
					x.g.P("}")
				}
				lastKeyName := keys[len(keys)-1].Name
				lastKeyType := keys[len(keys)-1].Type
				keyName := lastKeyName
				if lastKeyType == "int" {
					keyName = fmt.Sprintf("boolToInt(%s)", keyName)
				}
				x.g.P("if val, ok := conf.Get(", keyName, "); !ok {")
				x.g.P(`return nil, `, helper.FmtPackage.Ident("Errorf"), `("`, lastKeyName, `(%v) %w", `, lastKeyName, `, ErrNotFound)`)
				x.g.P("} else {")
				x.g.P(`return val.First, nil`)
				x.g.P("}")

			}
			x.g.P("}")
			x.g.P()

			nextKeys := keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				x.genOrderedMapGetters(fd.MapValue().Message(), depth+1, nextKeys)
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
