package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
)

func genHppIndexFinders(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P(helper.Indent(1), "// Index accessers.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			if len(index.ColFields) == 1 {
				// single-column index
				field := index.ColFields[0] // just take first field
				g.P(helper.Indent(1), "// Index: ", index.Index)
				g.P(" public:")
				vectorType := fmt.Sprintf("Index_%sVector", index.Name())
				mapType := fmt.Sprintf("Index_%sMap", index.Name())
				g.P(helper.Indent(1), "using ", vectorType, " = std::vector<const ", helper.ParseCppClassType(index.MD), "*>;")
				keyType := helper.ParseCppType(field.FD)
				g.P(helper.Indent(1), "using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, ">;")
				g.P(helper.Indent(1), "// Finds the index (", index.Index, ") to value (", vectorType, ") hash map.")
				g.P(helper.Indent(1), "// One key may correspond to multiple values, which are contained by a vector.")
				g.P(helper.Indent(1), "const ", mapType, "& Find", index.Name(), "() const;")
				g.P(helper.Indent(1), "// Finds a vector of all values of the given key.")
				g.P(helper.Indent(1), "const ", vectorType, "* Find", index.Name(), "(", helper.ToConstRefType(keyType), " ", helper.ParseIndexFieldNameAsFuncParam(field.FD), ") const;")
				g.P(helper.Indent(1), "// Finds the first value of the given key.")
				g.P(helper.Indent(1), "const ", helper.ParseCppClassType(index.MD), "* FindFirst", index.Name(), "(", helper.ToConstRefType(keyType), " ", helper.ParseIndexFieldNameAsFuncParam(field.FD), ") const;")
				g.P()

				g.P(" private:")
				indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
				g.P(helper.Indent(1), mapType, " ", indexContainerName, ";")
				g.P()
			} else {
				// multi-column index
				g.P(helper.Indent(1), "// Index: ", index.Index)
				g.P(" public:")
				keyType := fmt.Sprintf("Index_%sKey", index.Name())
				keyHasherType := fmt.Sprintf("Index_%sKeyHasher", index.Name())
				vectorType := fmt.Sprintf("Index_%sVector", index.Name())
				mapType := fmt.Sprintf("Index_%sMap", index.Name())

				// generate key struct
				g.P(helper.Indent(1), "struct ", keyType, " {")
				equality := ""
				for i, field := range index.ColFields {
					g.P(helper.Indent(2), helper.ParseCppType(field.FD), " ", helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD), ";")
					equality += helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD) + " == other." + helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD)
					if i != len(index.ColFields)-1 {
						equality += " && "
					}
				}
				g.P(helper.Indent(2), "bool operator==(const ", keyType, "& other) const {")
				g.P(helper.Indent(3), "return ", equality, ";")
				g.P(helper.Indent(2), "}")
				g.P(helper.Indent(1), "};")

				// generate key hasher struct
				g.P(helper.Indent(1), "struct ", keyHasherType, " {")
				combinedKeys := ""
				for i, field := range index.ColFields {
					key := "key." + helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD)
					combinedKeys += key
					if i != len(index.ColFields)-1 {
						combinedKeys += ", "
					}
				}
				g.P(helper.Indent(2), "std::size_t operator()(const ", keyType, "& key) const {")
				g.P(helper.Indent(3), "return util::SugaredHashCombine(", combinedKeys, ");")
				g.P(helper.Indent(2), "}")
				g.P(helper.Indent(1), "};")

				g.P(helper.Indent(1), "using ", vectorType, " = std::vector<const ", helper.ParseCppClassType(index.MD), "*>;")
				g.P(helper.Indent(1), "using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, ", ", keyHasherType, ">;")
				g.P(helper.Indent(1), "// Finds the index (", index.Index, ") to value (", vectorType, ") hash map.")
				g.P(helper.Indent(1), "// One key may correspond to multiple values, which are contained by a vector.")
				g.P(helper.Indent(1), "const ", mapType, "& Find", index.Name(), "() const;")
				g.P(helper.Indent(1), "// Finds a vector of all values of the given key.")
				g.P(helper.Indent(1), "const ", vectorType, "* Find", index.Name(), "(const ", keyType, "& key) const;")
				g.P(helper.Indent(1), "// Finds the first value of the given key.")
				g.P(helper.Indent(1), "const ", helper.ParseCppClassType(index.MD), "* FindFirst", index.Name(), "(const ", keyType, "& key) const;")
				g.P()

				g.P(" private:")
				indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
				g.P(helper.Indent(1), mapType, " ", indexContainerName, ";")
				g.P()
			}
		}
	}
}

