package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
)

func genIndexTypeDef(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	g.P(helper.Indent(2), "// Index types.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			mapType := fmt.Sprintf("Index_%sMap", index.Name())
			if len(index.ColFields) == 1 {
				// single-column index
				field := index.ColFields[0] // just take first field
				g.P(helper.Indent(2), "// Index: ", index.Index)
				keyType := helper.ParseCsharpType(field.FD)
				g.P(helper.Indent(2), "public class ", mapType, " : Dictionary<", keyType, ", List<", helper.ParseCsharpClassType(index.MD), ">> { }")
			} else {
				// multi-column index
				g.P(helper.Indent(2), "// Index: ", index.Index)
				keyType := fmt.Sprintf("%s_Index_%sKey", messagerName, index.Name())

				// generate key struct
				g.P(helper.Indent(2), "public struct ", keyType)
				g.P(helper.Indent(2), "{")
				for _, field := range index.ColFields {
					g.P(helper.Indent(3), "public ", helper.ParseCsharpType(field.FD), " ", helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD), ";")
				}
				g.P()
				var keyParams string
				for i, field := range index.ColFields {
					keyParams += helper.ParseCsharpType(field.FD) + " " + strcase.ToLowerCamel(helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD))
					if i != len(index.ColFields)-1 {
						keyParams += ", "
					}
				}
				g.P(helper.Indent(3), "public ", keyType, "(", keyParams, ")")
				g.P(helper.Indent(3), "{")
				for _, field := range index.ColFields {
					g.P(helper.Indent(4), helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD), " = ", strcase.ToLowerCamel(helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD)), ";")
				}
				g.P(helper.Indent(3), "}")
				g.P(helper.Indent(2), "}")
				g.P()
				g.P(helper.Indent(2), "public class ", mapType, " : Dictionary<", keyType, ", List<", helper.ParseCsharpClassType(index.MD), ">> { }")
			}
			g.P()
			indexContainerName := "Index" + strcase.ToCamel(index.Name()) + "Map"
			g.P(helper.Indent(2), "private ", mapType, " ", indexContainerName, " = new ", mapType, "();")
			g.P()
		}
	}
}

func genIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	g.P(helper.Indent(3), "// Index init.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			indexContainerName := "Index" + strcase.ToCamel(index.Name()) + "Map"
			g.P(helper.Indent(3), indexContainerName, ".Clear();")
		}
	}
	parentDataName := "Data_"
	depth := 1
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			genOneIndexLoader(gen, g, depth, index, parentDataName, messagerName)
		}
		itemName := fmt.Sprintf("item%d", depth)
		if levelMessage.FD == nil {
			break
		}
		g.P(helper.Indent(depth+2), "foreach (var ", itemName, " in ", parentDataName, ".", helper.ParseIndexFieldName(levelMessage.FD), ")")
		g.P(helper.Indent(depth+2), "{")
		parentDataName = itemName
		if levelMessage.FD.IsMap() {
			parentDataName = itemName + ".Value"
		}
		defer g.P(helper.Indent(depth+2), "}")
		depth++
	}
}

func genOneIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, depth int, index *index.LevelIndex, parentDataName string, messagerName string) {
	indexContainerName := "Index" + strcase.ToCamel(index.Name()) + "Map"
	g.P(helper.Indent(depth+2), "{")
	g.P(helper.Indent(depth+3), "// Index: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			fieldName := ""
			for i, leveledFd := range field.LeveledFDList {
				if i != 0 {
					fieldName += "?"
				}
				fieldName += "." + helper.ParseIndexFieldName(leveledFd)
			}
			g.P(helper.Indent(depth+3), "foreach (var ", itemName, " in ", parentDataName, fieldName, " ?? Enumerable.Empty<", helper.ParseCsharpType(field.FD), ">())")
			g.P(helper.Indent(depth+3), "{")
			g.P(helper.Indent(depth+4), "var key = ", itemName, ";")
			g.P(helper.Indent(depth+4), "if (!", indexContainerName, ".ContainsKey(key))")
			g.P(helper.Indent(depth+4), "{")
			g.P(helper.Indent(depth+5), indexContainerName, "[key] = new List<", helper.ParseCsharpClassType(index.MD), ">();")
			g.P(helper.Indent(depth+4), "}")
			g.P(helper.Indent(depth+4), indexContainerName, "[key].Add(", parentDataName, ");")
			g.P(helper.Indent(depth+3), "}")
		} else {
			fieldName := ""
			for i, leveledFd := range field.LeveledFDList {
				if i != 0 {
					fieldName += "?"
				}
				fieldName += "." + helper.ParseIndexFieldName(leveledFd)
			}
			key := parentDataName + fieldName
			if len(field.LeveledFDList) > 1 {
				key += " ?? " + helper.GetTypeEmptyValue(field.FD)
			}
			g.P(helper.Indent(depth+3), "var key = ", key, ";")
			g.P(helper.Indent(depth+3), "if (!", indexContainerName, ".ContainsKey(key))")
			g.P(helper.Indent(depth+3), "{")
			g.P(helper.Indent(depth+4), indexContainerName, "[key] = new List<", helper.ParseCsharpClassType(index.MD), ">();")
			g.P(helper.Indent(depth+3), "}")
			g.P(helper.Indent(depth+3), indexContainerName, "[key].Add(", parentDataName, ");")
		}
	} else {
		// multi-column index
		generateOneMulticolumnIndex(gen, g, depth, index, parentDataName, messagerName, nil)
	}
	if len(index.KeyFields) != 0 {
		g.P(helper.Indent(depth+3), "foreach (var item in ", indexContainerName, ")")
		g.P(helper.Indent(depth+3), "{")
		g.P(helper.Indent(depth+4), "item.Value.Sort((a, b) =>")
		g.P(helper.Indent(depth+4), "{")
		for i, field := range index.KeyFields {
			fieldName := ""
			for i, leveledFd := range field.LeveledFDList {
				if i != 0 {
					fieldName += "?"
				}
				fieldName += "." + helper.ParseIndexFieldName(leveledFd)
			}
			if len(field.LeveledFDList) > 1 {
				fieldName += " ?? " + helper.GetTypeEmptyValue(field.FD)
			}
			if i == len(index.KeyFields)-1 {
				if len(field.LeveledFDList) > 1 {
					g.P(helper.Indent(depth+5), "return (a", fieldName, ").CompareTo(b", fieldName, ");")
				} else {
					g.P(helper.Indent(depth+5), "return a", fieldName, ".CompareTo(b", fieldName, ");")
				}
			} else {
				g.P(helper.Indent(depth+5), "if (a", fieldName, " != b", fieldName, ")")
				g.P(helper.Indent(depth+5), "{")
				if len(field.LeveledFDList) > 1 {
					g.P(helper.Indent(depth+6), "return (a", fieldName, ").CompareTo(b", fieldName, ");")
				} else {
					g.P(helper.Indent(depth+6), "return a", fieldName, ".CompareTo(b", fieldName, ");")
				}
				g.P(helper.Indent(depth+5), "}")
			}
		}
		g.P(helper.Indent(depth+4), "});")
		g.P(helper.Indent(depth+3), "}")
	}
	g.P(helper.Indent(depth+2), "}")
}

