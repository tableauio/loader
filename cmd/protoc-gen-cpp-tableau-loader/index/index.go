package index

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Generator struct {
	g          *protogen.GeneratedFile
	descriptor *index.IndexDescriptor
	message    *protogen.Message
}

func NewGenerator(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, message *protogen.Message) *Generator {
	return &Generator{
		g:          g,
		descriptor: descriptor,
		message:    message,
	}
}

func (x *Generator) Generate() bool {
	return options.NeedGenIndex(x.message.Desc, options.LangCPP)
}

func (x *Generator) messagerName() string {
	return string(x.message.Desc.Name())
}

func (x *Generator) mapType(index *index.LevelIndex) string {
	return fmt.Sprintf("Index_%sMap", index.Name())
}

func (x *Generator) mapKeyType(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take first field
		return helper.ParseCppType(field.FD)
	} else {
		// multi-column index
		return fmt.Sprintf("Index_%sKey", index.Name())
	}
}

func (x *Generator) mapValueType(index *index.LevelIndex) string {
	return helper.ParseCppClassType(index.MD)
}

func (x *Generator) mapValueVectorType(index *index.LevelIndex) string {
	return fmt.Sprintf("Index_%sVector", index.Name())
}

func (x *Generator) indexContainerName(index *index.LevelIndex) string {
	return fmt.Sprintf("index_%s_map_", strcase.ToSnake(index.Name()))
}

func (x *Generator) indexKeys(index *index.LevelIndex) helper.MapKeys {
	var keys []helper.MapKey
	for _, field := range index.ColFields {
		keys = append(keys, helper.MapKey{
			Type:      helper.ParseCppType(field.FD),
			Name:      helper.ParseIndexFieldNameAsFuncParam(field.FD),
			FieldName: helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD),
		})
	}
	return keys
}

func (x *Generator) fieldGetter(fd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf(".%s()", helper.ParseIndexFieldName(fd))
}

func (x *Generator) parseKeyFieldName(field *index.LevelField) string {
	var fieldName string
	for _, leveledFd := range field.LeveledFDList {
		fieldName += x.fieldGetter(leveledFd)
	}
	return fieldName
}

func (x *Generator) GenHppIndexFinders() {
	if !x.Generate() {
		return
	}
	x.g.P()
	x.g.P(helper.Indent(1), "// Index accessers.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.g.P(helper.Indent(1), "// Index: ", index.Index)
			x.g.P(" public:")
			mapType := x.mapType(index)
			keyType := x.mapKeyType(index)
			vectorType := x.mapValueVectorType(index)
			valueType := x.mapValueType(index)
			keys := x.indexKeys(index)
			hasher := "" // std::hash by default
			if len(index.ColFields) != 1 {
				// multi-column index
				keyHasherType := fmt.Sprintf("Index_%sKeyHasher", index.Name())
				hasher = ", " + keyHasherType
				// Generate key struct
				x.g.P(helper.Indent(1), "struct ", keyType, " {")
				var equalities []string
				for _, key := range keys {
					x.g.P(helper.Indent(2), key.Type, " ", key.FieldName, ";")
					equalities = append(equalities, key.FieldName+" == other."+key.FieldName)
				}
				x.g.P("#if __cplusplus >= 202002L")
				x.g.P(helper.Indent(2), "bool operator==(const ", keyType, "& other) const = default;")
				x.g.P("#else")
				x.g.P(helper.Indent(2), "bool operator==(const ", keyType, "& other) const {")
				x.g.P(helper.Indent(3), "return ", strings.Join(equalities, " && "), ";")
				x.g.P(helper.Indent(2), "}")
				x.g.P("#endif")
				x.g.P(helper.Indent(1), "};")

				// Generate key hasher struct
				x.g.P(helper.Indent(1), "struct ", keyHasherType, " {")
				var combinedKeys []string
				for _, key := range keys {
					combinedKeys = append(combinedKeys, "key."+key.FieldName)
				}
				x.g.P(helper.Indent(2), "std::size_t operator()(const ", keyType, "& key) const {")
				x.g.P(helper.Indent(3), "return util::SugaredHashCombine(", strings.Join(combinedKeys, ", "), ");")
				x.g.P(helper.Indent(2), "}")
				x.g.P(helper.Indent(1), "};")
			}
			x.g.P(helper.Indent(1), "using ", vectorType, " = std::vector<const ", valueType, "*>;")
			x.g.P(helper.Indent(1), "using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, hasher, ">;")
			x.g.P(helper.Indent(1), "// Finds the index (", index.Index, ") to value (", vectorType, ") hash map.")
			x.g.P(helper.Indent(1), "// One key may correspond to multiple values, which are contained by a vector.")
			x.g.P(helper.Indent(1), "const ", mapType, "& Find", index.Name(), "() const;")
			x.g.P(helper.Indent(1), "// Finds a vector of all values of the given key(s).")
			x.g.P(helper.Indent(1), "const ", vectorType, "* Find", index.Name(), "(", keys.GenGetParams(), ") const;")
			x.g.P(helper.Indent(1), "// Finds the first value of the given key(s).")
			x.g.P(helper.Indent(1), "const ", valueType, "* FindFirst", index.Name(), "(", keys.GenGetParams(), ") const;")
			x.g.P()
			x.g.P(" private:")
			x.g.P(helper.Indent(1), mapType, " ", x.indexContainerName(index), ";")
			x.g.P()
		}
	}
}

