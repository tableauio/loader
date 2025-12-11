package indexes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/loadutil"
	"github.com/tableauio/loader/internal/options"
)

func (x *Generator) needGenerateIndex() bool {
	return options.NeedGenIndex(x.message.Desc, options.LangCPP)
}

func (x *Generator) indexMapType(index *index.LevelIndex) string {
	return fmt.Sprintf("Index_%sMap", index.Name())
}

func (x *Generator) indexMapKeyType(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take first field
		return helper.ParseCppType(field.FD)
	} else {
		// multi-column index
		return fmt.Sprintf("Index_%sKey", index.Name())
	}
}

func (x *Generator) indexMapValueVectorType(index *index.LevelIndex) string {
	return fmt.Sprintf("Index_%sVector", index.Name())
}

func (x *Generator) indexContainerName(index *index.LevelIndex, i int) string {
	if i == 0 {
		return fmt.Sprintf("index_%s_map_", strcase.ToSnake(index.Name()))
	}
	return fmt.Sprintf("index_%s_map%d_", strcase.ToSnake(index.Name()), i)
}

func (x *Generator) indexKeys(index *index.LevelIndex) helper.MapKeySlice {
	var keys helper.MapKeySlice
	for _, field := range index.ColFields {
		keys = keys.AddMapKey(helper.MapKey{
			Type: helper.ParseCppType(field.FD),
			Name: helper.ParseIndexFieldNameAsFuncParam(field.FD),
		})
	}
	return keys
}

func (x *Generator) genHppIndexFinders() {
	if !x.needGenerateIndex() {
		return
	}
	var once sync.Once
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.g.P()
			once.Do(func() { x.g.P(helper.Indent(1), "// Index accessers.") })
			x.g.P(helper.Indent(1), "// Index: ", index.Index)
			x.g.P(" public:")
			mapType := x.indexMapType(index)
			keyType := x.indexMapKeyType(index)
			vectorType := x.indexMapValueVectorType(index)
			valueType := x.mapValueType(index)
			keys := x.indexKeys(index)
			hasher := "" // std::hash by default
			if len(index.ColFields) != 1 {
				// multi-column index
				keyHasherType := fmt.Sprintf("Index_%sKeyHasher", index.Name())
				hasher = ", " + keyHasherType
				// Generate key struct
				x.g.P(helper.Indent(1), "struct ", keyType, " {")
				for _, key := range keys {
					x.g.P(helper.Indent(2), key.Type, " ", key.Name, ";")
				}
				x.g.P("#if __cplusplus >= 202002L")
				x.g.P(helper.Indent(2), "bool operator==(const ", keyType, "& other) const = default;")
				x.g.P("#else")
				x.g.P(helper.Indent(2), "bool operator==(const ", keyType, "& other) const {")
				x.g.P(helper.Indent(3), "return std::tie(", keys.GenGetArguments(), ") == std::tie(", keys.GenOtherArguments("other"), ");")
				x.g.P(helper.Indent(2), "}")
				x.g.P("#endif")
				x.g.P(helper.Indent(1), "};")

				// Generate key hasher struct
				x.g.P(helper.Indent(1), "struct ", keyHasherType, " {")
				x.g.P(helper.Indent(2), "std::size_t operator()(const ", keyType, "& key) const {")
				x.g.P(helper.Indent(3), "return util::SugaredHashCombine(", keys.GenOtherArguments("key"), ");")
				x.g.P(helper.Indent(2), "}")
				x.g.P(helper.Indent(1), "};")
			}
			x.g.P(helper.Indent(1), "using ", vectorType, " = std::vector<const ", valueType, "*>;")
			x.g.P(helper.Indent(1), "using ", mapType, " = std::unordered_map<", keyType, ", ", vectorType, hasher, ">;")
			x.g.P(helper.Indent(1), "// Finds the index: key(", index.Index, ") to value(", vectorType, ") hashmap.")
			x.g.P(helper.Indent(1), "// One key may correspond to multiple values, which are represented by a vector.")
			x.g.P(helper.Indent(1), "const ", mapType, "& Find", index.Name(), "Map() const;")
			x.g.P(helper.Indent(1), "// Finds a vector of all values of the given key(s).")
			x.g.P(helper.Indent(1), "const ", vectorType, "* Find", index.Name(), "(", keys.GenGetParams(), ") const;")
			x.g.P(helper.Indent(1), "// Finds the first value of the given key(s).")
			x.g.P(helper.Indent(1), "const ", valueType, "* FindFirst", index.Name(), "(", keys.GenGetParams(), ") const;")
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				partKeys := x.keys[:i]
				x.g.P(helper.Indent(1), "// Finds the index: key(", index.Index, ") to value(", vectorType, "),")
				x.g.P(helper.Indent(1), "// which is the upper ", loadutil.Ordinal(i), "-level hashmap specified by (", partKeys.GenGetArguments(), ").")
				x.g.P(helper.Indent(1), "// One key may correspond to multiple values, which are represented by a vector.")
				x.g.P(helper.Indent(1), "const ", mapType, "* Find", index.Name(), "Map(", partKeys.GenGetParams(), ") const;")
				x.g.P(helper.Indent(1), "// Finds a vector of all values of the given key(s) in the upper ", loadutil.Ordinal(i), "-level hashmap specified by (", partKeys.GenGetArguments(), ").")
				x.g.P(helper.Indent(1), "const ", vectorType, "* Find", index.Name(), "(", partKeys.GenGetParams(), ", ", keys.GenGetParams(), ") const;")
				x.g.P(helper.Indent(1), "// Finds the first value of the given key(s) in the upper ", loadutil.Ordinal(i), "-level hashmap specified by (", partKeys.GenGetArguments(), ").")
				x.g.P(helper.Indent(1), "const ", valueType, "* FindFirst", index.Name(), "(", partKeys.GenGetParams(), ", ", keys.GenGetParams(), ") const;")
			}
			x.g.P()

			x.g.P(" private:")
			x.g.P(helper.Indent(1), mapType, " ", x.indexContainerName(index, 0), ";")
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				if i == 1 {
					x.g.P(helper.Indent(1), "std::unordered_map<", x.keys[0].Type, ", ", mapType, "> ", x.indexContainerName(index, i), ";")
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P(helper.Indent(1), "std::unordered_map<", levelIndexKeyType, ", ", mapType, ", ", levelIndexKeyType, "Hasher> ", x.indexContainerName(index, i), ";")
				}
			}
		}
	}
}

