package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
)

func genOrderedIndexTypeDef(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P(helper.Indent(2), "// OrderedIndex types.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			mapType := fmt.Sprintf("OrderedIndex_%sMap", index.Name())
			// single-column index
			field := index.ColFields[0] // just take first field
			g.P(helper.Indent(2), "// OrderedIndex: ", index.Index)
			keyType := helper.ParseOrderedMapKeyType(field.FD)
			g.P(helper.Indent(2), "public class ", mapType, " : SortedDictionary<", keyType, ", List<", helper.ParseCsharpClassType(index.MD), ">>;")
			g.P()
			indexContainerName := "_orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			g.P(helper.Indent(2), "private readonly ", mapType, " ", indexContainerName, " = [];")
			g.P()
		}
	}
}

func genOrderedIndexLoader(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P(helper.Indent(3), "// OrderedIndex init.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			indexContainerName := "_orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			g.P(helper.Indent(3), indexContainerName, ".Clear();")
		}
	}
	parentDataName := "_data"
	depth := 1
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			genOneOrderedIndexLoader(g, depth, index, parentDataName)
		}
		itemName := fmt.Sprintf("item%d", depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGen() {
			break
		}
		g.P(helper.Indent(depth+2), "foreach (var ", itemName, " in ", parentDataName, ".", helper.ParseIndexFieldName(levelMessage.FD), ")")
		g.P(helper.Indent(depth+2), "{")
		parentDataName = itemName
		if levelMessage.FD.IsMap() {
			parentDataName = itemName + ".Value"
		}
		depth++
	}
	for i := depth - 1; i > 0; i-- {
		g.P(helper.Indent(i+2), "}")
	}
	genOrderedIndexSorter(g, descriptor)
}

func genOneOrderedIndexLoader(g *protogen.GeneratedFile, depth int, index *index.LevelIndex, parentDataName string) {
	indexContainerName := "_orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
	g.P(helper.Indent(depth+2), "{")
	g.P(helper.Indent(depth+3), "// OrderedIndex: ", index.Index)
	// single-column index
	field := index.ColFields[0] // just take the first field
	if field.FD.IsList() {
		itemName := fmt.Sprintf("item%d", depth)
		fieldName := ""
		suffix := ""
		for i, leveledFd := range field.LeveledFDList {
			if i != 0 {
				fieldName += "?"
			}
			fieldName += "." + helper.ParseIndexFieldName(leveledFd)
			if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
				switch leveledFd.Message().FullName() {
				case "google.protobuf.Timestamp", "google.protobuf.Duration":
					suffix = ".Seconds"
				default:
				}
			}
		}
		g.P(helper.Indent(depth+3), "foreach (var ", itemName, " in ", parentDataName, fieldName, " ?? Enumerable.Empty<", helper.ParseCsharpType(field.FD), ">())")
		g.P(helper.Indent(depth+3), "{")
		g.P(helper.Indent(depth+4), "var key = ", itemName, suffix, ";")
		g.P(helper.Indent(depth+4), "var list = ", indexContainerName, ".TryGetValue(key, out var existingList) ?")
		g.P(helper.Indent(depth+4), "existingList : ", indexContainerName, "[key] = [];")
		g.P(helper.Indent(depth+4), "list.Add(", parentDataName, ");")
		g.P(helper.Indent(depth+3), "}")
	} else {
		fieldName := ""
		suffix := ""
		for i, leveledFd := range field.LeveledFDList {
			if i != 0 {
				fieldName += "?"
			}
			fieldName += "." + helper.ParseIndexFieldName(leveledFd)
			if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
				switch leveledFd.Message().FullName() {
				case "google.protobuf.Timestamp", "google.protobuf.Duration":
					suffix = ".Seconds"
				default:
				}
			}
		}
		key := parentDataName + fieldName + suffix
		if len(field.LeveledFDList) > 1 {
			key += " ?? " + helper.GetTypeEmptyValue(field.FD)
		}
		g.P(helper.Indent(depth+3), "var key = ", key, ";")
		g.P(helper.Indent(depth+3), "var list = ", indexContainerName, ".TryGetValue(key, out var existingList) ?")
		g.P(helper.Indent(depth+3), "existingList : ", indexContainerName, "[key] = [];")
		g.P(helper.Indent(depth+3), "list.Add(", parentDataName, ");")
	}
	g.P(helper.Indent(depth+2), "}")
}

func genOrderedIndexSorter(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			indexContainerName := "_orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			if len(index.SortedColFields) != 0 {
				g.P(helper.Indent(3), "// Index(sort): ", index.Index)
				g.P(helper.Indent(3), "foreach (var item in ", indexContainerName, ")")
				g.P(helper.Indent(3), "{")
				g.P(helper.Indent(4), "item.Value.Sort((a, b) =>")
				g.P(helper.Indent(4), "{")
				for i, field := range index.SortedColFields {
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
					if i == len(index.SortedColFields)-1 {
						if len(field.LeveledFDList) > 1 {
							g.P(helper.Indent(5), "return (a", fieldName, ").CompareTo(b", fieldName, ");")
						} else {
							g.P(helper.Indent(5), "return a", fieldName, ".CompareTo(b", fieldName, ");")
						}
					} else {
						g.P(helper.Indent(5), "if (a", fieldName, " != b", fieldName, ")")
						g.P(helper.Indent(5), "{")
						if len(field.LeveledFDList) > 1 {
							g.P(helper.Indent(6), "return (a", fieldName, ").CompareTo(b", fieldName, ");")
						} else {
							g.P(helper.Indent(6), "return a", fieldName, ".CompareTo(b", fieldName, ");")
						}
						g.P(helper.Indent(5), "}")
					}
				}
				g.P(helper.Indent(4), "});")
				g.P(helper.Indent(3), "}")
			}
		}
	}
}

func genOrderedIndexFinders(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			mapType := fmt.Sprintf("OrderedIndex_%sMap", index.Name())
			indexContainerName := "_orderedIndex" + strcase.ToCamel(index.Name()) + "Map"
			g.P()
			g.P(helper.Indent(2), "// OrderedIndex: ", index.Index)
			g.P(helper.Indent(2), "public ref readonly ", mapType, " Find", index.Name(), "Map() => ref ", indexContainerName, ";")
			g.P()

			// single-column index
			field := index.ColFields[0] // just take first field
			keyType := helper.ParseOrderedMapKeyType(field.FD)
			keyName := helper.ParseIndexFieldNameAsFuncParam(field.FD)

			g.P(helper.Indent(2), "public List<", helper.ParseCsharpClassType(index.MD), ">? Find", index.Name(), "(", keyType, " ", keyName, ") =>")
			g.P(helper.Indent(3), indexContainerName, ".TryGetValue(", keyName, ", out var value) ? value : null;")
			g.P()

			g.P(helper.Indent(2), "public ", helper.ParseCsharpClassType(index.MD), "? FindFirst", index.Name(), "(", keyType, " ", keyName, ") =>")
			g.P(helper.Indent(3), "Find", index.Name(), "(", keyName, ")?.FirstOrDefault();")
		}
	}
}
