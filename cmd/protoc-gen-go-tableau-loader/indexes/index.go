package indexes

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/loadutil"
	"github.com/tableauio/loader/internal/options"
)

func (x *Generator) needGenerateIndex() bool {
	return options.NeedGenIndex(x.message.Desc, options.LangGO)
}

func (x *Generator) indexMapType(index *index.LevelIndex) string {
	return fmt.Sprintf("%s_Index_%sMap", x.messagerName(), index.Name())
}

func (x *Generator) indexMapKeyType(index *index.LevelIndex) string {
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take first field
		return helper.ParseGoType(x.gen, x.g, field.FD)
	} else {
		// multi-column index
		return fmt.Sprintf("%s_Index_%sKey", x.messagerName(), index.Name())
	}
}

func (x *Generator) indexContainerName(index *index.LevelIndex, i int) string {
	if i == 0 {
		return fmt.Sprintf("index%sMap", strcase.ToCamel(index.Name()))
	}
	return fmt.Sprintf("index%sMap%d", strcase.ToCamel(index.Name()), i)
}

func (x *Generator) indexKeys(index *index.LevelIndex) helper.MapKeySlice {
	var keys helper.MapKeySlice
	for _, field := range index.ColFields {
		keys = keys.AddMapKey(helper.MapKey{
			Type:      helper.ParseGoType(x.gen, x.g, field.FD),
			Name:      helper.ParseIndexFieldNameAsFuncParam(x.gen, field.FD),
			FieldName: helper.ParseIndexFieldNameAsKeyStructFieldName(x.gen, field.FD),
		})
	}
	return keys
}

func (x *Generator) genIndexTypeDef() {
	if !x.needGenerateIndex() {
		return
	}
	x.g.P("// Index types.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.g.P("// Index: ", index.Index)
			if len(index.ColFields) != 1 {
				// multi-column index
				keyType := x.indexMapKeyType(index)
				keys := x.indexKeys(index)

				// generate key struct
				// KeyType must be comparable, refer https://go.dev/blog/maps
				x.g.P("type ", keyType, " struct {")
				for _, key := range keys {
					x.g.P(key.FieldName, " ", key.Type)
				}
				x.g.P("}")
			}
			x.g.P("type ", x.indexMapType(index), " = map[", x.indexMapKeyType(index), "][]*", x.mapValueType(index))
			x.g.P()
		}
	}
}

