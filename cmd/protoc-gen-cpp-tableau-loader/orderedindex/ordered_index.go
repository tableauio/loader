package orderedindex

import (
	"fmt"
	"strings"
	"sync"

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

func (x *Generator) NeedGenerate() bool {
	return options.NeedGenOrderedIndex(x.message.Desc, options.LangCPP)
}

func (x *Generator) messagerName() string {
	return string(x.message.Desc.Name())
}

func (x *Generator) mapType(index *index.LevelIndex) string {
	return fmt.Sprintf("OrderedIndex_%sMap", index.Name())
}

func (x *Generator) mapKeyType(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take first field
		return helper.ParseOrderedIndexKeyType(field.FD)
	} else {
		// multi-column index
		return fmt.Sprintf("OrderedIndex_%sKey", index.Name())
	}
}

func (x *Generator) mapValueType(index *index.LevelIndex) string {
	return helper.ParseCppClassType(index.MD)
}

func (x *Generator) mapValueVectorType(index *index.LevelIndex) string {
	return fmt.Sprintf("OrderedIndex_%sVector", index.Name())
}

func (x *Generator) indexContainerName(index *index.LevelIndex) string {
	return fmt.Sprintf("ordered_index_%s_map_", strcase.ToSnake(index.Name()))
}

func (x *Generator) indexKeys(index *index.LevelIndex) helper.MapKeys {
	var keys helper.MapKeys
	for _, field := range index.ColFields {
		keys = keys.AddMapKey(helper.MapKey{
			Type: helper.ParseOrderedIndexKeyType(field.FD),
			Name: helper.ParseIndexFieldNameAsFuncParam(field.FD),
		})
	}
	return keys
}

func (x *Generator) fieldGetter(fd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf(".%s()", helper.ParseIndexFieldName(fd))
}

func (x *Generator) parseKeyFieldNameAndSuffix(field *index.LevelField) (string, string) {
	var fieldName, suffix string
	for i, leveledFd := range field.LeveledFDList {
		fieldName += x.fieldGetter(leveledFd)
		if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
			switch leveledFd.Message().FullName() {
			case "google.protobuf.Timestamp", "google.protobuf.Duration":
				suffix = ".seconds()"
			default:
			}
		}
	}
	return fieldName, suffix
}

func (x *Generator) GenHppOrderedIndexFinders() {
	if !x.NeedGenerate() {
		return
	}
	var once sync.Once
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P()
			once.Do(func() { x.g.P(helper.Indent(1), "// OrderedIndex accessers.") })
			x.g.P(helper.Indent(1), "// OrderedIndex: ", index.Index)
			x.g.P(" public:")
			mapType := x.mapType(index)
			keyType := x.mapKeyType(index)
			vectorType := x.mapValueVectorType(index)
			valueType := x.mapValueType(index)
			keys := x.indexKeys(index)
			if len(index.ColFields) != 1 {
				// multi-column index
				// Generate key struct
				x.g.P(helper.Indent(1), "struct ", keyType, " {")
				for _, key := range keys {
					x.g.P(helper.Indent(2), key.Type, " ", key.Name, ";")
				}
				x.g.P("#if __cplusplus >= 202002L")
				x.g.P(helper.Indent(2), "auto operator<=>(const ", keyType, "& other) const = default;")
				x.g.P("#else")
				x.g.P(helper.Indent(2), "bool operator<(const ", keyType, "& other) const {")
				x.g.P(helper.Indent(3), "return std::tie(", keys.GenGetArguments(), ") < std::tie(", keys.GenOtherArguments("other"), ");")
				x.g.P(helper.Indent(2), "}")
				x.g.P("#endif")
				x.g.P(helper.Indent(1), "};")
			}
			x.g.P(helper.Indent(1), "using ", vectorType, " = std::vector<const ", valueType, "*>;")
			x.g.P(helper.Indent(1), "using ", mapType, " = std::map<", keyType, ", ", vectorType, ">;")
			x.g.P(helper.Indent(1), "// Finds the ordered index (", index.Index, ") to value (", vectorType, ") map.")
			x.g.P(helper.Indent(1), "// One key may correspond to multiple values, which are contained by a vector.")
			x.g.P(helper.Indent(1), "const ", mapType, "& Find", index.Name(), "Map() const;")
			x.g.P(helper.Indent(1), "// Finds a vector of all values of the given key(s).")
			x.g.P(helper.Indent(1), "const ", vectorType, "* Find", index.Name(), "(", keys.GenGetParams(), ") const;")
			x.g.P(helper.Indent(1), "// Finds the first value of the given key(s).")
			x.g.P(helper.Indent(1), "const ", helper.ParseCppClassType(index.MD), "* FindFirst", index.Name(), "(", keys.GenGetParams(), ") const;")
			x.g.P()

			x.g.P(" private:")
			x.g.P(helper.Indent(1), mapType, " ", x.indexContainerName(index), ";")
		}
	}
}

