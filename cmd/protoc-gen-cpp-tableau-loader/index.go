package main

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
)

func genHppIndexFinders(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P("  // Index accessers.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			if len(index.ColFields) == 1 {
				// single-column index
				field := index.ColFields[0] // just take first field
				g.P("  // Index: ", index.Index)
				g.P(" public:")
				vectorType := fmt.Sprintf("Index_%sVector", index.Name())
				mapType := fmt.Sprintf("Index_%sMap", index.Name())
				g.P("  using ", vectorType, " = std::vector<const ", helper.ParseCppClassType(index.MD), "*>;")
				keyType := helper.ParseCppType(field.FD)
				g.P("  using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, ">;")
				g.P("  const ", mapType, "& Find", index.Name(), "() const;")
				g.P("  const ", vectorType, "* Find", index.Name(), "(", helper.ToConstRefType(keyType), " ", helper.ParseIndexFieldNameAsFuncParam(field.FD), ") const;")
				g.P("  const ", helper.ParseCppClassType(index.MD), "* FindFirst", index.Name(), "(", helper.ToConstRefType(keyType), " ", helper.ParseIndexFieldNameAsFuncParam(field.FD), ") const;")
				g.P()

				g.P(" private:")
				indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
				g.P("  ", mapType, " ", indexContainerName, ";")
				g.P()
			} else {
				// multi-column index
				g.P("  // Index: ", index.Index)
				g.P(" public:")
				keyType := fmt.Sprintf("Index_%sKey", index.Name())
				keyHasherType := fmt.Sprintf("Index_%sKeyHasher", index.Name())
				vectorType := fmt.Sprintf("Index_%sVector", index.Name())
				mapType := fmt.Sprintf("Index_%sMap", index.Name())

				// generate key struct
				g.P("  struct ", keyType, " {")
				equality := ""
				for i, field := range index.ColFields {
					g.P("    ", helper.ParseCppType(field.FD), " ", helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD), ";")
					equality += helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD) + " == other." + helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD)
					if i != len(index.ColFields)-1 {
						equality += " && "
					}
				}
				g.P("    bool operator==(const ", keyType, "& other) const {")
				g.P("      return ", equality, ";")
				g.P("    }")
				g.P("  };")

				// generate key hasher struct
				g.P("  struct ", keyHasherType, " {")
				combinedKeys := ""
				for i, field := range index.ColFields {
					key := "key." + helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD)
					combinedKeys += key
					if i != len(index.ColFields)-1 {
						combinedKeys += ", "
					}
				}
				g.P("    std::size_t operator()(const ", keyType, "& key) const {")
				g.P("      return util::SugaredHashCombine(", combinedKeys, ");")
				g.P("    }")
				g.P("  };")

				g.P("  using ", vectorType, " = std::vector<const ", helper.ParseCppClassType(index.MD), "*>;")
				g.P("  using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, ", ", keyHasherType, ">;")
				g.P("  const ", mapType, "& Find", index.Name(), "() const;")
				g.P("  const ", vectorType, "* Find", index.Name(), "(const ", keyType, "& key) const;")
				g.P("  const ", helper.ParseCppClassType(index.MD), "* FindFirst", index.Name(), "(const ", keyType, "& key) const;")
				g.P()

				g.P(" private:")
				indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
				g.P("  ", mapType, " ", indexContainerName, ";")
				g.P()
			}
		}
	}
}

func genCppIndexLoader(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor) {
	g.P("  // Index init.")
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
			g.P("  ", indexContainerName, ".clear();")
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
		g.P(strings.Repeat("  ", depth), "for (auto&& "+itemName+" : "+parentDataName+"."+helper.ParseIndexFieldName(levelMessage.FD)+"()) {")
		parentDataName = itemName
		if levelMessage.FD.IsMap() {
			parentDataName = itemName + ".second"
		}
		defer g.P(strings.Repeat("  ", depth), "}")
		depth++
	}
}

