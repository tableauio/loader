package indexes

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/options"
)

func (x *Generator) needGenerateOrderedIndex() bool {
	return options.NeedGenOrderedIndex(x.message.Desc, options.LangGO)
}

func (x *Generator) orderedIndexMapType(index *index.LevelIndex) string {
	return fmt.Sprintf("%s_OrderedIndex_%sMap", x.messagerName(), index.Name())
}

func (x *Generator) orderedIndexMapKeyType(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take first field
		return helper.ParseOrderedIndexKeyType(x.gen, x.g, field.FD)
	} else {
		// multi-column index
		return fmt.Sprintf("%s_OrderedIndex_%sKey", x.messagerName(), index.Name())
	}
}

func (x *Generator) orderedIndexContainerName(index *index.LevelIndex, i int) string {
	if i == 0 {
		return fmt.Sprintf("orderedIndex%sMap", strcase.ToCamel(index.Name()))
	}
	return fmt.Sprintf("orderedIndex%sMap%d", strcase.ToCamel(index.Name()), i)
}

func (x *Generator) mapCtor(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		return "New"
	} else {
		// multi-column index
		return "New2"
	}
}

func (x *Generator) orderedIndexKeys(index *index.LevelIndex) helper.MapKeys {
	var keys helper.MapKeys
	for _, field := range index.ColFields {
		keys = keys.AddMapKey(helper.MapKey{
			Type:      helper.ParseOrderedIndexKeyType(x.gen, x.g, field.FD),
			Name:      helper.ParseIndexFieldNameAsFuncParam(x.gen, field.FD),
			FieldName: helper.ParseIndexFieldNameAsKeyStructFieldName(x.gen, field.FD),
		})
	}
	return keys
}

func (x *Generator) genOrderedIndexTypeDef() {
	if !x.needGenerateOrderedIndex() {
		return
	}
	x.g.P("// OrderedIndex types.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P("// OrderedIndex: ", index.Index)
			if len(index.ColFields) != 1 {
				// multi-column index
				keyType := x.orderedIndexMapKeyType(index)
				keys := x.orderedIndexKeys(index)

				// generate key struct
				x.g.P("type ", keyType, " struct {")
				for _, key := range keys {
					x.g.P(key.FieldName, " ", key.Type)
				}
				x.g.P("}")
				x.g.P()

				// generate Less func to implement cmp.Ordered interface
				x.g.P("func (x ", keyType, ") Less(other ", keyType, ") bool {")
				for i, key := range keys {
					if i == len(keys)-1 {
						x.g.P("return x.", key.FieldName, " < other.", key.FieldName)
					} else {
						x.g.P("if x.", key.FieldName, " != other.", key.FieldName, " {")
						x.g.P("return x.", key.FieldName, " < other.", key.FieldName)
						x.g.P("}")
					}
				}
				x.g.P("}")
				x.g.P()
			}
			x.g.P("type ", x.orderedIndexMapType(index), " = ", helper.TreeMapPackage.Ident("TreeMap"), "[", x.orderedIndexMapKeyType(index), ", []*", x.mapValueType(index), "]")
			x.g.P()
		}
	}
}

func (x *Generator) genOrderedIndexField() {
	if !x.needGenerateOrderedIndex() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P(x.orderedIndexContainerName(index, 0), " *", x.orderedIndexMapType(index))
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				if i == 1 {
					x.g.P(x.orderedIndexContainerName(index, i), " map[", x.keys[0].Type, "]*", x.orderedIndexMapType(index))
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P(x.orderedIndexContainerName(index, i), " map[", levelIndexKeyType, "]*", x.orderedIndexMapType(index))
				}
			}
		}
	}
}

