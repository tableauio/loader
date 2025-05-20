package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index/desc"
	"google.golang.org/protobuf/compiler/protogen"
)

func genIndexTypeDef(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *desc.IndexDescriptor, messagerName string) {
	g.P("// Index types.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			if len(index.ColFields) == 1 {
				// single-column index
				field := index.ColFields[0] // just take first field
				g.P("// Index: ", index.Index)
				mapType := fmt.Sprintf("%s_Index_%sMap", messagerName, index.Name())
				keyType := helper.ParseGoType(gen, field.FD)
				g.P("type ", mapType, " = map[", keyType, "][]*", helper.FindMessageGoIdent(gen, index.MD))
			} else {
				// multi-column index
				g.P("// Index: ", index.Index)
				keyType := fmt.Sprintf("%s_Index_%sKey", messagerName, index.Name())
				mapType := fmt.Sprintf("%s_Index_%sMap", messagerName, index.Name())

				// generate key struct
				// KeyType must be comparable, refer https://go.dev/blog/maps
				g.P("type ", keyType, " struct {")
				for _, field := range index.ColFields {
					g.P(helper.ParseIndexFieldNameAsKeyStructFieldName(gen, field.FD), " ", helper.ParseGoType(gen, field.FD))
				}
				g.P("}")
				g.P("type ", mapType, " = map[", keyType, "][]*", helper.FindMessageGoIdent(gen, index.MD))
			}
			g.P()
		}
	}
}

func genIndexField(g *protogen.GeneratedFile, descriptor *desc.IndexDescriptor, messagerName string) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			indexContainerName := "index" + strcase.ToCamel(index.Name()) + "Map"
			mapType := fmt.Sprintf("%s_Index_%sMap", messagerName, index.Name())
			g.P(indexContainerName, " ", mapType)
		}
	}
}

func genIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *desc.IndexDescriptor, messagerName string) {
	g.P("  // Index init.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			indexContainerName := "index" + strcase.ToCamel(index.Name()) + "Map"
			g.P("x.", indexContainerName, " = make(", fmt.Sprintf("%s_Index_%sMap", messagerName, index.Name()), ")")
		}
	}
	parentDataName := "x.data"
	depth := 1
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			genOneIndexLoader(gen, g, depth, index, parentDataName, messagerName)
		}
		itemName := fmt.Sprintf("item%d", depth)
		if levelMessage.FD == nil {
			break
		}
		g.P("for _, ", itemName, " := range "+parentDataName+".Get"+helper.ParseIndexFieldName(gen, levelMessage.FD)+"() {")
		parentDataName = itemName
		depth++
		defer g.P("}")
	}
}

func genOneIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, depth int, index *desc.LevelIndex,
	parentDataName string, messagerName string) {
	indexContainerName := "index" + strcase.ToCamel(index.Name()) + "Map"
	g.P("{")
	g.P("  // Index: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			fieldName := ""
			for _, leveledFd := range field.LeveledFDList {
				fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
			}
			g.P("for _, ", itemName, " := range "+parentDataName+fieldName+" {")
			g.P("key := ", itemName)
			g.P("x.", indexContainerName, "[key] = append(x.", indexContainerName, "[key], ", parentDataName, ")")
			g.P("}")
		} else {
			fieldName := ""
			for _, leveledFd := range field.LeveledFDList {
				fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
			}
			g.P("key := ", parentDataName+fieldName)
			g.P("x.", indexContainerName, "[key] = append(x.", indexContainerName, "[key], ", parentDataName, ")")
		}
	} else {
		// multi-column index
		generateOneMulticolumnIndex(gen, g, depth, index, parentDataName, messagerName, nil)
	}
	if len(index.KeyFields) != 0 {
		g.P("for _, item := range x.", indexContainerName, " {")
		g.P(sortPackage.Ident("Slice"), "(item, func(i, j int) bool {")
		for i, field := range index.KeyFields {
			fieldName := ""
			for _, leveledFd := range field.LeveledFDList {
				fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
			}
			if i == len(index.KeyFields)-1 {
				g.P("return item[i]", fieldName, " < item[j]", fieldName)
			} else {
				g.P("if item[i]", fieldName, " != item[j]", fieldName, " {")
				g.P("return item[i]", fieldName, " < item[j]", fieldName)
				g.P("}")
			}
		}
		g.P("})")
		g.P("}")
	}
	g.P("}")
}

