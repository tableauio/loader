package main

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func genHppIndexFinders(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	g.P("  // Index accessers.")
	descriptors := index.ParseIndexDescriptor(gen, md)
	for _, descriptor := range descriptors {
		if len(descriptor.Fields) == 1 {
			// single-column index
			field := descriptor.Fields[0] // just take first field
			g.P("  // Index: ", descriptor.Index)
			g.P(" public:")
			vectorType := fmt.Sprintf("Index_%sVector", descriptor.Name)
			mapType := fmt.Sprintf("Index_%sMap", descriptor.Name)
			g.P("  using ", vectorType, " = std::vector<const ", helper.ParseCppClassType(descriptor.MD), "*>;")
			keyType := helper.ParseCppType(field.FD)
			g.P("  using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, ">;")
			g.P("  const ", mapType, "& Find", descriptor.Name, "() const;")
			g.P("  const ", vectorType, "* Find", descriptor.Name, "(", helper.ToConstRefType(keyType), " ", helper.ParseIndexFieldNameAsFuncParam(field.FD), ") const;")
			g.P("  const ", helper.ParseCppClassType(descriptor.MD), "* FindFirst", descriptor.Name, "(", helper.ToConstRefType(keyType), " ", helper.ParseIndexFieldNameAsFuncParam(field.FD), ") const;")
			g.P()

			g.P(" private:")
			indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"
			g.P("  ", mapType, " ", indexContainerName, ";")
			g.P()
		} else {
			// multi-column index
			g.P("  // Index: ", descriptor.Index)
			g.P(" public:")
			keyType := fmt.Sprintf("Index_%sKey", descriptor.Name)
			keyHasherType := fmt.Sprintf("Index_%sKeyHasher", descriptor.Name)
			vectorType := fmt.Sprintf("Index_%sVector", descriptor.Name)
			mapType := fmt.Sprintf("Index_%sMap", descriptor.Name)

			// generate key struct
			g.P("  struct ", keyType, " {")
			equality := ""
			for i, field := range descriptor.Fields {
				g.P("    ", helper.ParseCppType(field.FD), " ", helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD), ";")
				equality += helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD) + " == other." + helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD)
				if i != len(descriptor.Fields)-1 {
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
			for i, field := range descriptor.Fields {
				key := "key." + helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD)
				combinedKeys += key
				if i != len(descriptor.Fields)-1 {
					combinedKeys += ", "
				}
			}
			g.P("    std::size_t operator()(const ", keyType, "& key) const {")
			g.P("      return util::SugaredHashCombine(", combinedKeys, ");")
			g.P("    }")
			g.P("  };")

			g.P("  using ", vectorType, " = std::vector<const ", helper.ParseCppClassType(descriptor.MD), "*>;")
			g.P("  using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, ", ", keyHasherType, ">;")
			g.P("  const ", mapType, "& Find", descriptor.Name, "() const;")
			g.P("  const ", vectorType, "* Find", descriptor.Name, "(const ", keyType, "& key) const;")
			g.P("  const ", helper.ParseCppClassType(descriptor.MD), "* FindFirst", descriptor.Name, "(const ", keyType, "& key) const;")
			g.P()

			g.P(" private:")
			indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"
			g.P("  ", mapType, " ", indexContainerName, ";")
			g.P()
		}
	}
}

func genCppIndexLoader(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	g.P("  // Index init.")
	descriptors := index.ParseIndexDescriptor(gen, md)
	for _, descriptor := range descriptors {
		parentDataName := "data_"
		g.P("  // Index: ", descriptor.Index)
		genOneCppIndexLoader(1, descriptor, parentDataName, descriptor.LevelMessage, g)
	}
}