func generateOneMulticolumnIndex(gen *protogen.Plugin, g *protogen.GeneratedFile, depth int, index *index.LevelIndex, parentDataName string, messagerName string, keys []string) []string {
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
		indexContainerName := "Index" + strcase.ToCamel(index.Name()) + "Map"
		g.P(helper.Indent(depth+3), "var key = new ", keyType, "(", keyParams, ");")
		g.P(helper.Indent(depth+3), "if (!", indexContainerName, ".ContainsKey(key))")
		g.P(helper.Indent(depth+3), "{")
		g.P(helper.Indent(depth+4), indexContainerName, "[key] = new List<", helper.ParseCsharpClassType(index.MD), ">();")
		g.P(helper.Indent(depth+3), "}")
		g.P(helper.Indent(depth+3), indexContainerName, "[key].Add(", parentDataName, ");")
		return keys
	}
	field := index.ColFields[cursor]
	if field.FD.IsList() {
		itemName := fmt.Sprintf("indexItem%d", cursor)
		fieldName := ""
		for i, leveledFd := range field.LeveledFDList {
			if i != 0 {
				fieldName += "?"
			}
			fieldName += "." + helper.ParseIndexFieldName(leveledFd)
		}
		g.P(helper.Indent(depth+3), "foreach (var ", itemName, " in ", parentDataName, fieldName, " ?? Enumerable.Empty<", helper.ParseCsharpType(field.FD), ">())")
		g.P(helper.Indent(depth+3), "{")
		keys = append(keys, itemName)
		keys = generateOneMulticolumnIndex(gen, g, depth+1, index, parentDataName, messagerName, keys)
		g.P(helper.Indent(depth+3), "}")
	} else {
		fieldName := ""
		for i, leveledFd := range field.LeveledFDList {
			if i != 0 {
				fieldName += "?"
			}
			fieldName += "." + helper.ParseIndexFieldName(leveledFd)
		}
		key := parentDataName + fieldName
		if len(field.LeveledFDList) > 1 {
			key += " ?? " + helper.GetTypeEmptyValue(field.FD)
		}
		keys = append(keys, key)
		keys = generateOneMulticolumnIndex(gen, g, depth, index, parentDataName, messagerName, keys)
	}
	return keys
}

func genIndexFinders(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			mapType := fmt.Sprintf("Index_%sMap", index.Name())
			indexContainerName := "Index" + strcase.ToCamel(index.Name()) + "Map"
			g.P()
			g.P(helper.Indent(2), "// Index: ", index.Index)
			g.P(helper.Indent(2), "public ref readonly ", mapType, " Get", index.Name(), "Map() => ref ", indexContainerName, ";")
			g.P()

			var keyType, keyName string
			if len(index.ColFields) == 1 {
				// single-column index
				field := index.ColFields[0] // just take first field
				keyType = helper.ParseCsharpType(field.FD)
				keyName = helper.ParseIndexFieldNameAsFuncParam(field.FD)
			} else {
				// multi-column index
				keyType = fmt.Sprintf("%s_Index_%sKey", messagerName, index.Name())
				keyName = "key"
			}

			g.P(helper.Indent(2), "public List<", helper.ParseCsharpClassType(index.MD), ">? Get", index.Name(), "(", keyType, " ", keyName, ") => ", indexContainerName, ".TryGetValue(", keyName, ", out var value) ? value : null;")
			g.P()

			g.P(helper.Indent(2), "public ", helper.ParseCsharpClassType(index.MD), "? GetFirst", index.Name(), "(", keyType, " ", keyName, ") => Get", index.Name(), "(", keyName, ")?.FirstOrDefault();")
		}
	}
}
