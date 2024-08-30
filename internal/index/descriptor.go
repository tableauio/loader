package index

import (
	"strings"

	"github.com/iancoleman/strcase"
	cpphelper "github.com/tableauio/loader/internal/helper/cpp"
	gohelper "github.com/tableauio/loader/internal/helper/go"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Card int

const (
	CardUnknown Card = iota
	CardMap
	CardList
)

type IndexFieldName struct {
	CppName string
	GoName  string
}

type IndexDescriptor struct {
	*Index

	CppFullClassName string // C++ full class name
	GoIdent          protogen.GoIdent
	Name             string        // index name
	Fields           []*LevelField // index fields in the same struct (protobuf message), refer to the deepest level message's Fields.

	LevelMessage *LevelMessage // message hierarchy to the deepest level message which contains all index fields.
}

type LevelField struct {
	FD protoreflect.FieldDescriptor // index field descriptor

	Card       Card
	TypeStr    string
	CppTypeStr string
	GoType     any             // string or protogen.GoIdent
	ScalarName *IndexFieldName //  scalar name of incell-list element

	// leveled field names
	// For example, if you have a message described as below and created an index on "PathUserID"
	// Names are ["path", "user", "id"]
	//
	// message ItemConf {
	// 	option (tableau.worksheet) = {
	// 	  name: "ItemConf"
	// 	  index: "PathUserID"
	// 	};
	// 	map<uint32, Item> item_map = 1 [(tableau.field) = { key: "ID" layout: LAYOUT_VERTICAL }];
	// 	message Item {
	// 	  uint32 id = 1 [(tableau.field) = { name: "ID" }];
	// 	  Path path = 2 [(tableau.field) = { name: "Path" }];
	// 	  message Path {
	// 	    string dir = 1 [(tableau.field) = { name: "Dir" }];
	// 	    User user = 2 [(tableau.field) = { name: "User" }];
	// 	    message User {
	// 	      uint32 id = 1 [(tableau.field) = { name: "ID" }];
	// 	    }
	// 	  }
	// 	}
	// }
	Names []*IndexFieldName
}

// namespaced level info
type LevelMessage struct {
	NextLevel *LevelMessage

	// Current level's message descriptor
	MD protoreflect.MessageDescriptor

	// Current level mesage's field which contains index fields.
	// NOTE: FD, FieldName, and FieldCard are only valid when NextLevel is not nil.
	FD        protoreflect.FieldDescriptor // index field descriptor
	FieldName *IndexFieldName
	FieldCard Card

	// Deepest level message fields corresponding to index fields
	// NOTE: Fields is valid only when this level is the deepest level.
	Fields []*LevelField
}

// ParseRecursively parses multi-column index related info.
func ParseRecursively(gen *protogen.Plugin, cols []string, prefix string, md protoreflect.MessageDescriptor) *LevelMessage {
	levelInfo := &LevelMessage{
		MD: md,
	}
	msg := gohelper.FindMessage(gen, md)
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)

		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := fdOpts.GetName()
		if fd.IsMap() && fd.MapValue().Kind() == protoreflect.MessageKind {
			// assign current field name as the field name which contains index fields
			levelInfo.FD = fd
			levelInfo.FieldName = &IndexFieldName{
				CppName: string(fd.Name()),
				GoName:  msg.Fields[i].GoName,
			}
			levelInfo.FieldCard = CardMap
			levelInfo.NextLevel = ParseRecursively(gen, cols, prefix+fieldOptName, fd.MapValue().Message())
			if levelInfo.NextLevel != nil {
				return levelInfo
			}
		} else if fd.IsList() && fd.Kind() == protoreflect.MessageKind {
			levelInfo.FD = fd
			levelInfo.FieldName = &IndexFieldName{
				CppName: string(fd.Name()),
				GoName:  msg.Fields[i].GoName,
			}
			levelInfo.FieldCard = CardList
			levelInfo.NextLevel = ParseRecursively(gen, cols, prefix+fieldOptName, fd.Message())
			if levelInfo.NextLevel != nil {
				return levelInfo
			}
		}
	}
	levelInfo.Fields = ParseInSameLevel(gen, cols, prefix, md, nil)
	if len(levelInfo.Fields) != 0 {
		return levelInfo
	}
	return nil
}

func ParseInSameLevel(gen *protogen.Plugin, cols []string, prefix string, md protoreflect.MessageDescriptor, names []*IndexFieldName) []*LevelField {
	levelFields := []*LevelField{}
	msg := gohelper.FindMessage(gen, md)
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)

		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := fdOpts.GetName()

		for _, columnName := range cols {
			if prefix+fieldOptName == columnName {
				field := &LevelField{
					FD:         fd,
					CppTypeStr: cpphelper.ParseCppType(fd),
					GoType: func() any {
						switch fd.Kind() {
						case protoreflect.EnumKind:
							return gohelper.FindEnumGoIdent(gen, fd.Enum())
						default:
							return gohelper.ParseGoType(gen, fd)
						}
					}(),
					Names: append(names, &IndexFieldName{
						CppName: string(fd.Name()),
						GoName:  msg.Fields[i].GoName,
					}),
					ScalarName: &IndexFieldName{
						CppName: string(fd.Name()),
						GoName:  msg.Fields[i].GoName,
					},
				}
				if fd.IsMap() {
					field.Card = CardMap
				} else if fd.IsList() {
					field.Card = CardList
					// trim suffix "_list"
					// NOTE: use "name" instead list field "name_list"
					field.ScalarName = &IndexFieldName{
						CppName: strcase.ToSnake(fieldOptName),
						GoName:  strcase.ToCamel(fieldOptName),
					}

				}
				levelFields = append(levelFields, field)
				break
			} else if fd.Kind() == protoreflect.MessageKind && strings.HasPrefix(columnName, prefix+fieldOptName) {
				levelFields = append(levelFields, ParseInSameLevel(gen, cols, prefix+fieldOptName, fd.Message(), append(names, &IndexFieldName{
					CppName: string(fd.Name()),
					GoName:  msg.Fields[i].GoName,
				}))...)
			}
		}
	}
	return levelFields
}

func ParseIndexDescriptor(gen *protogen.Plugin, md protoreflect.MessageDescriptor) []*IndexDescriptor {
	descriptors := []*IndexDescriptor{}
	indexes := parseWSOptionIndex(md)
	for _, index := range indexes {
		if len(index.Cols) == 0 {
			continue
		}
		levelInfo := ParseRecursively(gen, index.Cols, "", md)
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
		descriptor.CppFullClassName = cpphelper.ParseCppClassType(deepestLevelMessage.MD)
		descriptor.GoIdent = gohelper.FindMessageGoIdent(gen, deepestLevelMessage.MD)
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
