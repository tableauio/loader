package indexes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/options"
)

func (x *Generator) needGenerateOrderedIndex() bool {
	return options.NeedGenOrderedIndex(x.message.Desc, options.LangCS)
}

func (x *Generator) orderedIndexMapType(index *index.LevelIndex) string {
	return fmt.Sprintf("OrderedIndex_%sMap", index.Name())
}

func (x *Generator) orderedIndexMapKeyType(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take first field
		return helper.ParseOrderedIndexKeyType(field.FD)
	} else {
		// multi-column index
		return fmt.Sprintf("OrderedIndex_%sKey", index.Name())
	}
}

func (x *Generator) orderedIndexContainerName(index *index.LevelIndex, i int) string {
	if i == 0 {
		return fmt.Sprintf("_orderedIndex%sMap", strcase.ToCamel(index.Name()))
	}
	return fmt.Sprintf("_orderedIndex%sMap%d", strcase.ToCamel(index.Name()), i)
}

func (x *Generator) orderedIndexKeys(index *index.LevelIndex) helper.MapKeySlice {
	var keys helper.MapKeySlice
	for _, field := range index.ColFields {
		keys = keys.AddMapKey(helper.MapKey{
			Type:      helper.ParseOrderedIndexKeyType(field.FD),
			Name:      helper.ParseIndexFieldNameAsFuncParam(field.FD),
			FieldName: helper.ParseIndexFieldNameAsKeyStructFieldName(field.FD),
		})
	}
	return keys
}

func (x *Generator) genOrderedIndexTypeDef() {
	if !x.needGenerateOrderedIndex() {
		return
	}
	var once sync.Once
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			once.Do(func() { x.g.P(helper.Indent(2), "// OrderedIndex types.") })
			x.g.P(helper.Indent(2), "// OrderedIndex: ", index.Index)
			mapType := x.orderedIndexMapType(index)
			keyType := x.orderedIndexMapKeyType(index)
			valueType := x.mapValueType(index)
			keys := x.orderedIndexKeys(index)
			if len(index.ColFields) != 1 {
				// Generate key struct
				x.g.P(helper.Indent(2), "public readonly struct ", keyType, " : IComparable<", keyType, ">")
				x.g.P(helper.Indent(2), "{")
				for _, key := range keys {
					x.g.P(helper.Indent(3), "public ", key.Type, " ", key.FieldName, " { get; }")
				}
				x.g.P()
				x.g.P(helper.Indent(3), "public ", keyType, "(", keys.GenGetParams(), ")")
				x.g.P(helper.Indent(3), "{")
				for _, key := range keys {
					x.g.P(helper.Indent(4), key.FieldName, " = ", key.Name, ";")
				}
				x.g.P(helper.Indent(3), "}")
				x.g.P()
				x.g.P(helper.Indent(3), "public int CompareTo(", keyType, " other) =>")
				x.g.P(helper.Indent(4), "(", keys.GenCustom(func(key helper.MapKey) string { return key.FieldName }, ", "), ").CompareTo((", keys.GenCustom(func(key helper.MapKey) string { return "other." + key.FieldName }, ", "), "));")
				x.g.P(helper.Indent(2), "}")
				x.g.P()
			}
			x.g.P(helper.Indent(2), "public class ", mapType, " : SortedDictionary<", keyType, ", List<", valueType, ">>;")
			x.g.P()

			x.g.P(helper.Indent(2), "private ", mapType, " ", x.orderedIndexContainerName(index, 0), " = [];")
			x.g.P()
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				if i == 1 {
					x.g.P(helper.Indent(2), "private Dictionary<", x.keys[0].Type, ", ", mapType, "> ", x.orderedIndexContainerName(index, i), " = [];")
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P(helper.Indent(2), "private Dictionary<", levelIndexKeyType, ", ", mapType, "> ", x.orderedIndexContainerName(index, i), " = [];")
				}
				x.g.P()
			}
		}
	}
}

func (x *Generator) genOrderedIndexLoader() {
	if !x.needGenerateOrderedIndex() {
		return
	}
	defer x.genOrderedIndexSorter()
	x.g.P(helper.Indent(3), "// OrderedIndex init.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P(helper.Indent(3), x.orderedIndexContainerName(index, 0), ".Clear();")
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				x.g.P(helper.Indent(3), x.orderedIndexContainerName(index, i), ".Clear();")
			}
		}
	}
	parentDataName := "_data"
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.genOneOrderedIndexLoader(levelMessage.MapDepth, levelMessage.Depth, index, parentDataName)
		}
		itemName := fmt.Sprintf("item%d", levelMessage.Depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGenOrderedIndex() {
			break
		}
		x.g.P(helper.Indent(levelMessage.Depth+2), "foreach (var ", itemName, " in ", parentDataName, x.fieldGetter(levelMessage.FD), ")")
		x.g.P(helper.Indent(levelMessage.Depth+2), "{")
		parentDataName = itemName
		if levelMessage.FD.IsMap() {
			x.g.P(helper.Indent(levelMessage.Depth+3), "var k", levelMessage.MapDepth, " = ", itemName, ".Key;")
			parentDataName = itemName + ".Value"
		}
		defer x.g.P(helper.Indent(levelMessage.Depth+2), "}")
	}
}