func (x *Generator) genIndexField() {
	if !x.needGenerateIndex() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.g.P(x.indexContainerName(index, 0), " ", x.indexMapType(index))
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				if i == 1 {
					x.g.P(x.indexContainerName(index, i), " map[", x.keys[0].Type, "]", x.indexMapType(index))
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P(x.indexContainerName(index, i), " map[", levelIndexKeyType, "]", x.indexMapType(index))
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
	x.g.P("// Index init.")
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.g.P("x.", x.indexContainerName(index, 0), " = make(", x.indexMapType(index), ")")
			for i := 1; i <= levelMessage.MapDepth-2; i++ {
				if i > len(x.keys) {
					break
				}
				if i == 1 {
					x.g.P("x.", x.indexContainerName(index, i), " = make(map[", x.keys[0].Type, "]", x.indexMapType(index), ")")
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P("x.", x.indexContainerName(index, i), " = make(map[", levelIndexKeyType, "]", x.indexMapType(index), ")")
				}
			}
		}
	}
	parentDataName := "x.data"
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			x.genOneIndexLoader(index, levelMessage.MapDepth, parentDataName)
		}
		keyName := fmt.Sprintf("k%d", levelMessage.MapDepth)
		valueName := fmt.Sprintf("v%d", levelMessage.Depth)
		if levelMessage.FD == nil {
			break
		}
		if !levelMessage.NextLevel.NeedGenIndex() {
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

func (x *Generator) genOneIndexLoader(index *index.LevelIndex, depth int, parentDataName string) {
	x.g.P("{")
	x.g.P("// Index: ", index.Index)
	if len(index.ColFields) == 1 {
		// single-column index
		field := index.ColFields[0] // just take the first field
		fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
		if field.FD.IsList() {
			valueName := fmt.Sprintf("v%d", depth)
			x.g.P("for _ , ", valueName, " := range ", parentDataName, fieldName, " {")
			x.g.P("key := ", valueName)
			x.genIndexLoaderCommon(depth, index, parentDataName)
			x.g.P("}")
		} else {
			x.g.P("key := ", parentDataName, fieldName)
			x.genIndexLoaderCommon(depth, index, parentDataName)
		}
	} else {
		// multi-column index
		x.generateOneMulticolumnIndex(depth, index, parentDataName, nil)
	}
	x.g.P("}")
}

func (x *Generator) generateOneMulticolumnIndex(depth int, index *index.LevelIndex, parentDataName string, keys helper.MapKeySlice) {
	cursor := len(keys)
	if cursor >= len(index.ColFields) {
		keyType := x.indexMapKeyType(index)
		x.g.P("key := ", keyType, " {", keys.GenGetArguments(), "}")
		x.genIndexLoaderCommon(depth, index, parentDataName)
		return
	}
	field := index.ColFields[cursor]
	fieldName, _ := x.parseKeyFieldNameAndSuffix(field)
	if field.FD.IsList() {
		itemName := fmt.Sprintf("indexItem%d", cursor)
		x.g.P("for _, ", itemName, " := range ", parentDataName, fieldName, " {")
		keys = keys.AddMapKey(helper.MapKey{Name: itemName})
		x.generateOneMulticolumnIndex(depth, index, parentDataName, keys)
		x.g.P("}")
	} else {
		key := parentDataName + fieldName
		keys = keys.AddMapKey(helper.MapKey{Name: key})
		x.generateOneMulticolumnIndex(depth, index, parentDataName, keys)
	}
}

func (x *Generator) genIndexLoaderCommon(depth int, index *index.LevelIndex, parentDataName string) {
	indexContainerName := x.indexContainerName(index, 0)
	x.g.P("x.", indexContainerName, "[key] = append(x.", indexContainerName, "[key], ", parentDataName, ")")
	for i := 1; i <= depth-2; i++ {
		if i > len(x.keys) {
			break
		}
		indexContainerName := x.indexContainerName(index, i)
		if i == 1 {
			x.g.P("if x.", indexContainerName, "[k1] == nil {")
			x.g.P("x.", indexContainerName, "[k1] = make(", x.indexMapType(index), ")")
			x.g.P("}")
			x.g.P("x.", indexContainerName, "[k1][key] = append(x.", indexContainerName, "[k1][key], ", parentDataName, ")")
		} else {
			var fields []string
			for j := 1; j <= i; j++ {
				fields = append(fields, fmt.Sprintf("k%d", j))
			}
			levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
			keyName := indexContainerName + "Keys"
			x.g.P(keyName, " := ", levelIndexKeyType, "{", strings.Join(fields, ", "), "}")
			x.g.P("if x.", indexContainerName, "[", keyName, "] == nil {")
			x.g.P("x.", indexContainerName, "[", keyName, "] = make(", x.indexMapType(index), ")")
			x.g.P("}")
			x.g.P("x.", indexContainerName, "[", keyName, "][key] = append(x.", indexContainerName, "[", keyName, "][key], ", parentDataName, ")")
		}
	}
}

func (x *Generator) genIndexSorter() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			if len(index.SortedColFields) != 0 {
				x.g.P("// Index(sort): ", index.Index)
				indexContainerName := x.indexContainerName(index, 0)
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
				x.g.P("for _, itemList := range x.", x.indexContainerName(index, 0), " {")
				x.g.P(helper.SortPackage.Ident("Slice"), "(itemList, ", indexContainerName, "Sorter(itemList))")
				x.g.P("}")
				for i := 1; i <= levelMessage.MapDepth-2; i++ {
					if i > len(x.keys) {
						break
					}
					x.g.P("for _, itemMap := range x.", x.indexContainerName(index, i), " {")
					x.g.P("for _, itemList := range itemMap {")
					x.g.P(helper.SortPackage.Ident("Slice"), "(itemList, ", indexContainerName, "Sorter(itemList))")
					x.g.P("}")
					x.g.P("}")
				}
			}
		}
	}
}