func genOneCppIndexLoader(depth int, descriptor *index.IndexDescriptor, parentDataName string, levelMessage *index.LevelMessage, g *protogen.GeneratedFile) {
	if levelMessage == nil {
		return
	}
	indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"
	if levelMessage.NextLevel != nil {
		itemName := fmt.Sprintf("item%d", depth)
		g.P(strings.Repeat("  ", depth), "for (auto&& "+itemName+" : "+parentDataName+"."+helper.ParseIndexFieldName(levelMessage.FD)+"()) {")

		parentDataName = itemName
		if levelMessage.FD.IsMap() {
			parentDataName = itemName + ".second"
		}
		genOneCppIndexLoader(depth+1, descriptor, parentDataName, levelMessage.NextLevel, g)
		g.P(strings.Repeat("  ", depth), "}")
	} else {
		if len(levelMessage.Fields) == 1 {
			// single-column index
			field := levelMessage.Fields[0] // just take the first field
			if field.FD.IsList() {
				itemName := fmt.Sprintf("item%d", depth)
				fieldName := ""
				for _, leveledFd := range field.LeveledFDList {
					fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
				}
				g.P(strings.Repeat("  ", depth), "for (auto&& "+itemName+" : "+parentDataName+fieldName+") {")
				key := itemName
				if field.FD.Enum() != nil {
					key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
				}
				g.P(strings.Repeat("  ", depth+1), indexContainerName, "["+key+"].push_back(&"+parentDataName+");")
				g.P(strings.Repeat("  ", depth), "}")
			} else {
				fieldName := ""
				for _, leveledFd := range field.LeveledFDList {
					fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
				}
				key := parentDataName + fieldName
				g.P(strings.Repeat("  ", depth), indexContainerName, "["+key+"].push_back(&"+parentDataName+");")
			}
		} else {
			// multi-column index
			var keys []string
			generateOneCppMulticolumnIndex(depth, parentDataName, descriptor, keys, g)
		}
	}
}

func generateOneCppMulticolumnIndex(depth int, parentDataName string, descriptor *index.IndexDescriptor, keys []string, g *protogen.GeneratedFile) []string {
	cursor := len(keys)
	if cursor >= len(descriptor.Fields) {
		var keyParams string
		for i, key := range keys {
			keyParams += key
			if i != len(keys)-1 {
				keyParams += ", "
			}
		}
		keyType := fmt.Sprintf("Index_%sKey", descriptor.Name)
		indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"
		g.P(strings.Repeat("  ", depth), keyType, " key{", keyParams, "};")
		g.P(strings.Repeat("  ", depth), indexContainerName, "[key].push_back(&"+parentDataName+");")
		return keys
	}
	field := descriptor.Fields[cursor]
	if field.FD.IsList() {
		itemName := fmt.Sprintf("index_item%d", cursor)
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
		}
		g.P(strings.Repeat("  ", depth), "for (auto&& "+itemName+" : "+parentDataName+fieldName+") {")
		key := itemName
		if field.FD.Enum() != nil {
			key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
		}
		keys = append(keys, key)
		keys = generateOneCppMulticolumnIndex(depth+1, parentDataName, descriptor, keys, g)
		g.P(strings.Repeat("  ", depth), "}")
	} else {
		fieldName := ""
		for _, leveledFd := range field.LeveledFDList {
			fieldName += "." + helper.ParseIndexFieldName(leveledFd) + "()"
		}
		key := parentDataName + fieldName
		keys = append(keys, key)
		keys = generateOneCppMulticolumnIndex(depth, parentDataName, descriptor, keys, g)
	}
	return keys
}

func genCppIndexFinders(gen *protogen.Plugin, messagerName string, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor) {
	descriptors := index.ParseIndexDescriptor(gen, md)
	for _, descriptor := range descriptors {
		vectorType := "Index_" + descriptor.Name + "Vector"
		mapType := "Index_" + descriptor.Name + "Map"
		indexContainerName := "index_" + strcase.ToSnake(descriptor.Name) + "_map_"

		g.P("// Index: ", descriptor.Index)
		g.P("const ", messagerName, "::", mapType, "& "+messagerName+"::Find", descriptor.Name, "() const { return "+indexContainerName+" ;}")
		g.P()

		var keyType, keyName string
		if len(descriptor.Fields) == 1 {
			// single-column index
			field := descriptor.Fields[0] // just take first field
			keyType = helper.ParseCppType(field.FD)
			keyName = helper.ParseIndexFieldNameAsFuncParam(field.FD)
		} else {
			// multi-column index
			keyType = fmt.Sprintf("const Index_%sKey&", descriptor.Name)
			keyName = "key"
		}

		g.P("const ", messagerName, "::", vectorType, "* "+messagerName+"::Find", descriptor.Name, "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
		g.P("  auto iter = ", indexContainerName, ".find(", keyName, ");")
		g.P("  if (iter == ", indexContainerName, ".end()) {")
		g.P("    return nullptr;")
		g.P("  }")
		g.P("  return &iter->second;")
		g.P("}")
		g.P()

		g.P("const ", helper.ParseCppClassType(descriptor.MD), "* "+messagerName+"::FindFirst", descriptor.Name, "(", helper.ToConstRefType(keyType), " ", keyName, ") const {")
		g.P("  auto conf = Find", descriptor.Name, "(", keyName, ");")
		g.P("  if (conf == nullptr || conf->size() == 0) {")
		g.P("    return nullptr;")
		g.P("  }")
		g.P("  return (*conf)[0];")
		g.P("}")
		g.P()

	}
}