func (x *Generator) GenCppIndexLoader() {
	if !x.Generate() {
		return
	}
	x.g.P(helper.Indent(1), "// Index init.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.g.P(helper.Indent(1), x.indexContainerName(index), ".clear();")
		}
	}
	parentDataName := "data_"
	depth := 1
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.genOneCppIndexLoader(depth, index, parentDataName)
		}
		itemName := fmt.Sprintf("item%d", depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGen() {
			break
		}
		x.g.P(helper.Indent(depth), "for (auto&& ", itemName, " : ", parentDataName, x.fieldGetter(levelMessage.FD), ") {")
		parentDataName = itemName
		if levelMessage.FD.IsMap() {
			parentDataName = itemName + ".second"
		}
		depth++
	}
	for i := depth - 1; i > 0; i-- {
		x.g.P(helper.Indent(i), "}")
	}
	x.genIndexSorter()
}

func (x *Generator) genOneCppIndexLoader(depth int, index *index.LevelIndex, parentDataName string) {
	x.g.P(helper.Indent(depth), "{")
	x.g.P(helper.Indent(depth+1), "// Index: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		fieldName := x.parseKeyFieldName(field)
		indexContainerName := x.indexContainerName(index)
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			x.g.P(helper.Indent(depth+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
			key := itemName
			if field.FD.Enum() != nil {
				key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
			}
			x.g.P(helper.Indent(depth+2), indexContainerName, "[", key, "].push_back(&", parentDataName, ");")
			x.g.P(helper.Indent(depth+1), "}")
		} else {
			key := parentDataName + fieldName
			x.g.P(helper.Indent(depth+1), indexContainerName, "[", key, "].push_back(&", parentDataName, ");")
		}
	} else {
		// multi-column index
		x.generateOneCppMulticolumnIndex(depth, index, parentDataName, nil)
	}
	x.g.P(helper.Indent(depth), "}")
}

func (x *Generator) generateOneCppMulticolumnIndex(depth int, index *index.LevelIndex, parentDataName string, keys helper.MapKeys) {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		keyType := x.mapKeyType(index)
		indexContainerName := x.indexContainerName(index)
		x.g.P(helper.Indent(depth+1), keyType, " key{", keys.GenGetArguments(), "};")
		x.g.P(helper.Indent(depth+1), indexContainerName, "[key].push_back(&", parentDataName, ");")
		return
	}
	field := index.ColFields[cursor]
	fieldName := x.parseKeyFieldName(field)
	if field.FD.IsList() {
		itemName := fmt.Sprintf("index_item%d", cursor)
		x.g.P(helper.Indent(depth+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
		key := itemName
		if field.FD.Enum() != nil {
			key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
		}
		keys = append(keys, helper.MapKey{Name: key})
		x.generateOneCppMulticolumnIndex(depth+1, index, parentDataName, keys)
		x.g.P(helper.Indent(depth+1), "}")
	} else {
		key := parentDataName + fieldName
		keys = append(keys, helper.MapKey{Name: key})
		x.generateOneCppMulticolumnIndex(depth, index, parentDataName, keys)
	}
}

func (x *Generator) genIndexSorter() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			if len(index.SortedColFields) != 0 {
				valueType := x.mapValueType(index)
				x.g.P(helper.Indent(1), "// Index(sort): ", index.Index)
				x.g.P(helper.Indent(1), "for (auto&& item : ", x.indexContainerName(index), ") {")
				x.g.P(helper.Indent(2), "std::sort(item.second.begin(), item.second.end(),")
				x.g.P(helper.Indent(7), "[](const ", valueType, "* a, const ", valueType, "* b) {")
				for i, field := range index.SortedColFields {
					fieldName := strings.Replace(x.parseKeyFieldName(field), ".", "->", 1)
					if i == len(index.SortedColFields)-1 {
						x.g.P(helper.Indent(8), "return a", fieldName, " < b", fieldName, ";")
					} else {
						x.g.P(helper.Indent(8), "if (a", fieldName, " != b", fieldName, ") {")
						x.g.P(helper.Indent(9), "return a", fieldName, " < b", fieldName, ";")
						x.g.P(helper.Indent(8), "}")
					}
				}
				x.g.P(helper.Indent(7), "});")
				x.g.P(helper.Indent(1), "}")
			}
		}
	}
}

