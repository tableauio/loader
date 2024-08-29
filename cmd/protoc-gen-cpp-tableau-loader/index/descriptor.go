package index

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

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
	*Index

	FullClassName string        // C++ full class name
	Name          string        // index name
	Fields        []*LevelField // index fields in the same struct (protobuf message), refer to the deepest level message's Fields.

	LevelMessage *LevelMessage // message hierarchy to the deepest level message which contains all index fields.
}

type LevelField struct {
	FD protoreflect.FieldDescriptor // index field descriptor

	Card       Card
	Type       Type
	TypeStr    string
	Names      []string // protobuf field name
	ScalarName string   // scalar name of incell-list element
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
		}
	}
	levelInfo.Fields = InternalParseIndexLevelInfo(cols, prefix, md, nil)
	if len(levelInfo.Fields) != 0 {
		return levelInfo
	}
	return nil
}

func InternalParseIndexLevelInfo(cols []string, prefix string, md protoreflect.MessageDescriptor, names []string) []*LevelField {
	levelFields := []*LevelField{}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)

		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := fdOpts.GetName()

		for _, columnName := range cols {
			if prefix+fieldOptName == columnName {
				field := &LevelField{
					FD:         fd,
					TypeStr:    helper.ParseCppType(fd),
					Names:      append(names, string(fd.Name())),
					ScalarName: string(fd.Name()),
				}
				if fd.IsMap() {
					field.Card = CardMap
				} else if fd.IsList() {
					field.Card = CardList
					// trim suffix "_list"
					// NOTE: use "name" instead list field "name_list"
					field.ScalarName = strcase.ToSnake(fieldOptName)
				}
				// treated as scalar or enum type
				if fd.Kind() == protoreflect.EnumKind {
					field.Type = TypeEnum
				} else {
					field.Type = TypeScalar
				}
				levelFields = append(levelFields, field)
				break
			} else if fd.Kind() == protoreflect.MessageKind && strings.HasPrefix(columnName, prefix+fieldOptName) {
				levelFields = append(levelFields, InternalParseIndexLevelInfo(cols, prefix+fieldOptName, fd.Message(), append(names, string(fd.Name())))...)
			}
		}
	}
	return levelFields
}

func ParseIndexDescriptor(md protoreflect.MessageDescriptor) []*IndexDescriptor {
	descriptors := []*IndexDescriptor{}
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
			Index:        index,
			LevelMessage: levelInfo,
		}
		deepestLevelMessage := descriptor.LevelMessage
		for deepestLevelMessage.NextLevel != nil {
			deepestLevelMessage = deepestLevelMessage.NextLevel
		}
		descriptor.FullClassName = helper.ParseCppClassType(deepestLevelMessage.MD)
		descriptor.Fields = deepestLevelMessage.Fields
		descriptor.Name = index.Name
		if descriptor.Name == "" {
			// use index field's parent message name if not set.
			descriptor.Name = string(deepestLevelMessage.MD.Name())
		}
		descriptors = append(descriptors, descriptor)
	}
	return descriptors
}
