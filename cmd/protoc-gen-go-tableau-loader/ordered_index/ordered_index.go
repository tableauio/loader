package orderedindex

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Generator struct {
	gen        *protogen.Plugin
	g          *protogen.GeneratedFile
	descriptor *index.IndexDescriptor
	message    *protogen.Message
}

func NewGenerator(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, message *protogen.Message) *Generator {
	return &Generator{
		gen:        gen,
		g:          g,
		descriptor: descriptor,
		message:    message,
	}
}

func (x *Generator) generate() bool {
	return options.NeedGenOrderedIndex(x.message.Desc, options.LangGO)
}

func (x *Generator) messagerName() string {
	return string(x.message.Desc.Name())
}

func (x *Generator) mapType(index *index.LevelIndex) string {
	return fmt.Sprintf("OrderedIndex_%s_%sMap", x.messagerName(), index.Name())
}

func (x *Generator) mapKeyType(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take first field
		return helper.ParseOrderedMapKeyType(x.gen, x.g, field.FD)
	} else {
		// multi-column index
		return fmt.Sprintf("OrderedIndex_%s_%sKey", x.messagerName(), index.Name())
	}
}

func (x *Generator) mapValueType(index *index.LevelIndex) protogen.GoIdent {
	return helper.FindMessageGoIdent(x.gen, index.MD)
}

func (x *Generator) indexContainerName(index *index.LevelIndex) string {
	return fmt.Sprintf("orderedIndex%sMap", strcase.ToCamel(index.Name()))
}

func (x *Generator) mapCtor(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		return "New"
	} else {
		// multi-column index
		return "New2"
	}
}

func (x *Generator) indexKeys(index *index.LevelIndex) helper.MapKeys {
	var keys []helper.MapKey
	for _, field := range index.ColFields {
		keys = append(keys, helper.MapKey{
			Type:      helper.ParseOrderedMapKeyType(x.gen, x.g, field.FD),
			Name:      helper.ParseIndexFieldNameAsFuncParam(x.gen, field.FD),
			FieldName: helper.ParseIndexFieldNameAsKeyStructFieldName(x.gen, field.FD),
		})
	}
	return keys
}

func (x *Generator) fieldGetter(fd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf(".Get%s()", helper.ParseIndexFieldName(x.gen, fd))
}

func (x *Generator) parseKeyFieldNameAndSuffix(field *index.LevelField) (string, string) {
	var fieldName, suffix string
	for i, leveledFd := range field.LeveledFDList {
		fieldName += x.fieldGetter(leveledFd)
		if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
			switch leveledFd.Message().FullName() {
			case "google.protobuf.Timestamp", "google.protobuf.Duration":
				suffix = ".GetSeconds()"
			default:
			}
		}
	}
	return fieldName, suffix
}

func (x *Generator) GenOrderedIndexTypeDef() {
	if !x.generate() {
		return
	}
	x.g.P("// OrderedIndex types.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P("// OrderedIndex: ", index.Index)
			if len(index.ColFields) != 1 {
				// multi-column index
				keyType := x.mapKeyType(index)
				keys := x.indexKeys(index)

				// generate key struct
				x.g.P("type ", keyType, " struct {")
				for _, key := range keys {
					x.g.P(key.FieldName, " ", key.Type)
				}
				x.g.P("}")
				x.g.P()

				// generate Less func to implement cmp.Ordered interface
				x.g.P("func (x ", keyType, ") Less(other ", keyType, ") bool {")
				for i, key := range keys {
					if i == len(keys)-1 {
						x.g.P("return x.", key.FieldName, " < other.", key.FieldName)
					} else {
						x.g.P("if x.", key.FieldName, " != other.", key.FieldName, " {")
						x.g.P("return x.", key.FieldName, " < other.", key.FieldName)
						x.g.P("}")
					}
				}
				x.g.P("}")
				x.g.P()
			}
			x.g.P("type ", x.mapType(index), " = ", helper.TreeMapPackage.Ident("TreeMap"), "[", x.mapKeyType(index), ", []*", x.mapValueType(index), "]")
			x.g.P()
		}
	}
}

func (x *Generator) GenOrderedIndexField() {
	if !x.generate() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P(x.indexContainerName(index), " *", x.mapType(index))
		}
	}
}

func (x *Generator) GenOrderedIndexLoader() {
	if !x.generate() {
		return
	}
	x.g.P("// OrderedIndex init.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P("x.", x.indexContainerName(index), " = ", helper.TreeMapPackage.Ident(x.mapCtor(index)), "[", x.mapKeyType(index), ", []*", x.mapValueType(index), "]()")
		}
	}
	parentDataName := "x.data"
	depth := 1
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.genOneOrderedIndexLoader(index, depth, parentDataName)
		}
		itemName := fmt.Sprintf("item%d", depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGen() {
			break
		}
		x.g.P("for _, ", itemName, " := range ", parentDataName, x.fieldGetter(levelMessage.FD), " {")
		parentDataName = itemName
		depth++
	}
	for i := depth - 1; i > 0; i-- {
		x.g.P("}")
	}
	x.genOrderedIndexSorter()
}

