package index

import (
	"fmt"
	"strings"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type IndexDescriptor struct {
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

	// Current level message's field which contains index fields.
	// Only valid when NextLevel is not nil.
	FD protoreflect.FieldDescriptor

	// Current level message's all index fields
	Indexes []*LevelIndex
}

func (l *LevelMessage) NeedGen() bool {
	if l == nil {
		return false
	}
	return len(l.Indexes) != 0 || l.NextLevel.NeedGen()
}

type LevelIndex struct {
	*Index
	MD        protoreflect.MessageDescriptor
	ColFields []*LevelField
	KeyFields []*LevelField
}

func (l *LevelIndex) Name() string {
	name := l.Index.Name
	if name == "" {
		name = string(l.MD.Name())
	}
	return name
}

func parseLevelMessage(md protoreflect.MessageDescriptor) *LevelMessage {
	levelInfo := &LevelMessage{}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() && fd.MapValue().Kind() == protoreflect.MessageKind {
			levelInfo.NextLevel = parseLevelMessage(fd.MapValue().Message())
			levelInfo.FD = fd
			return levelInfo
		} else if fd.IsList() && fd.Kind() == protoreflect.MessageKind {
			levelInfo.NextLevel = parseLevelMessage(fd.Message())
			levelInfo.FD = fd
			return levelInfo
		}
	}
	return &LevelMessage{}
}

// parseRecursively parses multi-column index related info.
func parseRecursively(index *Index, prefix string, md protoreflect.MessageDescriptor, levelMessage *LevelMessage) {
	colFields := parseInSameLevel(index.Cols, prefix, md, nil)
	keyFields := parseInSameLevel(index.Keys, prefix, md, nil)
	if len(colFields) != 0 {
		// index belongs to current level
		levelMessage.Indexes = append(levelMessage.Indexes, &LevelIndex{
			Index:     index,
			MD:        md,
			ColFields: colFields,
			KeyFields: keyFields,
		})
	} else if levelMessage != nil && levelMessage.NextLevel != nil {
		// index invalid or belongs to deeper level
		fd := levelMessage.FD
		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := fdOpts.GetName()
		if fd.IsMap() {
			parseRecursively(index, prefix+fieldOptName, fd.MapValue().Message(), levelMessage.NextLevel)
		} else {
			parseRecursively(index, prefix+fieldOptName, fd.Message(), levelMessage.NextLevel)
		}
	}
}

func parseInSameLevel(cols []string, prefix string, md protoreflect.MessageDescriptor, leveledFDList []protoreflect.FieldDescriptor) []*LevelField {
	var levelFields []*LevelField
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
			} else if fd.Kind() == protoreflect.MessageKind &&
				!fd.IsMap() && !fd.IsList() &&
				strings.HasPrefix(columnName, prefix+fieldOptName) {
				levelFields = append(levelFields,
					parseInSameLevel(
						cols, prefix+fieldOptName, fd.Message(),
						append(leveledFDList, fd),
					)...,
				)
			}
		}
	}
	return levelFields
}

func ParseIndexDescriptor(md protoreflect.MessageDescriptor) *IndexDescriptor {
	descriptor := &IndexDescriptor{
		LevelMessage: parseLevelMessage(md),
	}
	indexes := ParseWSOptionIndex(md)
	// parse indexes into level message
	for _, index := range indexes {
		if len(index.Cols) == 0 {
			continue
		}
		parseRecursively(index, "", md, descriptor.LevelMessage)
	}
	// check duplicate index name
	indexNameMap := map[string]*Index{}
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		for _, index := range levelMessage.Indexes {
			name := index.Name()
			if existingIndex, ok := indexNameMap[name]; ok {
				panic(fmt.Sprintf("duplicate index name on %v in %v: %v and %v", md.Name(), md.ParentFile().Path(), index, existingIndex))
			} else {
				indexNameMap[name] = index.Index
			}
		}
	}
	return descriptor
}