func generateOneMulticolumnIndex(gen *protogen.Plugin, g *protogen.GeneratedFile,
	depth int, index *desc.LevelIndex, parentDataName string, messagerName string, keys []string) []string {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		var keyParams string
		for i, key := range keys {
			keyParams += key
			if i != len(keys)-1 {
				keyParams += ", "
			}
		}
		keyType := fmt.Sprintf("%s_Index_%sKey", messagerName, index.Name())
		indexContainerName := "index" + strcase.ToCamel(index.Name()) + "Map"
		g.P("key := ", keyType, " {", keyParams, "}")
		g.P("x.", indexContainerName, "[key] = append(x.", indexContainerName, "[key], ", parentDataName, ")")
		return keys
	}
	field := index.ColFields[cursor]
	if field.FD.IsList() {
		itemName := fmt.Sprintf("indexItem%d", cursor)
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
		}
		g.P("for _, " + itemName + " := range " + parentDataName + fieldName + " {")
		keys = append(keys, itemName)
		keys = generateOneMulticolumnIndex(gen, g, depth+1, index, parentDataName, messagerName, keys)
		g.P("}")
	} else {
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
		}
		key := parentDataName + fieldName
		keys = append(keys, key)
		keys = generateOneMulticolumnIndex(gen, g, depth, index, parentDataName, messagerName, keys)
	}
	return keys
}

func genIndexFinders(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *desc.IndexDescriptor, messagerName string) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			mapType := fmt.Sprintf("%s_Index_%sMap", messagerName, index.Name())
			indexContainerName := "index" + strcase.ToCamel(index.Name()) + "Map"

			g.P("// Index: ", index.Index)
			g.P()

			g.P("// Find", index.Name(), "Map returns the index(", index.Index, ") to value(", helper.FindMessageGoIdent(gen, index.MD), ") map.")
			g.P("// One key may correspond to multiple values, which are contained by a slice.")
			g.P("func (x *", messagerName, ") Find", index.Name(), "Map() ", mapType, " {")
			g.P("return x.", indexContainerName)
			g.P("}")
			g.P()

			var keyType any
			var keyName string
			if len(index.ColFields) == 1 {
				// single-column index
				field := index.ColFields[0] // just take first field
				keyType = helper.ParseGoType(gen, field.FD)
				keyName = helper.ParseIndexFieldNameAsFuncParam(gen, field.FD)
			} else {
				// multi-column index
				keyType = fmt.Sprintf("%s_Index_%sKey", messagerName, index.Name())
				keyName = "key"
			}

			g.P("// Find", index.Name(), " returns a slice of all values of the given key.")
			g.P("func (x *", messagerName, ") Find", index.Name(), "(", keyName, " ", keyType, ") []*", helper.FindMessageGoIdent(gen, index.MD), " {")
			g.P("return x.", indexContainerName, "[", keyName, "]")
			g.P("}")
			g.P()

			g.P("// FindFirst", index.Name(), " returns the first value of the given key,")
			g.P("// or nil if the key correspond to no value.")
			g.P("func (x *", messagerName, ") FindFirst", index.Name(), "(", keyName, " ", keyType, ") *", helper.FindMessageGoIdent(gen, index.MD), " {")
			g.P("val := x.", indexContainerName, "[", keyName, "]")
			g.P("if len(val) > 0 {")
			g.P("return val[0]")
			g.P("}")
			g.P("return nil")
			g.P("}")
			g.P()
		}
	}
}