func (x *Generator) genOneOrderedIndexLoader(index *index.LevelIndex, depth int, parentDataName string) {
	x.g.P("{")
	x.g.P("// OrderedIndex: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		fieldName, suffix := x.parseKeyFieldNameAndSuffix(field)
		indexContainerName := x.indexContainerName(index)
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			x.g.P("for _, ", itemName, " := range ", parentDataName, fieldName, " {")
			x.g.P("key := ", itemName, suffix)
			x.g.P("value, _ := x.", indexContainerName, ".Get(key)")
			x.g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
			x.g.P("}")
		} else {
			x.g.P("key := ", parentDataName, fieldName, suffix)
			x.g.P("value, _ := x.", indexContainerName, ".Get(key)")
			x.g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
		}
	} else {
		// multi-column index
		x.generateOneMulticolumnOrderedIndex(depth, index, parentDataName, nil)
	}
	x.g.P("}")
}

func (x *Generator) generateOneMulticolumnOrderedIndex(depth int, index *index.LevelIndex, parentDataName string, keys helper.MapKeys) {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		keyType := x.mapKeyType(index)
		indexContainerName := x.indexContainerName(index)
		x.g.P("key := ", keyType, " {", keys.GenGetArguments(), "}")
		x.g.P("value, _ := x.", indexContainerName, ".Get(key)")
		x.g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
		return
	}
	field := index.ColFields[cursor]
	fieldName, suffix := x.parseKeyFieldNameAndSuffix(field)
	if field.FD.IsList() {
		itemName := fmt.Sprintf("indexItem%d", cursor)
		x.g.P("for _, ", itemName, " := range ", parentDataName, fieldName, " {")
		key := itemName + suffix
		keys = append(keys, helper.MapKey{Name: key})
		x.generateOneMulticolumnOrderedIndex(depth+1, index, parentDataName, keys)
		x.g.P("}")
	} else {
		key := parentDataName + fieldName + suffix
		keys = append(keys, helper.MapKey{Name: key})
		x.generateOneMulticolumnOrderedIndex(depth, index, parentDataName, keys)
	}
}

func (x *Generator) genOrderedIndexSorter() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			if len(index.SortedColFields) != 0 {
				x.g.P("// OrderedIndex(sort): ", index.Index)
				x.g.P("x.", x.indexContainerName(index), ".Range(func(key ", x.mapKeyType(index), ", item []*", x.mapValueType(index), ") bool {")
				x.g.P(helper.SortPackage.Ident("Slice"), "(item, func(i, j int) bool {")
				for i, field := range index.SortedColFields {
					fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
					if i == len(index.SortedColFields)-1 {
						x.g.P("return item[i]", fieldName, " < item[j]", fieldName)
					} else {
						x.g.P("if item[i]", fieldName, " != item[j]", fieldName, " {")
						x.g.P("return item[i]", fieldName, " < item[j]", fieldName)
						x.g.P("}")
					}
				}
				x.g.P("})")
				x.g.P("return true")
				x.g.P("})")
			}
		}
	}
}

func (x *Generator) GenOrderedIndexFinders() {
	if !x.generate() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			indexContainerName := x.indexContainerName(index)
			messagerName := x.messagerName()
			x.g.P("// OrderedIndex: ", index.Index)
			x.g.P()

			x.g.P("// Find", index.Name(), "Map finds the ordered index (", index.Index, ") to value (", x.mapValueType(index), ") treemap.")
			x.g.P("// One key may correspond to multiple values, which are contained by a slice.")
			x.g.P("func (x *", messagerName, ") Find", index.Name(), "Map() *", x.mapType(index), " {")
			x.g.P("return x.", indexContainerName)
			x.g.P("}")
			x.g.P()

			keys := x.indexKeys(index)
			params := keys.GenGetParams()
			args := keys.GenGetArguments()
			x.g.P("// Find", index.Name(), " finds a slice of all values of the given key.")
			x.g.P("func (x *", messagerName, ") Find", index.Name(), "(", params, ") []*", x.mapValueType(index), " {")
			if len(index.ColFields) == 1 {
				x.g.P("val, _ := x.", indexContainerName, ".Get(", args, ")")
			} else {
				x.g.P("val, _ := x.", indexContainerName, ".Get(", x.mapKeyType(index), "{", args, "})")
			}
			x.g.P("return val")
			x.g.P("}")
			x.g.P()

			x.g.P("// FindFirst", index.Name(), " finds the first value of the given key,")
			x.g.P("// or nil if no value found.")
			x.g.P("func (x *", messagerName, ") FindFirst", index.Name(), "(", params, ") *", x.mapValueType(index), " {")
			x.g.P("val := x.Find", index.Name(), "(", args, ")")
			x.g.P("if len(val) > 0 {")
			x.g.P("return val[0]")
			x.g.P("}")
			x.g.P("return nil")
			x.g.P("}")
			x.g.P()
		}
	}
}
