package orderedmap

import (
	"fmt"
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
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
	return options.NeedGenOrderedMap(x.message.Desc, options.LangCPP)
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
	return fmt.Sprintf("OrderedMap_%sMap", x.orderedMapPrefix(mapFd))
}

func (x *Generator) mapValueType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("OrderedMap_%sValue", x.orderedMapPrefix(mapFd))
}

func (x *Generator) mapValueFieldType(fd protoreflect.FieldDescriptor) string {
	nextMapFD := getNextLevelMapFD(fd.MapValue())
	if nextMapFD != nil {
		return x.mapValueType(fd)
	}
	return helper.ParseMapValueType(fd)
}

func (x *Generator) GenHppOrderedMapGetters() {
	if !x.NeedGenerate() {
		return
	}
	x.g.P()
	x.g.P(helper.Indent(1), "// OrderedMap accessers.")
	x.g.P(" public:")
	x.genHppOrderedMapGetters(x.message.Desc, 1, nil)
}

func (x *Generator) genHppOrderedMapGetters(md protoreflect.MessageDescriptor, depth int, keys helper.MapKeys) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			nextKeys := keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			keyType := nextKeys[len(nextKeys)-1].Type
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				x.genHppOrderedMapGetters(fd.MapValue().Message(), depth+1, nextKeys)
			}
			orderedMap := x.mapType(fd)
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				orderedMapValue := x.mapValueType(fd)
				currValueType := helper.ParseCppType(fd.MapValue())
				nextOrderedMap := x.mapType(nextMapFD)
				x.g.P(helper.Indent(1), "using ", orderedMapValue, " = std::pair<", nextOrderedMap, ", const ", currValueType, "*>;")
			}
			x.g.P(helper.Indent(1), "using ", orderedMap, " = std::map<", keyType, ", ", x.mapValueFieldType(fd), ">;")
			x.g.P(helper.Indent(1), "const ", orderedMap, "* GetOrderedMap(", keys.GenGetParams(), ") const;")
			x.g.P()
			if depth == 1 {
				x.g.P(" private:")
				x.g.P(helper.Indent(1), orderedMap, " ordered_map_;")
			}
			return
		}
	}
}

func (x *Generator) GenOrderedMapLoader() {
	if !x.NeedGenerate() {
		return
	}
	x.g.P(helper.Indent(1), "// OrderedMap init.")
	x.g.P(helper.Indent(1), "ordered_map_.clear();")
	x.genOrderedMapLoader(x.message.Desc, 1)
}

func (x *Generator) genOrderedMapLoader(md protoreflect.MessageDescriptor, depth int) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			orderedMapValue := x.mapValueType(fd)
			itemName := fmt.Sprintf("item%d", depth)

			tmpOrderedMapName := fmt.Sprintf("ordered_map%d", depth)

			prevItemName := fmt.Sprintf("item%d", depth-1)
			prevContainer := prevItemName + ".second"
			prevTmpOrderedMapName := fmt.Sprintf("ordered_map%d", depth-1)
			if depth == 1 {
				prevContainer = "data_"
				prevTmpOrderedMapName = "ordered_map_"
			}
			x.g.P(helper.Indent(depth), "for (auto&& ", itemName, " : ", prevContainer, ".", string(fd.Name()), "()) {")
			nextMapFD := getNextLevelMapFD(fd.MapValue())
			if nextMapFD != nil {
				nextOrderedMap := x.mapType(nextMapFD)
				x.g.P(helper.Indent(depth+1), prevTmpOrderedMapName, "[", itemName, ".first] = ", orderedMapValue, "(", nextOrderedMap, "(), &", itemName, ".second);")
				x.g.P(helper.Indent(depth+1), "auto&& ", tmpOrderedMapName, " = ", prevTmpOrderedMapName, "[", itemName, ".first].first;")
			} else {
				ref := "&"
				if fd.MapValue().Kind() != protoreflect.MessageKind {
					ref = "" // scalar value type just do value copy.
				}
				x.g.P(helper.Indent(depth+1), prevTmpOrderedMapName, "[", itemName, ".first] = ", ref, itemName, ".second;")
			}
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				x.genOrderedMapLoader(fd.MapValue().Message(), depth+1)
			}
			x.g.P(helper.Indent(depth), "}")
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

func (x *Generator) genOrderedMapGetters(md protoreflect.MessageDescriptor, depth int, keys helper.MapKeys) {
	messagerName := x.messagerName()
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			orderedMap := x.mapType(fd)

			x.g.P("const ", messagerName, "::", orderedMap, "* ", messagerName, "::GetOrderedMap(", keys.GenGetParams(), ") const {")
			if depth == 1 {
				x.g.P(helper.Indent(1), "return &ordered_map_; ")
			} else {
				lastKeyName := keys[len(keys)-1].Name
				prevKeys := keys[:len(keys)-1]
				x.g.P(helper.Indent(1), "const auto* conf = GetOrderedMap(", prevKeys.GenGetArguments(), ");")
				x.g.P(helper.Indent(1), "if (conf == nullptr) {")
				x.g.P(helper.Indent(2), "return nullptr;")
				x.g.P(helper.Indent(1), "}")
				x.g.P(helper.Indent(1), "auto iter = conf->find(", lastKeyName, ");")
				x.g.P(helper.Indent(1), "if (iter == conf->end()) {")
				x.g.P(helper.Indent(2), "return nullptr;")
				x.g.P(helper.Indent(1), "}")
				x.g.P(helper.Indent(1), "return &iter->second.first;")

			}
			x.g.P("}")
			x.g.P()

			keys = keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				x.genOrderedMapGetters(fd.MapValue().Message(), depth+1, keys)
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