func (x *Generator) genOrderedIndexLoader() {
	if !x.needGenerateOrderedIndex() {
		return
	}
	defer x.genOrderedIndexSorter()
	x.g.P("// OrderedIndex init.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.g.P("x.", x.orderedIndexContainerName(index, 0), " = ", helper.TreeMapPackage.Ident(x.mapCtor(index)), "[", x.orderedIndexMapKeyType(index), ", []*", x.mapValueType(index), "]()")
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				if i == 1 {
					x.g.P("x.", x.orderedIndexContainerName(index, i), " = make(map[", x.keys[0].Type, "]*", x.orderedIndexMapType(index), ")")
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P("x.", x.orderedIndexContainerName(index, i), " = make(map[", levelIndexKeyType, "]*", x.orderedIndexMapType(index), ")")
				}
			}
		}
	}
	parentDataName := "x.data"
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			x.genOneOrderedIndexLoader(index, levelMessage.MapDepth, parentDataName)
		}
		keyName := fmt.Sprintf("k%d", levelMessage.MapDepth)
		valueName := fmt.Sprintf("v%d", levelMessage.Depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGenOrderedIndex() {
			break
		}
		if levelMessage.FD.IsMap() {
			x.g.P("for ", keyName, ", ", valueName, " := range ", parentDataName, x.fieldGetter(levelMessage.FD), " {")
			x.g.P("_ = ", keyName)
		} else {
			x.g.P("for _ , ", valueName, " := range ", parentDataName, x.fieldGetter(levelMessage.FD), " {")
		}
		parentDataName = valueName
		defer x.g.P("}")
	}
}

func (x *Generator) genOneOrderedIndexLoader(index *index.LevelIndex, depth int, parentDataName string) {
	x.g.P("{")
	x.g.P("// OrderedIndex: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		fieldName, suffix := x.parseKeyFieldNameAndSuffix(field)
		if field.FD.IsList() {
			valueName := fmt.Sprintf("v%d", depth)
			x.g.P("for _ , ", valueName, " := range ", parentDataName, fieldName, " {")
			x.g.P("key := ", valueName, suffix)
			x.genOrderedIndexLoaderCommon(depth, index, parentDataName)
			x.g.P("}")
		} else {
			x.g.P("key := ", parentDataName, fieldName, suffix)
			x.genOrderedIndexLoaderCommon(depth, index, parentDataName)
		}
	} else {
		// multi-column index
		x.generateOneMulticolumnOrderedIndex(depth, index, parentDataName, nil)
	}
	x.g.P("}")
}

func (x *Generator) generateOneMulticolumnOrderedIndex(depth int, index *index.LevelIndex, parentDataName string, keys helper.MapKeys) {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		keyType := x.orderedIndexMapKeyType(index)
		x.g.P("key := ", keyType, " {", keys.GenGetArguments(), "}")
		x.genOrderedIndexLoaderCommon(depth, index, parentDataName)
		return
	}
	field := index.ColFields[cursor]
	fieldName, suffix := x.parseKeyFieldNameAndSuffix(field)
	if field.FD.IsList() {
		itemName := fmt.Sprintf("indexItem%d", cursor)
		x.g.P("for _, ", itemName, " := range ", parentDataName, fieldName, " {")
		key := itemName + suffix
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneMulticolumnOrderedIndex(depth, index, parentDataName, keys)
		x.g.P("}")
	} else {
		key := parentDataName + fieldName + suffix
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneMulticolumnOrderedIndex(depth, index, parentDataName, keys)
	}
}

