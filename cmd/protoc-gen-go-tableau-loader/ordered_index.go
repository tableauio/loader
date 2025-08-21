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
			// single-column index
			field := index.ColFields[0] // just take first field
			g.P("// OrderedIndex: ", index.Index)
			mapType := fmt.Sprintf("%s_OrderedIndex_%sMap", messagerName, index.Name())
			keyType := helper.ParseOrderedMapKeyType(field.FD)
			g.P("type ", mapType, " = ", treeMapPackage.Ident("TreeMap"), "[", keyType, ", []*", helper.FindMessageGoIdent(gen, index.MD), "]")
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

func genOrderedIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P("// OrderedIndex init.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			field := index.ColFields[0] // just take first field
			indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			keyType := helper.ParseOrderedMapKeyType(field.FD)
			g.P("x.", indexContainerName, " = ", treeMapPackage.Ident("New"), "[", keyType, ", []*", helper.FindMessageGoIdent(gen, index.MD), "]()")
		}
	}
	parentDataName := "x.data"
	depth := 1
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			genOneOrderedIndexLoader(gen, g, depth, index, parentDataName)
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
	genOrderedIndexSorter(gen, g, descriptor)
}

func genOneOrderedIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, depth int, index *index.LevelIndex,
	parentDataName string) {
	indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
	g.P("{")
	g.P("// OrderedIndex: ", index.Index)
	// single-column index
	field := index.ColFields[0] // just take the first field
	if field.FD.IsList() {
		itemName := fmt.Sprintf("item%d", depth)
		fieldName := ""
		suffix := ""
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
		g.P("for _, ", itemName, " := range ", parentDataName, fieldName, " {")
		g.P("key := ", itemName, suffix)
		g.P("value, _ := x.", indexContainerName, ".Get(key)")
		g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
		g.P("}")
	} else {
		fieldName := ""
		suffix := ""
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
		g.P("key := ", parentDataName, fieldName, suffix)
		g.P("value, _ := x.", indexContainerName, ".Get(key)")
		g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
	}

	g.P("}")
}

func genOrderedIndexSorter(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			indexContainerName := "orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			if len(index.SortedColFields) != 0 {
				g.P("// OrderedIndex(sort): ", index.Index)
				g.P("x.", indexContainerName, ".Range(func(key int64, item []*", helper.FindMessageGoIdent(gen, index.MD), ") bool {")
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

			g.P("// Find", index.Name(), "OrderedMap returns the index(", index.Index, ") to value(", helper.FindMessageGoIdent(gen, index.MD), ") treemap.")
			g.P("// One key may correspond to multiple values, which are contained by a slice.")
			g.P("func (x *", messagerName, ") Find", index.Name(), "OrderedMap() *", mapType, " {")
			g.P("return x.", indexContainerName)
			g.P("}")
			g.P()
		}
	}
}
