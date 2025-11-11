package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
)

func genOrderedIndexTypeDef(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	g.P("// OrderedIndex types.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			if len(index.ColFields) == 1 {
				// single-column index
				field := index.ColFields[0] // just take first field
				g.P("// OrderedIndex: ", index.Index)
				mapType := fmt.Sprintf("%s_OrderedIndex_%sMap", messagerName, index.Name())
				keyType := helper.ParseOrderedMapKeyType(gen, g, field.FD)
				g.P("type ", mapType, " = ", treeMapPackage.Ident("TreeMap"), "[", keyType, ", []*", helper.FindMessageGoIdent(gen, index.MD), "]")
			} else {
				// multi-column index
				g.P("// OrderedIndex: ", index.Index)
				keyType := fmt.Sprintf("%s_OrderedIndex_%sKey", messagerName, index.Name())
				mapType := fmt.Sprintf("%s_OrderedIndex_%sMap", messagerName, index.Name())

				// generate key struct
				// KeyType must be comparable, refer https://go.dev/blog/maps
				g.P("type ", keyType, " struct {")
				for _, field := range index.ColFields {
					g.P(helper.ParseIndexFieldNameAsKeyStructFieldName(gen, field.FD), " ", helper.ParseOrderedMapKeyType(gen, g, field.FD))
				}
				g.P("}")
				g.P()
				g.P("func (x ", keyType, ") Less(other ", keyType, ") bool {")
				for i, field := range index.ColFields {
					fieldName := helper.ParseIndexFieldNameAsKeyStructFieldName(gen, field.FD)
					if i == len(index.ColFields)-1 {
						g.P("return x.", fieldName, " < other.", fieldName)
					} else {
						g.P("if x.", fieldName, " != other.", fieldName, " {")
						g.P("return x.", fieldName, " < other.", fieldName)
						g.P("}")
					}
				}
				g.P("}")
				g.P()
				g.P("type ", mapType, " = ", treeMapPackage.Ident("TreeMap"), "[", keyType, ", []*", helper.FindMessageGoIdent(gen, index.MD), "]")
			}
			g.P()
		}
	}
}

func genOrderedIndexField(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			mapType := fmt.Sprintf("%s_OrderedIndex_%sMap", messagerName, index.Name())
			g.P(indexContainerName, " *", mapType)
		}
	}
}

func genOrderedIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	g.P("// OrderedIndex init.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			field := index.ColFields[0] // just take first field
			indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			if len(index.ColFields) == 1 {
				keyType := helper.ParseOrderedMapKeyType(gen, g, field.FD)
				g.P("x.", indexContainerName, " = ", treeMapPackage.Ident("New"), "[", keyType, ", []*", helper.FindMessageGoIdent(gen, index.MD), "]()")
			} else {
				keyType := fmt.Sprintf("%s_OrderedIndex_%sKey", messagerName, index.Name())
				g.P("x.", indexContainerName, " = ", treeMapPackage.Ident("New2"), "[", keyType, ", []*", helper.FindMessageGoIdent(gen, index.MD), "]()")
			}
		}
	}
	parentDataName := "x.data"
	depth := 1
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			genOneOrderedIndexLoader(gen, g, depth, index, parentDataName, messagerName)
		}
		itemName := fmt.Sprintf("item%d", depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGen() {
			break
		}
		g.P("for _, ", itemName, " := range ", parentDataName, ".Get", helper.ParseIndexFieldName(gen, levelMessage.FD), "() {")
		parentDataName = itemName
		depth++
	}
	for i := depth - 1; i > 0; i-- {
		g.P("}")
	}
	genOrderedIndexSorter(gen, g, descriptor, messagerName)
}

func genOneOrderedIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, depth int, index *index.LevelIndex,
	parentDataName string, messagerName string) {
	indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
	g.P("{")
	g.P("// OrderedIndex: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			fieldName, suffix := parseOrderedIndexKeyFieldNameAndSuffix(gen, field)
			g.P("for _, ", itemName, " := range ", parentDataName, fieldName, " {")
			g.P("key := ", itemName, suffix)
			g.P("value, _ := x.", indexContainerName, ".Get(key)")
			g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
			g.P("}")
		} else {
			fieldName, suffix := parseOrderedIndexKeyFieldNameAndSuffix(gen, field)
			g.P("key := ", parentDataName, fieldName, suffix)
			g.P("value, _ := x.", indexContainerName, ".Get(key)")
			g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
		}
	} else {
		// multi-column index
		generateOneMulticolumnOrderedIndex(gen, g, depth, index, parentDataName, messagerName, nil)
	}
	g.P("}")
}