func (x *Generator) genOrderedIndexLoaderCommon(depth int, index *index.LevelIndex, parentDataName string) {
	indexContainerName := x.orderedIndexContainerName(index, 0)
	x.g.P("value, _ := x.", indexContainerName, ".Get(key)")
	x.g.P("x.", indexContainerName, ".Put(key, append(value, ", parentDataName, "))")
	for i := 1; i <= depth-2; i++ {
		if i > len(x.keys) {
			break
		}
		orderedIndexContainerName := x.orderedIndexContainerName(index, i)
		valueName := orderedIndexContainerName + "Value"
		if i == 1 {
			x.g.P("if x.", orderedIndexContainerName, "[k1] == nil {")
			x.g.P("x.", orderedIndexContainerName, "[k1] = ", helper.TreeMapPackage.Ident(x.mapCtor(index)), "[", x.orderedIndexMapKeyType(index), ", []*", x.mapValueType(index), "]()")
			x.g.P("}")
			x.g.P(valueName, ", _ := x.", orderedIndexContainerName, "[k1].Get(key)")
			x.g.P("x.", orderedIndexContainerName, "[k1].Put(key, append(", valueName, ", ", parentDataName, "))")
		} else {
			var fields []string
			for j := 1; j <= i; j++ {
				fields = append(fields, fmt.Sprintf("k%d", j))
			}
			levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
			keyName := orderedIndexContainerName + "Keys"
			x.g.P(keyName, " := ", levelIndexKeyType, "{", strings.Join(fields, ", "), "}")
			x.g.P("if x.", orderedIndexContainerName, "[", keyName, "] == nil {")
			x.g.P("x.", orderedIndexContainerName, "[", keyName, "] = ", helper.TreeMapPackage.Ident(x.mapCtor(index)), "[", x.orderedIndexMapKeyType(index), ", []*", x.mapValueType(index), "]()")
			x.g.P("}")
			x.g.P(valueName, ", _ := x.", orderedIndexContainerName, "[", keyName, "].Get(key)")
			x.g.P("x.", orderedIndexContainerName, "[", keyName, "].Put(key, append(", valueName, ", ", parentDataName, "))")
		}
	}
}