func (x *Generator) genIndexLoader() {
	if !x.needGenerateIndex() {
		return
	}
	defer x.genIndexSorter()
	x.g.P(helper.Indent(1), "// Index init.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.g.P(helper.Indent(1), x.indexContainerName(index, 0), ".clear();")
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				x.g.P(helper.Indent(1), x.indexContainerName(index, i), ".clear();")
			}
		}
	}
	parentDataName := "data_"
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.genOneCppIndexLoader(levelMessage.MapDepth, levelMessage.Depth, index, parentDataName)
		}
		itemName := fmt.Sprintf("item%d", levelMessage.Depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGenIndex() {
			break
		}
		x.g.P(helper.Indent(levelMessage.Depth), "for (auto&& ", itemName, " : ", parentDataName, x.fieldGetter(levelMessage.FD), ") {")
		parentDataName = itemName
		if levelMessage.FD.IsMap() {
			x.g.P(helper.Indent(levelMessage.Depth+1), "auto k", levelMessage.MapDepth, " = ", itemName, ".first;")
			parentDataName = itemName + ".second"
		}
		defer x.g.P(helper.Indent(levelMessage.Depth), "}")
	}
}

func (x *Generator) genOneCppIndexLoader(depth int, ident int, index *index.LevelIndex, parentDataName string) {
	x.g.P(helper.Indent(ident), "{")
	x.g.P(helper.Indent(ident+1), "// Index: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			x.g.P(helper.Indent(ident+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
			key := itemName
			if field.FD.Enum() != nil {
				key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
			}
			x.genLoader(depth, ident+2, index, key, parentDataName)
			x.g.P(helper.Indent(ident+1), "}")
		} else {
			key := parentDataName + fieldName
			x.genLoader(depth, ident+1, index, key, parentDataName)
		}
	} else {
		// multi-column index
		x.generateOneCppMulticolumnIndex(depth, ident, index, parentDataName, nil)
	}
	x.g.P(helper.Indent(ident), "}")
}

func (x *Generator) generateOneCppMulticolumnIndex(depth, ident int, index *index.LevelIndex, parentDataName string, keys helper.MapKeySlice) {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		keyType := x.indexMapKeyType(index)
		x.g.P(helper.Indent(ident+1), keyType, " key{", keys.GenGetArguments(), "};")
		x.genLoader(depth, ident+1, index, "key", parentDataName)
		return
	}
	field := index.ColFields[cursor]
	fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
	if field.FD.IsList() {
		itemName := fmt.Sprintf("index_item%d", cursor)
		x.g.P(helper.Indent(ident+1), "for (auto&& ", itemName, " : ", parentDataName, fieldName, ") {")
		key := itemName
		if field.FD.Enum() != nil {
			key = "static_cast<" + helper.ParseCppType(field.FD) + ">(" + key + ")"
		}
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneCppMulticolumnIndex(depth, ident+1, index, parentDataName, keys)
		x.g.P(helper.Indent(ident+1), "}")
	} else {
		key := parentDataName + fieldName
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneCppMulticolumnIndex(depth, ident, index, parentDataName, keys)
	}
}

func (x *Generator) genLoader(depth, ident int, index *index.LevelIndex, key, parentDataName string) {
	x.g.P(helper.Indent(ident), x.indexContainerName(index, 0), "[", key, "].push_back(&", parentDataName, ");")
	for i := 1; i <= depth-2; i++ {
		if i > len(x.keys) {
			break
		}
		if i == 1 {
			x.g.P(helper.Indent(ident), x.indexContainerName(index, i), "[k1][", key, "].push_back(&", parentDataName, ");")
		} else {
			var fields []string
			for j := 1; j <= i; j++ {
				fields = append(fields, fmt.Sprintf("k%d", j))
			}
			x.g.P(helper.Indent(ident), x.indexContainerName(index, i), "[{", strings.Join(fields, ", "), "}][", key, "].push_back(&", parentDataName, ");")
		}
	}
}