func parseOrderedIndexKeyFieldNameAndSuffix(gen *protogen.Plugin, field *index.LevelField) (string, string) {
	var fieldName, suffix string
	for i, leveledFd := range field.LeveledFDList {
		fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
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

func generateOneMulticolumnOrderedIndex(gen *protogen.Plugin, g *protogen.GeneratedFile,
	depth int, index *index.LevelIndex, parentDataName string, messagerName string, keys []string) []string {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		var keyParams string
		for i, key := range keys {
			keyParams += key
			if i != len(keys)-1 {
				keyParams += ", "
			}
		}
		keyType := fmt.Sprintf("%s_OrderedIndex_%sKey", messagerName, index.Name())
		indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
		g.P("key := ", keyType, " {", keyParams, "}")
		g.P("value, _ := x.", indexContainerName, ".Get(key)")
		g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
		return keys
	}
	field := index.ColFields[cursor]
	if field.FD.IsList() {
		itemName := fmt.Sprintf("indexItem%d", cursor)
		fieldName, suffix := parseOrderedIndexKeyFieldNameAndSuffix(gen, field)
		g.P("for _, ", itemName, " := range ", parentDataName, fieldName, " {")
		key := itemName + suffix
		keys = append(keys, key) // fixme:keys
		keys = generateOneMulticolumnOrderedIndex(gen, g, depth+1, index, parentDataName, messagerName, keys)
		g.P("}")
	} else {
		fieldName, suffix := parseOrderedIndexKeyFieldNameAndSuffix(gen, field)
		key := parentDataName + fieldName + suffix
		keys = append(keys, key)
		keys = generateOneMulticolumnOrderedIndex(gen, g, depth, index, parentDataName, messagerName, keys)
	}
	return keys
}

func genOrderedIndexSorter(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			if len(index.SortedColFields) != 0 {
				g.P("// OrderedIndex(sort): ", index.Index)
				var keyType string
				if len(index.ColFields) == 1 {
					// single-column index
					field := index.ColFields[0] // just take first field
					keyType = helper.ParseOrderedMapKeyType(gen, g, field.FD)
				} else {
					// multi-column index
					keyType = fmt.Sprintf("%s_OrderedIndex_%sKey", messagerName, index.Name())
				}
				g.P("x.", indexContainerName, ".Range(func(key ", keyType, ", item []*", helper.FindMessageGoIdent(gen, index.MD), ") bool {")
				g.P(sortPackage.Ident("Slice"), "(item, func(i, j int) bool {")
				for i, field := range index.SortedColFields {
					fieldName := ""
					for _, leveledFd := range field.LeveledFDList {
						fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
					}
					if i == len(index.SortedColFields)-1 {
						g.P("return item[i]", fieldName, " < item[j]", fieldName)
					} else {
						g.P("if item[i]", fieldName, " != item[j]", fieldName, " {")
						g.P("return item[i]", fieldName, " < item[j]", fieldName)
						g.P("}")
					}
				}
				g.P("})")
				g.P("return true")
				g.P("})")
			}
		}
	}
}

func genOrderedIndexFinders(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			mapType := fmt.Sprintf("%s_OrderedIndex_%sMap", messagerName, index.Name())
			indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"

			g.P("// OrderedIndex: ", index.Index)
			g.P()

			g.P("// Find", index.Name(), "Map finds the ordered index (", index.Index, ") to value (", helper.FindMessageGoIdent(gen, index.MD), ") treemap.")
			g.P("// One key may correspond to multiple values, which are contained by a slice.")
			g.P("func (x *", messagerName, ") Find", index.Name(), "Map() *", mapType, " {")
			g.P("return x.", indexContainerName)
			g.P("}")
			g.P()

			var keys []helper.MapKey
			for _, field := range index.ColFields {
				keys = append(keys, helper.MapKey{
					Type: helper.ParseOrderedMapKeyType(gen, g, field.FD),
					Name: helper.ParseIndexFieldNameAsFuncParam(gen, field.FD),
				})
			}

			g.P("// Find", index.Name(), " finds a slice of all values of the given key.")
			g.P("func (x *", messagerName, ") Find", index.Name(), "(", helper.GenGetParams(keys), ") []*", helper.FindMessageGoIdent(gen, index.MD), " {")
			if len(index.ColFields) == 1 {
				g.P("val, _ := x.", indexContainerName, ".Get(", helper.GenGetArguments(keys), ")")
				g.P("return val")
			} else {
				g.P("val, _ := x.", indexContainerName, ".Get(", fmt.Sprintf("%s_OrderedIndex_%sKey", messagerName, index.Name()), "{", helper.GenGetArguments(keys), "})")
				g.P("return val")
			}
			g.P("}")
			g.P()

			g.P("// FindFirst", index.Name(), " finds the first value of the given key,")
			g.P("// or nil if no value found.")
			g.P("func (x *", messagerName, ") FindFirst", index.Name(), "(", helper.GenGetParams(keys), ") *", helper.FindMessageGoIdent(gen, index.MD), " {")
			g.P("val := x.Find", index.Name(), "(", helper.GenGetArguments(keys), ")")
			g.P("if len(val) > 0 {")
			g.P("return val[0]")
			g.P("}")
			g.P("return nil")
			g.P("}")
			g.P()
		}
	}
}