func (x *Generator) genIndexFinders() {
	if !x.needGenerateIndex() {
		return
	}
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			indexContainerName := x.indexContainerName(index, 0)
			messagerName := x.messagerName()
			x.g.P("// Index: ", index.Index)
			x.g.P()

			x.g.P("// Find", index.Name(), "Map finds the index: key(", index.Index, ") to value(", x.mapValueType(index), ") map.")
			x.g.P("// One key may correspond to multiple values, which are represented by a slice.")
			x.g.P("func (x *", messagerName, ") Find", index.Name(), "Map() ", x.indexMapType(index), " {")
			x.g.P("return x.", indexContainerName)
			x.g.P("}")
			x.g.P()

			keys := x.indexKeys(index)
			params := keys.GenGetParams()
			args := keys.GenGetArguments()
			x.g.P("// Find", index.Name(), " finds a slice of all values of the given key(s).")
			x.g.P("func (x *", messagerName, ") Find", index.Name(), "(", params, ") []*", x.mapValueType(index), " {")
			if len(index.ColFields) == 1 {
				x.g.P("return x.", indexContainerName, "[", args, "]")
			} else {
				x.g.P("return x.", indexContainerName, "[", x.indexMapKeyType(index), "{", args, "}]")
			}
			x.g.P("}")
			x.g.P()

			x.g.P("// FindFirst", index.Name(), " finds the first value of the given key(s),")
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
				indexContainerName := x.indexContainerName(index, i)
				partKeys := x.keys[:i]
				partParams := partKeys.GenGetParams()
				partArgs := partKeys.GenGetArguments()

				x.g.P("// Find", index.Name(), "Map", i, " finds the index: key(", index.Index, ") to value(", x.mapValueType(index), "),")
				x.g.P("// which is the upper ", loadutil.Ordinal(i), "-level map specified by (", partArgs, ").")
				x.g.P("// One key may correspond to multiple values, which are represented by a slice.")
				x.g.P("func (x *", messagerName, ") Find", index.Name(), "Map", i, "(", partParams, ") ", x.indexMapType(index), " {")
				if len(partKeys) == 1 {
					x.g.P("return x.", indexContainerName, "[", partArgs, "]")
				} else {
					levelIndexKeyType := x.levelKeyType(x.mapFds[i-1])
					x.g.P("return x.", indexContainerName, "[", levelIndexKeyType, "{", partArgs, "}]")
				}
				x.g.P("}")
				x.g.P()

				x.g.P("// Find", index.Name(), i, " finds a slice of all values of the given key(s) in the upper ", loadutil.Ordinal(i), "-level map")
				x.g.P("// specified by (", partArgs, ").")
				x.g.P("func (x *", messagerName, ") Find", index.Name(), i, "(", partParams, ", ", params, ") []*", x.mapValueType(index), " {")
				if len(index.ColFields) == 1 {
					x.g.P("return x.Find", index.Name(), "Map", i, "(", partArgs, ")[", args, "]")
				} else {
					x.g.P("return x.Find", index.Name(), "Map", i, "(", partArgs, ")[", x.indexMapKeyType(index), "{", args, "}]")
				}
				x.g.P("}")
				x.g.P()

				x.g.P("// FindFirst", index.Name(), i, " finds the first value of the given key(s) in the upper ", loadutil.Ordinal(i), "-level map")
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