func genCppIndexLoader(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P(helper.Indent(1), "// Index init.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
			g.P(helper.Indent(1), indexContainerName, ".clear();")
		}
	}
	parentDataName := "data_"
	depth := 1
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			genOneCppIndexLoader(g, depth, index, parentDataName)
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
	genIndexSorter(g, descriptor)
}

func genOneCppIndexLoader(g *protogen.GeneratedFile, depth int, index *index.LevelIndex, parentDataName string) {
	indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
	g.P(helper.Indent(depth), "{")
	g.P(helper.Indent(depth+1), "// Index: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			fieldName := ""
			for _, leveledFd := range field.LeveledFDList {
				fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
			}
			g.P(helper.Indent(depth+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
			key := itemName
			if field.FD.Enum() != nil {
				key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
			}
			g.P(helper.Indent(depth+2), indexContainerName, "[", key, "].push_back(&", parentDataName, ");")
			g.P(helper.Indent(depth+1), "}")
		} else {
			fieldName := ""
			for _, leveledFd := range field.LeveledFDList {
				fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
			}
			key := parentDataName + fieldName
			g.P(helper.Indent(depth+1), indexContainerName, "[", key, "].push_back(&", parentDataName, ");")
		}
	} else {
		// multi-column index
		generateOneCppMulticolumnIndex(g, depth, index, parentDataName, nil)
	}
	g.P(helper.Indent(depth), "}")
}

func genIndexSorter(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
			if len(index.SortedColFields) != 0 {
				g.P(helper.Indent(1), "// Index(sort): ", index.Index)
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

func generateOneCppMulticolumnIndex(g *protogen.GeneratedFile, depth int, index *index.LevelIndex, parentDataName string, keys []string) []string {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		var keyParams string
		for i, key := range keys {
			keyParams += key
			if i != len(keys)-1 {
				keyParams += ", "
			}
		}
		keyType := fmt.Sprintf("Index_%sKey", index.Name())
		indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
		g.P(helper.Indent(depth+1), keyType, " key{", keyParams, "};")
		g.P(helper.Indent(depth+1), indexContainerName, "[key].push_back(&", parentDataName, ");")
		return keys
	}
	field := index.ColFields[cursor]
	if field.FD.IsList() {
		itemName := fmt.Sprintf("index_item%d", cursor)
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
		}
		g.P(helper.Indent(depth+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
		key := itemName
		if field.FD.Enum() != nil {
			key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
		}
		keys = append(keys, key)
		keys = generateOneCppMulticolumnIndex(g, depth+1, index, parentDataName, keys)
		g.P(helper.Indent(depth+1), "}")
	} else {
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
		}
		key := parentDataName + fieldName
		keys = append(keys, key)
		keys = generateOneCppMulticolumnIndex(g, depth, index, parentDataName, keys)
	}
	return keys
}

func genCppIndexFinders(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, messagerName string) {
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			vectorType := "Index_" + index.Name() + "Vector"
			mapType := "Index_" + index.Name() + "Map"
			indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"

			g.P("// Index: ", index.Index)
			g.P("const ", messagerName, "::", mapType, "& ", messagerName, "::Find", index.Name(), "() const { return ", indexContainerName, " ;}")
			g.P()

			var keyType, keyName string
			if len(index.ColFields) == 1 {
				// single-column index
				field := index.ColFields[0] // just take first field
				keyType = helper.ParseCppType(field.FD)
				keyName = helper.ParseIndexFieldNameAsFuncParam(field.FD)
			} else {
				// multi-column index
				keyType = fmt.Sprintf("const Index_%sKey&", index.Name())
				keyName = "key"
			}

			g.P("const ", messagerName, "::", vectorType, "* ", messagerName, "::Find", index.Name(), "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
			g.P(helper.Indent(1), "auto iter = ", indexContainerName, ".find(", keyName, ");")
			g.P(helper.Indent(1), "if (iter == ", indexContainerName, ".end()) {")
			g.P(helper.Indent(2), "return nullptr;")
			g.P(helper.Indent(1), "}")
			g.P(helper.Indent(1), "return &iter->second;")
			g.P("}")
			g.P()

			g.P("const ", helper.ParseCppClassType(index.MD), "* ", messagerName, "::FindFirst", index.Name(), "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
			g.P(helper.Indent(1), "auto conf = Find", index.Name(), "(", keyName, ");")
			g.P(helper.Indent(1), "if (conf == nullptr || conf->empty()) {")
			g.P(helper.Indent(2), "return nullptr;")
			g.P(helper.Indent(1), "}")
			g.P(helper.Indent(1), "return conf->front();")
			g.P("}")
			g.P()
		}
	}
}