func (x *Generator) GenCppIndexFinders() {
	if !x.Generate() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			vectorType := x.mapValueVectorType(index)
			mapType := x.mapType(index)
			indexContainerName := x.indexContainerName(index)
			messagerName := x.messagerName()

			x.g.P("// Index: ", index.Index)
			x.g.P("const ", messagerName, "::", mapType, "& ", messagerName, "::Find", index.Name(), "() const { return ", indexContainerName, " ;}")
			x.g.P()

			keys := x.indexKeys(index)
			params := keys.GenGetParams()
			args := keys.GenGetArguments()
			x.g.P("const ", messagerName, "::", vectorType, "* ", messagerName, "::Find", index.Name(), "(", params, ") const {")
			if len(index.ColFields) == 1 {
				x.g.P(helper.Indent(1), "auto iter = ", indexContainerName, ".find(", args, ");")
			} else {
				x.g.P(helper.Indent(1), "auto iter = ", indexContainerName, ".find({", args, "});")
			}
			x.g.P(helper.Indent(1), "if (iter == ", indexContainerName, ".end()) {")
			x.g.P(helper.Indent(2), "return nullptr;")
			x.g.P(helper.Indent(1), "}")
			x.g.P(helper.Indent(1), "return &iter->second;")
			x.g.P("}")
			x.g.P()

			x.g.P("const ", x.mapValueType(index), "* ", messagerName, "::FindFirst", index.Name(), "(", params, ") const {")
			x.g.P(helper.Indent(1), "auto conf = Find", index.Name(), "(", args, ");")
			x.g.P(helper.Indent(1), "if (conf == nullptr || conf->empty()) {")
			x.g.P(helper.Indent(2), "return nullptr;")
			x.g.P(helper.Indent(1), "}")
			x.g.P(helper.Indent(1), "return conf->front();")
			x.g.P("}")
			x.g.P()
		}
	}
}
