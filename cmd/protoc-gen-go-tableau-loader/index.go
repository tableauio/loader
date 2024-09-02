package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
)

func genIndexTypeDef(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptors []*index.IndexDescriptor, messagerName string) {
	g.P("// Index types.")
	for _, descriptor := range descriptors {
		if len(descriptor.Fields) == 1 {
			// single-column index
			field := descriptor.Fields[0] // just take first field
			g.P("// Index: ", descriptor.Index)
			mapType := fmt.Sprintf("%s_Index_%sMap", messagerName, descriptor.Name)
			keyType := helper.ParseGoType(gen, field.FD)
			g.P("type ", mapType, " = map[", keyType, "][]*", helper.FindMessageGoIdent(gen, descriptor.MD))
		} else {
			// multi-column index
			g.P("// Index: ", descriptor.Index)
			keyType := fmt.Sprintf("%s_Index_%sKey", messagerName, descriptor.Name)
			mapType := fmt.Sprintf("%s_Index_%sMap", messagerName, descriptor.Name)

			// generate key struct
			// KeyType must be comparable, refer https://go.dev/blog/maps
			g.P("type ", keyType, " struct {")
			for _, field := range descriptor.Fields {
				g.P(helper.ParseIndexFieldNameAsKeyStructFieldName(gen, field.FD), " ", helper.ParseGoType(gen, field.FD))
			}
			g.P("}")
			g.P("type ", mapType, " = map[", keyType, "][]*", helper.FindMessageGoIdent(gen, descriptor.MD))
		}
		g.P()
	}
}

func genIndexField(g *protogen.GeneratedFile, descriptors []*index.IndexDescriptor, messagerName string) {
	for _, descriptor := range descriptors {
		indexContainerName := "index" + strcase.ToCamel(descriptor.Name) + "Map"
		mapType := fmt.Sprintf("%s_Index_%sMap", messagerName, descriptor.Name)
		g.P(indexContainerName, " ", mapType)
	}
}

func genIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptors []*index.IndexDescriptor, messagerName string) {
	g.P("  // Index init.")
	for _, descriptor := range descriptors {
		parentDataName := "x.data"
		g.P("  // Index: ", descriptor.Index)
		genOneIndexLoader(gen, g, 1, descriptor, parentDataName, descriptor.LevelMessage, messagerName)
	}
}

func genOneIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, depth int, descriptor *index.IndexDescriptor,
	parentDataName string, levelMessage *index.LevelMessage, messagerName string) {
	if levelMessage == nil {
		return
	}
	indexContainerName := "index" + strcase.ToCamel(descriptor.Name) + "Map"
	if depth == 1 {
		g.P("x.", indexContainerName, " = make(", fmt.Sprintf("%s_Index_%sMap", messagerName, descriptor.Name), ")")
	}
	if levelMessage.NextLevel != nil {
		itemName := fmt.Sprintf("item%d", depth)
		g.P("for _, ", itemName, " := range "+parentDataName+".Get"+helper.ParseIndexFieldName(gen, levelMessage.FD)+"() {")
		parentDataName = itemName
		genOneIndexLoader(gen, g, depth+1, descriptor, parentDataName, levelMessage.NextLevel, messagerName)
		g.P("}")
	} else {
		if len(levelMessage.Fields) == 1 {
			// single-column index
			field := levelMessage.Fields[0] // just take the first field
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
			var keys []string
			generateOneMulticolumnIndex(gen, g, depth, descriptor, parentDataName, keys)
		}
	}
}

func generateOneMulticolumnIndex(gen *protogen.Plugin, g *protogen.GeneratedFile,
	depth int, descriptor *index.IndexDescriptor, parentDataName string, keys []string) []string {
	cursor := len(keys)
	if cursor >= len(descriptor.Fields) {
		var keyParams string
		for i, key := range keys {
			keyParams += key
			if i != len(keys)-1 {
				keyParams += ", "
			}
		}
		keyType := fmt.Sprintf("%s_Index_%sKey", descriptor.LevelMessage.FD.ContainingMessage().Name(), descriptor.Name)
		indexContainerName := "index" + strcase.ToCamel(descriptor.Name) + "Map"
		g.P("key := ", keyType, " {", keyParams, "}")
		g.P("x.", indexContainerName, "[key] = append(x.", indexContainerName, "[key], ", parentDataName, ")")
		return keys
	}
	field := descriptor.Fields[cursor]
	if field.FD.IsList() {
		itemName := fmt.Sprintf("indexItem%d", cursor)
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
		}
		g.P("for _, " + itemName + " := range " + parentDataName + fieldName + " {")
		keys = append(keys, itemName)
		keys = generateOneMulticolumnIndex(gen, g, depth+1, descriptor, parentDataName, keys)
		g.P("}")
	} else {
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += ".Get" + helper.ParseIndexFieldName(gen, leveledFd) + "()"
		}
		key := parentDataName + fieldName
		keys = append(keys, key)
		keys = generateOneMulticolumnIndex(gen, g, depth, descriptor, parentDataName, keys)
	}
	return keys
}

func genIndexFinders(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptors []*index.IndexDescriptor, messagerName string) {
	for _, descriptor := range descriptors {
		// sliceType := "[]*" + descriptor.GoIdent
		mapType := fmt.Sprintf("%s_Index_%sMap", messagerName, descriptor.Name)
		indexContainerName := "index" + strcase.ToCamel(descriptor.Name) + "Map"

		g.P("// Index: ", descriptor.Index)
		g.P()

		g.P("// Find", descriptor.Name, "Map returns the index(", descriptor.Index, ") to value(", helper.FindMessageGoIdent(gen, descriptor.MD), ") map.")
		g.P("// One key may correspond to multiple values, which are contained by a slice.")
		g.P("func (x *", messagerName, ") Find", descriptor.Name, "Map() ", mapType, " {")
		g.P("return x.", indexContainerName)
		g.P("}")
		g.P()

		var keyType any
		var keyName string
		if len(descriptor.Fields) == 1 {
			// single-column index
			field := descriptor.Fields[0] // just take first field
			keyType = helper.ParseGoType(gen, field.FD)
			keyName = helper.ParseIndexFieldNameAsFuncParam(gen, field.FD)
		} else {
			// multi-column index
			keyType = fmt.Sprintf("%s_Index_%sKey", descriptor.LevelMessage.FD.ContainingMessage().Name(), descriptor.Name)
			keyName = "key"
		}

		g.P("// Find", descriptor.Name, " returns a slice of all values of the given key.")
		g.P("func (x *", messagerName, ") Find", descriptor.Name, "(", keyName, " ", keyType, ") []*", helper.FindMessageGoIdent(gen, descriptor.MD), " {")
		g.P("return x.", indexContainerName, "[", keyName, "]")
		g.P("}")
		g.P()

		g.P("// FindFirst", descriptor.Name, " returns the first value of the given key,")
		g.P("// or nil if the key correspond to no value.")
		g.P("func (x *", messagerName, ") FindFirst", descriptor.Name, "(", keyName, " ", keyType, ") *", helper.FindMessageGoIdent(gen, descriptor.MD), " {")
		g.P("val := x.", indexContainerName, "[", keyName, "]")
		g.P("if len(val) > 0 {")
		g.P("return val[0]")
		g.P("}")
		g.P("return nil")
		g.P("}")
		g.P()

	}
}