func (x *Generator) genOneOrderedIndexLoader(depth int, ident int, index *index.LevelIndex, parentDataName string) {
	x.g.P(helper.Indent(ident+2), "{")
	x.g.P(helper.Indent(ident+3), "// OrderedIndex: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		fieldName, suffix := x.parseKeyFieldNameAndSuffix(field)
		if field.FD.IsList() {
			itemName := fmt.Sprintf("item%d", depth)
			x.g.P(helper.Indent(ident+3), "foreach (var ", itemName, " in ", parentDataName, fieldName, ")")
			x.g.P(helper.Indent(ident+3), "{")
			x.g.P(helper.Indent(ident+4), "var key = ", itemName, suffix, ";")
			x.genOrderedIndexLoaderCommon(depth, ident+4, index, "key", parentDataName)
			x.g.P(helper.Indent(ident+3), "}")
		} else {
			key := parentDataName + fieldName + suffix
			x.g.P(helper.Indent(ident+3), "var key = ", key, ";")
			x.genOrderedIndexLoaderCommon(depth, ident+3, index, "key", parentDataName)
		}
	} else {
		// multi-column index
		x.generateOneMulticolumnOrderedIndex(depth, ident+2, index, parentDataName, nil)
	}
	x.g.P(helper.Indent(ident+2), "}")
}

func (x *Generator) generateOneMulticolumnOrderedIndex(depth, ident int, index *index.LevelIndex, parentDataName string, keys helper.MapKeySlice) {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		keyType := x.orderedIndexMapKeyType(index)
		x.g.P(helper.Indent(ident+1), "var key = new ", keyType, "(", keys.GenGetArguments(), ");")
		x.genOrderedIndexLoaderCommon(depth, ident+1, index, "key", parentDataName)
		return
	}
	field := index.ColFields[cursor]
	fieldName, suffix := x.parseKeyFieldNameAndSuffix(field)
	if field.FD.IsList() {
		itemName := fmt.Sprintf("indexItem%d", cursor)
		x.g.P(helper.Indent(ident+1), "foreach (var ", itemName, " in ", parentDataName, fieldName, ")")
		x.g.P(helper.Indent(ident+1), "{")
		key := itemName
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneMulticolumnOrderedIndex(depth, ident+1, index, parentDataName, keys)
		x.g.P(helper.Indent(ident+1), "}")
	} else {
		key := parentDataName + fieldName + suffix
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneMulticolumnOrderedIndex(depth, ident, index, parentDataName, keys)
	}
}

func (x *Generator) genOrderedIndexLoaderCommon(depth, ident int, index *index.LevelIndex, key, parentDataName string) {
	x.g.P(helper.Indent(ident), "{")
	x.g.P(helper.Indent(ident+1), "var list = ", x.orderedIndexContainerName(index, 0), ".TryGetValue(", key, ", out var existingList) ?")
	x.g.P(helper.Indent(ident+1), "existingList : ", x.orderedIndexContainerName(index, 0), "[", key, "] = [];")
	x.g.P(helper.Indent(ident+1), "list.Add(", parentDataName, ");")
	x.g.P(helper.Indent(ident), "}")
	for i := 1; i <= depth-2; i++ {
		if i > len(x.keys) {
			break
		}
		x.g.P(helper.Indent(ident), "{")
		if i == 1 {
			x.g.P(helper.Indent(ident+1), "var map = ", x.orderedIndexContainerName(index, i), ".TryGetValue(k1, out var existingMap) ?")
			x.g.P(helper.Indent(ident+1), "existingMap : ", x.orderedIndexContainerName(index, i), "[k1] = [];")
			x.g.P(helper.Indent(ident+1), "var list = map.TryGetValue(", key, ", out var existingList) ?")
			x.g.P(helper.Indent(ident+1), "existingList : map[", key, "] = [];")
			x.g.P(helper.Indent(ident+1), "list.Add(", parentDataName, ");")
		} else {
			var fields []string
			for j := 1; j <= i; j++ {
				fields = append(fields, fmt.Sprintf("k%d", j))
			}
			levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
			x.g.P(helper.Indent(ident+1), "var mapKey = new ", levelIndexKeyType, "(", strings.Join(fields, ", "), ");")
			x.g.P(helper.Indent(ident+1), "var map = ", x.orderedIndexContainerName(index, i), ".TryGetValue(mapKey, out var existingMap) ?")
			x.g.P(helper.Indent(ident+1), "existingMap : ", x.orderedIndexContainerName(index, i), "[mapKey] = [];")
			x.g.P(helper.Indent(ident+1), "var list = map.TryGetValue(", key, ", out var existingList) ?")
			x.g.P(helper.Indent(ident+1), "existingList : map[", key, "] = [];")
			x.g.P(helper.Indent(ident+1), "list.Add(", parentDataName, ");")
		}
		x.g.P(helper.Indent(ident), "}")
	}
}

