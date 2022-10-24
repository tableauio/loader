package index

import (
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Single-column index:
//	- ID
//	- ID@Item
//
// Multi-column index (composite index):
//  - (ID, Name)
//  - (ID, Name)@Item

const colAndNameSep = "@"
const multiColSep = ","
const multiColGroupCutset = "()"

type Index struct {
	Cols []string // single column name or multi column names
	Name string   // index name
}

// parse worksheet option index
func parseWSOptionIndex(md protoreflect.MessageDescriptor) []*Index {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	return parseIndexFrom(wsOpts.Index)
}

func parseColsFrom(multiColGroup string) []string {
	trimmedStr := strings.Trim(multiColGroup, multiColGroupCutset)
	cols := strings.Split(trimmedStr, multiColSep)
	for i, col := range cols {
		cols[i] = strings.TrimSpace(col)
	}
	return cols
}

func parseIndex(indexStr string) *Index {
	var cols []string
	var name string
	splits := strings.SplitN(indexStr, colAndNameSep, 2)
	switch len(splits) {
	case 1:
		cols = parseColsFrom(splits[0])
	case 2:
		cols = parseColsFrom(splits[0])
		name = splits[1]
	default:
		return nil
	}

	return &Index{
		Cols: cols,
		Name: name,
	}
}

func parseIndexFrom(indexList []string) []*Index {
	indexes := []*Index{}
	for _, indexStr := range indexList {
		index := parseIndex(indexStr)
		if index != nil {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func NeedGenIndex(md protoreflect.MessageDescriptor) bool {
	indexes := parseWSOptionIndex(md)
	return len(indexes) != 0
}

type Type int

const (
	TypeUnknown Type = iota
	TypeMap
	TypeList
	TypeStruct
	TypeEnum
	TypeScalar
)

type Card int

const (
	CardUnknown Card = iota
	CardMap
	CardList
)

type IndexDescriptor struct {
	FullClassName string
	Name          string        // index name
	Fields        []*LevelField // index fields in the same struct (protobuf message)

	LevelMessage *LevelMessage
}

type LevelField struct {
	FD protoreflect.FieldDescriptor // index field descriptor

	Card    Card
	Type    Type
	TypeStr string
	Name    string
}

// namespaced level info
type LevelMessage struct {
	NextLevel *LevelMessage

	// Current level's message descriptor
	MD protoreflect.MessageDescriptor

	// Current level mesage's field which contains index fields.
	// NOTE: FD, FieldName, FieldCard, and FieldType are only valid when NextLevel is not nil.
	FD        protoreflect.FieldDescriptor // index field descriptor
	FieldName string
	FieldCard Card
	FieldType Type

	// Deepest level message fields corresponding to index fields
	// NOTE: Fields is valid only when this level is the deepest level.
	Fields []*LevelField
}

// ParseIndexLevelInfo parses multi-column index related info.
func ParseIndexLevelInfo(cols []string, prefix string, md protoreflect.MessageDescriptor) *LevelMessage {
	// fmt.Println("indexColumnName: ", indexColumnName)
	levelInfo := &LevelMessage{
		MD: md,
	}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)

		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := fdOpts.GetName()
		if fd.IsMap() && fd.MapValue().Kind() == protoreflect.MessageKind {
			// assign current field name as the field name which contains index fields
			levelInfo.FD = fd
			levelInfo.FieldName = string(fd.Name())
			levelInfo.FieldType = TypeMap
			levelInfo.FieldCard = CardMap
			levelInfo.NextLevel = ParseIndexLevelInfo(cols, prefix+fieldOptName, fd.MapValue().Message())
			if levelInfo.NextLevel != nil {
				return levelInfo
			}
		} else if fd.IsList() && fd.Kind() == protoreflect.MessageKind {
			levelInfo.FD = fd
			levelInfo.FieldName = string(fd.Name())
			levelInfo.FieldType = TypeList
			levelInfo.FieldCard = CardList
			levelInfo.NextLevel = ParseIndexLevelInfo(cols, prefix+fieldOptName, fd.Message())
			if levelInfo.NextLevel != nil {
				return levelInfo
			}
		} else if fd.Kind() == protoreflect.MessageKind {
			levelInfo.FD = fd
			levelInfo.FieldName = string(fd.Name())
			levelInfo.FieldType = TypeStruct
			levelInfo.NextLevel = ParseIndexLevelInfo(cols, prefix+fieldOptName, fd.Message())
			if levelInfo.NextLevel != nil {
				return levelInfo
			}
		} else {
			found := false
			for _, columnName := range cols {
				if prefix+fieldOptName == columnName {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			// If found, and then find all other index fields in this same message
			for i := 0; i < md.Fields().Len(); i++ {
				fd := md.Fields().Get(i)

				opts := fd.Options().(*descriptorpb.FieldOptions)
				fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
				fieldOptName := fdOpts.GetName()

				for _, columnName := range cols {
					if prefix+fieldOptName == columnName {
						field := &LevelField{
							FD:      fd,
							TypeStr: helper.ParseCppType(fd),
							Name:    string(fd.Name()),
						}
						if fd.IsMap() {
							field.Card = CardMap
						} else if fd.IsList() {
							field.Card = CardList
						}
						// treated as scalar or enum type
						if fd.Kind() == protoreflect.EnumKind {
							field.Type = TypeEnum
						} else {
							field.Type = TypeScalar
						}
						levelInfo.Fields = append(levelInfo.Fields, field)
						break
					}
				}
			}

			return levelInfo
		}
	}
	return nil
}

func ParseIndexDescriptor(md protoreflect.MessageDescriptor) []*IndexDescriptor {
	indexInfos := []*IndexDescriptor{}
	indexes := parseWSOptionIndex(md)
	for _, index := range indexes {
		if len(index.Cols) == 0 {
			continue
		}
		levelInfo := ParseIndexLevelInfo(index.Cols, "", md)
		if levelInfo == nil {
			continue
		}
		descriptor := &IndexDescriptor{
			LevelMessage: levelInfo,
		}
		deepestLevelMessage := descriptor.LevelMessage
		for deepestLevelMessage.NextLevel != nil {
			deepestLevelMessage = deepestLevelMessage.NextLevel
		}
		descriptor.FullClassName = helper.ParseCppClassType(deepestLevelMessage.MD)
		descriptor.Fields = deepestLevelMessage.Fields
		// descriptor.IndexFieldType = deepestLevelMessage.Type
		// descriptor.IndexFieldCard = deepestLevelMessage.Card
		// descriptor.IndexFieldTypeStr = helper.ParseCppType(deepestLevelMessage.FD)
		// descriptor.IndexFieldName = string(deepestLevelMessage.FD.Name())
		descriptor.Name = index.Name
		if descriptor.Name == "" {
			// use index field's parent message name if not set.
			descriptor.Name = string(deepestLevelMessage.MD.Name())
		}
		indexInfos = append(indexInfos, descriptor)
	}
	return indexInfos
}
