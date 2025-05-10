package index

import (
	"fmt"
	"strings"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type IndexDescriptor struct {
	*Index

	MD   protoreflect.MessageDescriptor // deepest level message descriptor
	Name string                         // index name, e.g.: name of (ID, Name)@ItemInfo is "ItemInfo"

	Fields    []*LevelField // index fields in the same struct (protobuf message), refer to the deepest level message's Fields.
	KeyFields []*LevelField // key fields in the same struct (protobuf message), refer to the deepest level message's Fields.

	LevelMessage *LevelMessage // message hierarchy to the deepest level message which contains all index fields.
}

type LevelField struct {
	FD protoreflect.FieldDescriptor // index field descriptor

	// leveled fd list
	// For example, if you have a message described as below and created an index on "PathUserID"
	// fds are ["path", "user", "id"]
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
	LeveledFDList []protoreflect.FieldDescriptor
}

// namespaced level info
type LevelMessage struct {
	NextLevel *LevelMessage

	// Current level mesage's field which contains index fields.
	// NOTE: FD, FieldName, and FieldCard are only valid when NextLevel is not nil.
	FD protoreflect.FieldDescriptor // index field descriptor

	// Deepest level message fields corresponding to index fields
	// NOTE: Fields is valid only when this level is the deepest level.
	Fields []*LevelField

	// Deepest level message fields corresponding to key fields
	// NOTE: Fields is valid only when this level is the deepest level.
	KeyFields []*LevelField
}

// parseRecursively parses multi-column index related info.
func parseRecursively(gen *protogen.Plugin, cols, keys []string, prefix string, md protoreflect.MessageDescriptor) *LevelMessage {
	levelInfo := &LevelMessage{}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)

		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := fdOpts.GetName()
		if fd.IsMap() && fd.MapValue().Kind() == protoreflect.MessageKind {
			levelInfo.NextLevel = parseRecursively(gen, cols, keys, prefix+fieldOptName, fd.MapValue().Message())
			if levelInfo.NextLevel != nil {
				levelInfo.FD = fd
				return levelInfo
			}
		} else if fd.IsList() && fd.Kind() == protoreflect.MessageKind {
			levelInfo.NextLevel = parseRecursively(gen, cols, keys, prefix+fieldOptName, fd.Message())
			if levelInfo.NextLevel != nil {
				levelInfo.FD = fd
				return levelInfo
			}
		}
	}
	levelInfo.Fields = parseInSameLevel(gen, cols, prefix, md, nil)
	levelInfo.KeyFields = parseInSameLevel(gen, keys, prefix, md, nil)
	if len(levelInfo.Fields) != 0 {
		return levelInfo
	}
	return nil
}

func parseInSameLevel(gen *protogen.Plugin, cols []string, prefix string, md protoreflect.MessageDescriptor, leveledFDList []protoreflect.FieldDescriptor) []*LevelField {
	levelFields := []*LevelField{}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)

		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := fdOpts.GetName()

		for _, columnName := range cols {
			if prefix+fieldOptName == columnName {
				field := &LevelField{
					FD:            fd,
					LeveledFDList: append(leveledFDList, fd),
				}
				levelFields = append(levelFields, field)
				break
			} else if fd.Kind() == protoreflect.MessageKind && strings.HasPrefix(columnName, prefix+fieldOptName) {
				levelFields = append(levelFields,
					parseInSameLevel(
						gen, cols, prefix+fieldOptName, fd.Message(),
						append(leveledFDList, fd),
					)...,
				)
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
		levelInfo := parseRecursively(gen, index.Cols, index.Keys, "", md)
		if levelInfo == nil {
			continue
		}
		descriptor := &IndexDescriptor{
			Index:        index,
			LevelMessage: levelInfo,
		}
		deepestLevelMessage := descriptor.LevelMessage
		for deepestLevelMessage.NextLevel != nil {
			if fd := deepestLevelMessage.FD; fd.IsMap() {
				descriptor.MD = fd.MapValue().Message()
			} else {
				descriptor.MD = fd.Message()
			}
			deepestLevelMessage = deepestLevelMessage.NextLevel
		}
		descriptor.Fields = deepestLevelMessage.Fields
		descriptor.KeyFields = deepestLevelMessage.KeyFields
		descriptor.Name = index.Name
		if descriptor.Name == "" {
			// use index field's parent message name if not set.
			descriptor.Name = string(descriptor.MD.Name())
		}
		descriptors = append(descriptors, descriptor)
	}
	// check duplicate index name
	indexNameMap := map[string]int{}
	for i, descriptor := range descriptors {
		if j, ok := indexNameMap[descriptor.Name]; ok {
			panic(fmt.Sprintf("duplicate index name on %v: %v and %v", md.Name(), indexes[j], indexes[i]))
		} else {
			indexNameMap[descriptor.Name] = i
		}
	}
	return descriptors
}