func (x *Generator) genOrderedIndexSorter() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			if len(index.SortedColFields) != 0 {
				valueType := x.mapValueType(index)
				x.g.P(helper.Indent(3), "// OrderedIndex(sort): ", index.Index)
				indexContainerName := x.orderedIndexContainerName(index, 0)
				sorter := strings.TrimPrefix(indexContainerName, "_") + "Comparison"
				x.g.P(helper.Indent(3), "Comparison<", valueType, "> ", sorter, " = (a, b) =>")
				var keys helper.MapKeySlice
				for _, field := range index.SortedColFields {
					fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
					keys = keys.AddMapKey(helper.MapKey{Name: fieldName})
				}
				x.g.P(helper.Indent(4), "(", keys.GenCustom(func(key helper.MapKey) string { return "a" + key.Name }, ", "), ").CompareTo((", keys.GenCustom(func(key helper.MapKey) string { return "b" + key.Name }, ", "), "));")
				x.g.P(helper.Indent(3), "foreach (var itemList in ", indexContainerName, ".Values)")
				x.g.P(helper.Indent(3), "{")
				x.g.P(helper.Indent(4), "itemList.Sort(", sorter, ");")
				x.g.P(helper.Indent(3), "}")
				for i := 1; i <= levelMessage.MapDepth-2; i++ {
					if i > len(x.keys) {
						break
					}
					x.g.P(helper.Indent(3), "foreach (var itemDict in ", x.orderedIndexContainerName(index, i), ".Values)")
					x.g.P(helper.Indent(3), "{")
					x.g.P(helper.Indent(4), "foreach (var itemList in itemDict.Values)")
					x.g.P(helper.Indent(4), "{")
					x.g.P(helper.Indent(5), "itemList.Sort(", sorter, ");")
					x.g.P(helper.Indent(4), "}")
					x.g.P(helper.Indent(3), "}")
				}
			}
		}
	}
}

func (x *Generator) genOrderedIndexFinders() {
	if !x.needGenerateOrderedIndex() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			mapType := x.orderedIndexMapType(index)
			mapValueType := x.mapValueType(index)
			indexContainerName := x.orderedIndexContainerName(index, 0)

			x.g.P()
			x.g.P(helper.Indent(2), "// OrderedIndex: ", index.Index)
			x.g.P(helper.Indent(2), "public ref readonly ", mapType, " Find", index.Name(), "Map() => ref ", indexContainerName, ";")
			x.g.P()

			keyType := x.orderedIndexMapKeyType(index)
			keys := x.orderedIndexKeys(index)
			params := keys.GenGetParams()
			args := keys.GenGetArguments()
			x.g.P(helper.Indent(2), "public List<", mapValueType, ">? Find", index.Name(), "(", params, ") =>")
			if len(index.ColFields) == 1 {
				x.g.P(helper.Indent(3), indexContainerName, ".TryGetValue(", args, ", out var value) ? value : null;")
			} else {
				x.g.P(helper.Indent(3), indexContainerName, ".TryGetValue(new ", keyType, "(", args, "), out var value) ? value : null;")
			}
			x.g.P()

			x.g.P(helper.Indent(2), "public ", mapValueType, "? FindFirst", index.Name(), "(", params, ") =>")
			x.g.P(helper.Indent(3), "Find", index.Name(), "(", args, ")?.FirstOrDefault();")

			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				indexContainerName := x.orderedIndexContainerName(index, i)
				partKeys := x.keys[:i]
				partParams := partKeys.GenGetParams()
				partArgs := partKeys.GenGetArguments()
				x.g.P()
				x.g.P(helper.Indent(2), "public ", mapType, "? Find", index.Name(), "Map", i, "(", partParams, ") =>")
				if len(partKeys) == 1 {
					x.g.P(helper.Indent(3), indexContainerName, ".TryGetValue(", partArgs, ", out var value) ? value : null;")
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P(helper.Indent(3), indexContainerName, ".TryGetValue(new ", levelIndexKeyType, "(", partArgs, "), out var value) ? value : null;")
				}

				x.g.P()
				x.g.P(helper.Indent(2), "public List<", mapValueType, ">? Find", index.Name(), i, "(", partParams, ", ", params, ") =>")
				if len(index.ColFields) == 1 {
					x.g.P(helper.Indent(3), "Find", index.Name(), "Map", i, "(", partArgs, ")?.TryGetValue(", args, ", out var value) == true ? value : null;")
				} else {
					x.g.P(helper.Indent(3), "Find", index.Name(), "Map", i, "(", partArgs, ")?.TryGetValue(new ", keyType, "(", args, "), out var value) == true ? value : null;")
				}

				x.g.P()
				x.g.P(helper.Indent(2), "public ", mapValueType, "? FindFirst", index.Name(), i, "(", partParams, ", ", params, ") =>")
				x.g.P(helper.Indent(3), "Find", index.Name(), i, "(", partArgs, ", ", args, ")?.FirstOrDefault();")
			}
		}
	}
}
