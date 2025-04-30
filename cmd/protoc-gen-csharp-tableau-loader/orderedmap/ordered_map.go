package orderedmap

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Generator struct {
	g       *protogen.GeneratedFile
	message *protogen.Message
}

func NewGenerator(g *protogen.GeneratedFile, message *protogen.Message) *Generator {
	return &Generator{
		g:       g,
		message: message,
	}
}

func (x *Generator) NeedGenerate() bool {
	return options.NeedGenOrderedMap(x.message.Desc, options.LangCS)
}

func (x *Generator) mapType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("OrderedMap_%sMap", helper.ParseLeveledMapPrefix(x.message.Desc, mapFd))
}

func (x *Generator) mapValueType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("OrderedMap_%sValue", helper.ParseLeveledMapPrefix(x.message.Desc, mapFd))
}

func (x *Generator) mapValueFieldType(fd protoreflect.FieldDescriptor) string {
	nextMapFD := getNextLevelMapFD(fd.MapValue())
	if nextMapFD != nil {
		return x.mapValueType(fd)
	}
	return helper.ParseMapValueType(fd)
}

func (x *Generator) GenOrderedMapTypeDef() {
	if !x.NeedGenerate() {
		return
	}
	x.g.P(helper.Indent(2), "// OrderedMap types.")
	x.genOrderedMapTypeDef(x.message.Desc, 1, nil)
}

func (x *Generator) genOrderedMapTypeDef(md protoreflect.MessageDescriptor, depth int, keys helper.MapKeySlice) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			nextKeys := keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldNameAsKeyStructFieldName(fd),
			})
			keyType := nextKeys[len(nextKeys)-1].Type
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				x.genOrderedMapTypeDef(fd.MapValue().Message(), depth+1, nextKeys)
			}
			orderedMap := x.mapType(fd)
			orderedMapValue := x.mapValueType(fd)
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				currValueType := helper.ParseCsharpType(fd.MapValue())
				nextOrderedMap := x.mapType(nextMapFD)
				x.g.P(helper.Indent(2), "public class ", orderedMapValue, "(", nextOrderedMap, " item1, ", currValueType, " item2)")
				x.g.P(helper.Indent(3), ": Tuple<", nextOrderedMap, ", ", currValueType, ">(item1, item2);")
			}
			x.g.P(helper.Indent(2), "public class ", orderedMap, " : SortedDictionary<", keyType, ", ", x.mapValueFieldType(fd), ">;")
			x.g.P()
			if depth == 1 {
				x.g.P(helper.Indent(2), "private ", orderedMap, " _orderedMap = [];")
				x.g.P()
			}
			return
		}
	}
}

func (x *Generator) GenOrderedMapLoader() {
	if !x.NeedGenerate() {
		return
	}
	x.g.P(helper.Indent(3), "// OrderedMap init.")
	x.g.P(helper.Indent(3), "_orderedMap.Clear();")
	x.genOrderedMapLoader(x.message.Desc, 1)
}

func (x *Generator) genOrderedMapLoader(md protoreflect.MessageDescriptor, depth int) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			orderedMapValue := x.mapValueType(fd)
			keyName := fmt.Sprintf("key%d", depth)
			valueName := fmt.Sprintf("value%d", depth)

			tmpOrderedMapName := fmt.Sprintf("ordered_map%d", depth)

			prevContainer := fmt.Sprintf("value%d", depth-1)
			prevTmpOrderedMapName := fmt.Sprintf("ordered_map%d", depth-1)
			if depth == 1 {
				prevContainer = "_data"
				prevTmpOrderedMapName = "_orderedMap"
			}
			x.g.P(helper.Indent(depth+2), "foreach (var (", keyName, ", ", valueName, ") in ", prevContainer, ".", strcase.ToCamel(string(fd.Name())), ")")
			x.g.P(helper.Indent(depth+2), "{")
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				nextOrderedMap := x.mapType(nextMapFD)
				x.g.P(helper.Indent(depth+3), "var ", tmpOrderedMapName, " = new ", nextOrderedMap, "();")
			}
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				x.genOrderedMapLoader(fd.MapValue().Message(), depth+1)
			}

			if nextMapFD != nil {
				x.g.P(helper.Indent(depth+3), prevTmpOrderedMapName, "[", keyName, "] = new ", orderedMapValue, "(", tmpOrderedMapName, ", ", valueName, ");")
			} else {
				x.g.P(helper.Indent(depth+3), prevTmpOrderedMapName, "[", keyName, "] = ", valueName, ";")
			}
			x.g.P(helper.Indent(depth+2), "}")
			break
		}
	}
}

func (x *Generator) GenOrderedMapGetters() {
	if !x.NeedGenerate() {
		return
	}
	x.genOrderedMapGetters(x.message.Desc, 1, nil)
}

func (x *Generator) genOrderedMapGetters(md protoreflect.MessageDescriptor, depth int, keys helper.MapKeySlice) {
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
			x.g.P()
			if depth == 1 {
				x.g.P(helper.Indent(2), "// OrderedMap accessers.")
			}
			getter := genGetterName(depth)
			orderedMap := x.mapType(fd)
			if depth == 1 {
				x.g.P(helper.Indent(2), "public ref readonly ", orderedMap, " ", getter, "() => ref _orderedMap;")
			} else {
				lastKeyName := keys[len(keys)-1].Name
				if depth == 2 {
					x.g.P(helper.Indent(2), "public ", orderedMap, "? ", getter, "(", keys.GenGetParams(), ") =>")
					x.g.P(helper.Indent(3), "_orderedMap.TryGetValue(", lastKeyName, ", out var value) ? value.Item1 : null;")
				} else {
					prevKeys := keys[:len(keys)-1]
					prevGetter := genGetterName(depth - 1)
					x.g.P(helper.Indent(2), "public ", orderedMap, "? ", getter, "(", keys.GenGetParams(), ") =>")
					x.g.P(helper.Indent(3), prevGetter, "(", prevKeys.GenGetArguments(), ")?.TryGetValue(", lastKeyName, ", out var value) == true ? value.Item1 : null;")
				}
			}

			nextKeys := keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldNameAsFuncParam(fd),
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
