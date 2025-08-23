package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
)

func genHppOrderedIndexFinders(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P(helper.Indent(1), "// OrderedIndex accessers.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			// single-column index
			field := index.ColFields[0] // just take first field
			g.P(helper.Indent(1), "// OrderedIndex: ", index.Index)
			g.P(" public:")
			vectorType := fmt.Sprintf("OrderedIndex_%sVector", index.Name())
			mapType := fmt.Sprintf("OrderedIndex_%sMap", index.Name())
			g.P(helper.Indent(1), "using ", vectorType, " = std::vector<const ", helper.ParseCppClassType(index.MD), "*>;")
			keyType := helper.ParseOrderedMapKeyType(field.FD)
			g.P(helper.Indent(1), "using ", mapType, " = std::map<", keyType, ", ", vectorType, ">;")
			g.P(helper.Indent(1), "const ", mapType, "& Search", index.Name(), "() const;")
			g.P(helper.Indent(1), "const ", vectorType, "* Search", index.Name(), "(", helper.ToConstRefType(keyType), " ", helper.ParseIndexFieldNameAsFuncParam(field.FD), ") const;")
			g.P(helper.Indent(1), "const ", helper.ParseCppClassType(index.MD), "* SearchFirst", index.Name(), "(", helper.ToConstRefType(keyType), " ", helper.ParseIndexFieldNameAsFuncParam(field.FD), ") const;")
			g.P()

			g.P(" private:")
			indexContainerName := "ordered_index_" + strcase.ToSnake(index.Name()) + "_map_"
			g.P(helper.Indent(1), mapType, " ", indexContainerName, ";")
			g.P()
		}
	}
}

func genCppOrderedIndexLoader(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P(helper.Indent(1), "// OrderedIndex init.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			indexContainerName := "ordered_index_" + strcase.ToSnake(index.Name()) + "_map_"
			g.P(helper.Indent(1), indexContainerName, ".clear();")
		}
	}
	parentDataName := "data_"
	depth := 1
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			genOneCppOrderedIndexLoader(g, depth, index, parentDataName)
		}
		itemName := fmt.Sprintf("item%d", depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGen() {
			break
		}
		g.P(helper.Indent(depth), "for (auto&& ", itemName, " : ", parentDataName, ".", helper.ParseIndexFieldName(levelMessage.FD), "()) {")
		parentDataName = itemName
		if levelMessage.FD.IsMap() {
			parentDataName = itemName + ".second"
		}
		depth++
	}
	for i := depth - 1; i > 0; i-- {
		g.P(helper.Indent(i), "}")
	}
	genOrderedIndexSorter(g, descriptor)
}

func genOneCppOrderedIndexLoader(g *protogen.GeneratedFile, depth int, index *index.LevelIndex, parentDataName string) {
	indexContainerName := "ordered_index_" + strcase.ToSnake(index.Name()) + "_map_"
	g.P(helper.Indent(depth), "{")
	g.P(helper.Indent(depth+1), "// OrderedIndex: ", index.Index)
	// single-column index
	field := index.ColFields[0] // just take the first field
	if field.FD.IsList() {
		itemName := fmt.Sprintf("item%d", depth)
		fieldName := ""
		suffix := ""
		for i, leveledFd := range field.LeveledFDList {
			fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
			if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
				switch leveledFd.Message().FullName() {
				case "google.protobuf.Timestamp", "google.protobuf.Duration":
					suffix = ".seconds()"
				default:
				}
			}
		}
		g.P(helper.Indent(depth+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
		key := itemName + suffix
		if field.FD.Enum() != nil {
			key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
		}
		g.P(helper.Indent(depth+2), indexContainerName, "[", key, "].push_back(&", parentDataName, ");")
		g.P(helper.Indent(depth+1), "}")
	} else {
		fieldName := ""
		suffix := ""
		for i, leveledFd := range field.LeveledFDList {
			fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
			if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
				switch leveledFd.Message().FullName() {
				case "google.protobuf.Timestamp", "google.protobuf.Duration":
					suffix = ".seconds()"
				default:
				}
			}
		}
		key := parentDataName + fieldName + suffix
		g.P(helper.Indent(depth+1), indexContainerName, "[", key, "].push_back(&", parentDataName, ");")
	}
	g.P(helper.Indent(depth), "}")
}

func genOrderedIndexSorter(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			indexContainerName := "ordered_index_" + strcase.ToSnake(index.Name()) + "_map_"
			if len(index.SortedColFields) != 0 {
				g.P(helper.Indent(1), "// OrderedIndex(sort): ", index.Index)
				g.P(helper.Indent(1), "for (auto&& item : ", indexContainerName, ") {")
				g.P(helper.Indent(2), "std::sort(item.second.begin(), item.second.end(),")
				g.P(helper.Indent(7), "[](const ", helper.ParseCppClassType(index.MD), "* a, const ", helper.ParseCppClassType(index.MD), "* b) {")
				for i, field := range index.SortedColFields {
					fieldName := ""
					for i, leveledFd := range field.LeveledFDList {
						accessOperator := "."
						if i == 0 {
							accessOperator = "->"
						}
						fieldName += accessOperator + helper.ParseIndexFieldName(leveledFd) + "()"
					}
					if i == len(index.SortedColFields)-1 {
						g.P(helper.Indent(8), "return a", fieldName, " < b", fieldName, ";")
					} else {
						g.P(helper.Indent(8), "if (a", fieldName, " != b", fieldName, ") {")
						g.P(helper.Indent(9), "return a", fieldName, " < b", fieldName, ";")
						g.P(helper.Indent(8), "}")
					}
				}
				g.P(helper.Indent(7), "});")
				g.P(helper.Indent(1), "}")
			}
		}
	}
}

func genCppOrderedIndexFinders(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			vectorType := "OrderedIndex_" + index.Name() + "Vector"
			mapType := "OrderedIndex_" + index.Name() + "Map"
			indexContainerName := "ordered_index_" + strcase.ToSnake(index.Name()) + "_map_"

			g.P("// OrderedIndex: ", index.Index)
			g.P("const ", messagerName, "::", mapType, "& ", messagerName, "::Search", index.Name(), "() const { return ", indexContainerName, " ;}")
			g.P()

			// single-column index
			field := index.ColFields[0] // just take first field
			keyType := helper.ParseOrderedMapKeyType(field.FD)
			keyName := helper.ParseIndexFieldNameAsFuncParam(field.FD)

			g.P("const ", messagerName, "::", vectorType, "* ", messagerName, "::Search", index.Name(), "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
			g.P(helper.Indent(1), "auto iter = ", indexContainerName, ".find(", keyName, ");")
			g.P(helper.Indent(1), "if (iter == ", indexContainerName, ".end()) {")
			g.P(helper.Indent(2), "return nullptr;")
			g.P(helper.Indent(1), "}")
			g.P(helper.Indent(1), "return &iter->second;")
			g.P("}")
			g.P()

			g.P("const ", helper.ParseCppClassType(index.MD), "* ", messagerName, "::SearchFirst", index.Name(), "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
			g.P(helper.Indent(1), "auto conf = Search", index.Name(), "(", keyName, ");")
			g.P(helper.Indent(1), "if (conf == nullptr || conf->empty()) {")
			g.P(helper.Indent(2), "return nullptr;")
			g.P(helper.Indent(1), "}")
			g.P(helper.Indent(1), "return conf->front();")
			g.P("}")
			g.P()
		}
	}
}