func (x *Generator) GenCppOrderedIndexLoader() {
	if !x.NeedGenerate() {
		return
	}
	x.g.P(helper.Indent(1), "// OrderedIndex init.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P(helper.Indent(1), x.indexContainerName(index), ".clear();")
		}
	}
	parentDataName := "data_"
	depth := 1
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.genOneCppOrderedIndexLoader(depth, index, parentDataName)
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
	x.genOrderedIndexSorter()
}

func (x *Generator) genOneCppOrderedIndexLoader(depth int, index *index.LevelIndex, parentDataName string) {
	x.g.P(helper.Indent(depth), "{")
	x.g.P(helper.Indent(depth+1), "// OrderedIndex: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		fieldName, suffix := x.parseKeyFieldNameAndSuffix(field)
		indexContainerName := x.indexContainerName(index)
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			x.g.P(helper.Indent(depth+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
			key := itemName + suffix
			if field.FD.Enum() != nil {
				key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
			}
			x.g.P(helper.Indent(depth+2), indexContainerName, "[", key, "].push_back(&", parentDataName, ");")
			x.g.P(helper.Indent(depth+1), "}")
		} else {
			key := parentDataName + fieldName + suffix
			x.g.P(helper.Indent(depth+1), indexContainerName, "[", key, "].push_back(&", parentDataName, ");")
		}
	} else {
		// multi-column index
		x.generateOneCppMulticolumnOrderedIndex(depth, index, parentDataName, nil)
	}
	x.g.P(helper.Indent(depth), "}")
}

func (x *Generator) generateOneCppMulticolumnOrderedIndex(depth int, index *index.LevelIndex, parentDataName string, keys helper.MapKeys) {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		keyType := x.mapKeyType(index)
		indexContainerName := x.indexContainerName(index)
		x.g.P(helper.Indent(depth+1), keyType, " key{", keys.GenGetArguments(), "};")
		x.g.P(helper.Indent(depth+1), indexContainerName, "[key].push_back(&", parentDataName, ");")
		return
	}
	field := index.ColFields[cursor]
	fieldName, suffix := x.parseKeyFieldNameAndSuffix(field)
	if field.FD.IsList() {
		itemName := fmt.Sprintf("index_item%d", cursor)
		x.g.P(helper.Indent(depth+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
		key := itemName + suffix
		if field.FD.Enum() != nil {
			key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
		}
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneCppMulticolumnOrderedIndex(depth+1, index, parentDataName, keys)
		x.g.P(helper.Indent(depth+1), "}")
	} else {
		key := parentDataName + fieldName + suffix
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneCppMulticolumnOrderedIndex(depth, index, parentDataName, keys)
	}
}

func (x *Generator) genOrderedIndexSorter() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			if len(index.SortedColFields) != 0 {
				valueType := x.mapValueType(index)
				x.g.P(helper.Indent(1), "// OrderedIndex(sort): ", index.Index)
				x.g.P(helper.Indent(1), "for (auto&& item : ", x.indexContainerName(index), ") {")
				x.g.P(helper.Indent(2), "std::sort(item.second.begin(), item.second.end(),")
				x.g.P(helper.Indent(7), "[](const ", valueType, "* a, const ", valueType, "* b) {")
				for i, field := range index.SortedColFields {
					fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
					fieldName = strings.Replace(fieldName, ".", "->", 1)
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

func (x *Generator) GenCppOrderedIndexFinders() {
	if !x.NeedGenerate() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			vectorType := x.mapValueVectorType(index)
			mapType := x.mapType(index)
			indexContainerName := x.indexContainerName(index)
			messagerName := x.messagerName()

			x.g.P("// OrderedIndex: ", index.Index)
			x.g.P("const ", messagerName, "::", mapType, "& ", messagerName, "::Find", index.Name(), "Map() const { return ", indexContainerName, " ;}")
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