func (x *Generator) genOrderedIndexSorter() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.OrderedIndexes {
			if len(index.SortedColFields) != 0 {
				x.g.P("// OrderedIndex(sort): ", index.Index)
				indexContainerName := x.orderedIndexContainerName(index, 0)
				x.g.P(indexContainerName, "Sorter := func(itemList []*", x.mapValueType(index), ") func(i, j int) bool {")
				x.g.P("return func(i, j int) bool {")
				for i, field := range index.SortedColFields {
					fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
					if i == len(index.SortedColFields)-1 {
						x.g.P("return itemList[i]", fieldName, " < itemList[j]", fieldName)
					} else {
						x.g.P("if itemList[i]", fieldName, " != itemList[j]", fieldName, " {")
						x.g.P("return itemList[i]", fieldName, " < itemList[j]", fieldName)
						x.g.P("}")
					}
				}
				x.g.P("}")
				x.g.P("}")
				x.g.P("x.", x.orderedIndexContainerName(index, 0), ".Range(func(key ", x.orderedIndexMapKeyType(index), ", itemList []*", x.mapValueType(index), ") bool {")
				x.g.P(helper.SortPackage.Ident("Slice"), "(itemList, ", indexContainerName, "Sorter(itemList))")
				x.g.P("return true")
				x.g.P("})")
				for i := 1; i <= levelMessage.MapDepth-2; i++ {
					if i > len(x.keys) {
						break
					}
					x.g.P("for _, itemMap := range x.", x.orderedIndexContainerName(index, i), " {")
					x.g.P("itemMap.Range(func(key ", x.orderedIndexMapKeyType(index), ", itemList []*", x.mapValueType(index), ") bool {")
					x.g.P(helper.SortPackage.Ident("Slice"), "(itemList, ", indexContainerName, "Sorter(itemList))")
					x.g.P("return true")
					x.g.P("})")
					x.g.P("}")
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
			indexContainerName := x.orderedIndexContainerName(index, 0)
			messagerName := x.messagerName()
			x.g.P("// OrderedIndex: ", index.Index)
			x.g.P()

			x.g.P("// Find", index.Name(), "Map finds the ordered index (", index.Index, ") to value (", x.mapValueType(index), ") treemap.")
			x.g.P("// One key may correspond to multiple values, which are contained by a slice.")
			x.g.P("func (x *", messagerName, ") Find", index.Name(), "Map() *", x.orderedIndexMapType(index), " {")
			x.g.P("return x.", indexContainerName)
			x.g.P("}")
			x.g.P()

			keys := x.orderedIndexKeys(index)
			params := keys.GenGetParams()
			args := keys.GenGetArguments()
			x.g.P("// Find", index.Name(), " finds a slice of all values of the given key.")
			x.g.P("func (x *", messagerName, ") Find", index.Name(), "(", params, ") []*", x.mapValueType(index), " {")
			if len(index.ColFields) == 1 {
				x.g.P("val, _ := x.", indexContainerName, ".Get(", args, ")")
			} else {
				x.g.P("val, _ := x.", indexContainerName, ".Get(", x.orderedIndexMapKeyType(index), "{", args, "})")
			}
			x.g.P("return val")
			x.g.P("}")
			x.g.P()

			x.g.P("// FindFirst", index.Name(), " finds the first value of the given key,")
			x.g.P("// or nil if no value found.")
			x.g.P("func (x *", messagerName, ") FindFirst", index.Name(), "(", params, ") *", x.mapValueType(index), " {")
			x.g.P("val := x.Find", index.Name(), "(", args, ")")
			x.g.P("if len(val) > 0 {")
			x.g.P("return val[0]")
			x.g.P("}")
			x.g.P("return nil")
			x.g.P("}")
			x.g.P()

			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				orderedIndexContainerName := x.orderedIndexContainerName(index, i)
				partKeys := x.keys[:i]
				partParams := partKeys.GenGetParams()
				partArgs := partKeys.GenGetArguments()

				x.g.P("// Find", index.Name(), "Map", i, " finds the index (", index.Index, ") to value (", x.mapValueType(index), ") ", i, "-level treemap")
				x.g.P("// specified by (", partArgs, ").")
				x.g.P("// One key may correspond to multiple values, which are contained by a slice.")
				x.g.P("func (x *", messagerName, ") Find", index.Name(), "Map", i, "(", partParams, ") *", x.orderedIndexMapType(index), " {")
				if len(partKeys) == 1 {
					x.g.P("return x.", orderedIndexContainerName, "[", partArgs, "]")
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P("return x.", orderedIndexContainerName, "[", levelIndexKeyType, "{", partArgs, "}]")
				}
				x.g.P("}")
				x.g.P()

				x.g.P("// Find", index.Name(), i, " finds a slice of all values of the given key in the ", i, "-level treemap")
				x.g.P("// specified by (", partArgs, ").")
				x.g.P("func (x *", messagerName, ") Find", index.Name(), i, "(", partParams, ", ", params, ") []*", x.mapValueType(index), " {")
				if len(index.ColFields) == 1 {
					x.g.P("val, _ := x.Find", index.Name(), "Map", i, "(", partArgs, ").Get(", args, ")")
				} else {
					x.g.P("val, _ := x.Find", index.Name(), "Map", i, "(", partArgs, ").Get(", x.orderedIndexMapKeyType(index), "{", args, "})")
				}
				x.g.P("return val")
				x.g.P("}")
				x.g.P()

				x.g.P("// FindFirst", index.Name(), i, " finds the first value of the given key in the ", i, "-level treemap")
				x.g.P("// specified by (", partArgs, "), or nil if no value found.")
				x.g.P("func (x *", messagerName, ") FindFirst", index.Name(), i, "(", partParams, ", ", params, ") *", x.mapValueType(index), " {")
				x.g.P("val := x.Find", index.Name(), i, "(", partArgs, ", ", args, ")")
				x.g.P("if len(val) > 0 {")
				x.g.P("return val[0]")
				x.g.P("}")
				x.g.P("return nil")
				x.g.P("}")
				x.g.P()
			}
		}
	}
}