func genOneCppIndexLoader(g *protogen.GeneratedFile, depth int, index *index.LevelIndex, parentDataName string) {
	indexContainerName := "index_" + strcase.ToSnake(index.Name()) + "_map_"
	g.P(strings.Repeat("  ", depth), "{")
	g.P(strings.Repeat("  ", depth+1), "// Index: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			fieldName := ""
			for _, leveledFd := range field.LeveledFDList {
				fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
			}
			g.P(strings.Repeat("  ", depth+1), "for (auto&& "+itemName+" : "+parentDataName+fieldName+") {")
			key := itemName
			if field.FD.Enum() != nil {
				key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
			}
			g.P(strings.Repeat("  ", depth+2), indexContainerName, "["+key+"].push_back(&"+parentDataName+");")
			g.P(strings.Repeat("  ", depth+1), "}")
		} else {
			fieldName := ""
			for _, leveledFd := range field.LeveledFDList {
				fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
			}
			key := parentDataName + fieldName
			g.P(strings.Repeat("  ", depth+1), indexContainerName, "["+key+"].push_back(&"+parentDataName+");")
		}
	} else {
		// multi-column index
		generateOneCppMulticolumnIndex(g, depth, index, parentDataName, nil)
	}
	if len(index.KeyFields) != 0 {
		g.P(strings.Repeat("  ", depth+1), "for (auto&& item : ", indexContainerName, ") {")
		g.P(strings.Repeat("  ", depth+2), "std::sort(item.second.begin(), item.second.end(),")
		g.P(strings.Repeat("  ", depth+7), "[](const ", helper.ParseCppClassType(index.MD), "* a, const ", helper.ParseCppClassType(index.MD), "* b) {")
		for i, field := range index.KeyFields {
			fieldName := ""
			for i, leveledFd := range field.LeveledFDList {
				accessOperator := "."
				if i == 0 {
					accessOperator = "->"
				}
				fieldName += accessOperator + helper.ParseIndexFieldName(leveledFd) + "()"
			}
			if i == len(index.KeyFields)-1 {
				g.P(strings.Repeat("  ", depth+8), "return a", fieldName, " < b", fieldName, ";")
			} else {
				g.P(strings.Repeat("  ", depth+8), "if (a", fieldName, " != b", fieldName, ") {")
				g.P(strings.Repeat("  ", depth+9), "return a", fieldName, " < b", fieldName, ";")
				g.P(strings.Repeat("  ", depth+8), "}")
			}
		}
		g.P(strings.Repeat("  ", depth+7), "});")
		g.P(strings.Repeat("  ", depth+1), "}")
	}
	g.P(strings.Repeat("  ", depth), "}")
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
		g.P(strings.Repeat("  ", depth+1), keyType, " key{", keyParams, "};")
		g.P(strings.Repeat("  ", depth+1), indexContainerName, "[key].push_back(&"+parentDataName+");")
		return keys
	}
	field := index.ColFields[cursor]
	if field.FD.IsList() {
		itemName := fmt.Sprintf("index_item%d", cursor)
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
		}
		g.P(strings.Repeat("  ", depth+1), "for (auto&& "+itemName+" : "+parentDataName+fieldName+") {")
		key := itemName
		if field.FD.Enum() != nil {
			key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
		}
		keys = append(keys, key)
		keys = generateOneCppMulticolumnIndex(g, depth+1, index, parentDataName, keys)
		g.P(strings.Repeat("  ", depth+1), "}")
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
			g.P("const ", messagerName, "::", mapType, "& "+messagerName+"::Find", index.Name(), "() const { return "+indexContainerName+" ;}")
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

			g.P("const ", messagerName, "::", vectorType, "* "+messagerName+"::Find", index.Name(), "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
			g.P("  auto iter = ", indexContainerName, ".find(", keyName, ");")
			g.P("  if (iter == ", indexContainerName, ".end()) {")
			g.P("    return nullptr;")
			g.P("  }")
			g.P("  return &iter->second;")
			g.P("}")
			g.P()

			g.P("const ", helper.ParseCppClassType(index.MD), "* "+messagerName+"::FindFirst", index.Name(), "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
			g.P("  auto conf = Find", index.Name(), "(", keyName, ");")
			g.P("  if (conf == nullptr || conf->size() == 0) {")
			g.P("    return nullptr;")
			g.P("  }")
			g.P("  return (*conf)[0];")
			g.P("}")
			g.P()
		}
	}
}