func (x *Generator) genIndexSorter() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			if len(index.SortedColFields) != 0 {
				valueType := x.mapValueType(index)
				x.g.P(helper.Indent(1), "// Index(sort): ", index.Index)
				indexContainerName := x.indexContainerName(index, 0)
				sorterDef := "auto " + indexContainerName + "sorter = []("
				x.g.P(helper.Indent(1), sorterDef, "const ", valueType, "* a,")
				x.g.P(helper.Indent(1), helper.Whitespace(len(sorterDef)), "const ", valueType, "* b) {")
				for i, field := range index.SortedColFields {
					fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
					fieldName = strings.Replace(fieldName, ".", "->", 1)
					if i == len(index.SortedColFields)-1 {
						x.g.P(helper.Indent(2), "return a", fieldName, " < b", fieldName, ";")
					} else {
						x.g.P(helper.Indent(2), "if (a", fieldName, " != b", fieldName, ") {")
						x.g.P(helper.Indent(3), "return a", fieldName, " < b", fieldName, ";")
						x.g.P(helper.Indent(2), "}")
					}
				}
				x.g.P(helper.Indent(1), "};")
				x.g.P(helper.Indent(1), "for (auto&& item : ", indexContainerName, ") {")
				x.g.P(helper.Indent(2), "std::sort(item.second.begin(), item.second.end(), ", indexContainerName, "sorter);")
				x.g.P(helper.Indent(1), "}")
				for i := 1; i <= levelMessage.MapDepth-2; i++ {
					if i > len(x.keys) {
						break
					}
					x.g.P(helper.Indent(1), "for (auto&& item : ", x.indexContainerName(index, i), ") {")
					x.g.P(helper.Indent(2), "for (auto&& item1 : item.second) {")
					x.g.P(helper.Indent(3), "std::sort(item1.second.begin(), item1.second.end(), ", indexContainerName, "sorter);")
					x.g.P(helper.Indent(2), "}")
					x.g.P(helper.Indent(1), "}")
				}
			}
		}
	}
}

func (x *Generator) genCppIndexFinders() {
	if !x.needGenerateIndex() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			vectorType := x.indexMapValueVectorType(index)
			mapType := x.indexMapType(index)
			indexContainerName := x.indexContainerName(index, 0)
			messagerName := x.messagerName()

			x.g.P("// Index: ", index.Index)
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

			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				indexContainerName := x.indexContainerName(index, i)
				partKeys := x.keys[:i]
				partParams := partKeys.GenGetParams()
				partArgs := partKeys.GenGetArguments()
				x.g.P("const ", messagerName, "::", mapType, "* ", messagerName, "::Find", index.Name(), "Map(", partParams, ") const {")
				if len(partKeys) == 1 {
					x.g.P(helper.Indent(1), "auto iter = ", indexContainerName, ".find(", partArgs, ");")
				} else {
					x.g.P(helper.Indent(1), "auto iter = ", indexContainerName, ".find({", partArgs, "});")
				}
				x.g.P(helper.Indent(1), "if (iter == ", indexContainerName, ".end()) {")
				x.g.P(helper.Indent(2), "return nullptr;")
				x.g.P(helper.Indent(1), "}")
				x.g.P(helper.Indent(1), "return &iter->second;")
				x.g.P("}")
				x.g.P()

				x.g.P("const ", messagerName, "::", vectorType, "* ", messagerName, "::Find", index.Name(), "(", partParams, ", ", params, ") const {")
				x.g.P(helper.Indent(1), "auto map = Find", index.Name(), "Map(", partArgs, ");")
				x.g.P(helper.Indent(1), "if (map == nullptr) {")
				x.g.P(helper.Indent(2), "return nullptr;")
				x.g.P(helper.Indent(1), "}")
				if len(index.ColFields) == 1 {
					x.g.P(helper.Indent(1), "auto iter = map->find(", args, ");")
				} else {
					x.g.P(helper.Indent(1), "auto iter = map->find({", args, "});")
				}
				x.g.P(helper.Indent(1), "if (iter == map->end()) {")
				x.g.P(helper.Indent(2), "return nullptr;")
				x.g.P(helper.Indent(1), "}")
				x.g.P(helper.Indent(1), "return &iter->second;")
				x.g.P("}")
				x.g.P()

				x.g.P("const ", x.mapValueType(index), "* ", messagerName, "::FindFirst", index.Name(), "(", partParams, ", ", params, ") const {")
				x.g.P(helper.Indent(1), "auto conf = Find", index.Name(), "(", partArgs, ", ", args, ");")
				x.g.P(helper.Indent(1), "if (conf == nullptr || conf->empty()) {")
				x.g.P(helper.Indent(2), "return nullptr;")
				x.g.P(helper.Indent(1), "}")
				x.g.P(helper.Indent(1), "return conf->front();")
				x.g.P("}")
				x.g.P()
			}
		}
	}
}
