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

	// FD is the container field (map or list) through which this level is entered
	// from its parent. It is nil for the top-level (root) LevelMessage.
	FD protoreflect.FieldDescriptor

	// Current level message's all index fields
	Indexes, OrderedIndexes []*LevelIndex

	// Depth is the 0-based depth of message hierarchy.
	// For example, the top-level message has Depth=0, the next level has Depth=1, and so on.
	Depth int
	// MapDepth is the 0-based map depth of message hierarchy.
	// It only increments when the current level is entered via a map field (i.e., FD.IsMap()).
	// For example, the top-level message has MapDepth=0, the next level (if entered via map) has MapDepth=1, and so on.
	MapDepth int
}

func (l *LevelMessage) NeedGenIndex() bool {
	if l == nil {
		return false
	}
	return len(l.Indexes) != 0 || l.NextLevel.NeedGenIndex()
}

func (l *LevelMessage) NeedGenOrderedIndex() bool {
	if l == nil {
		return false
	}
	return len(l.OrderedIndexes) != 0 || l.NextLevel.NeedGenOrderedIndex()
}

// NeedGenAnyIndex reports whether this level or any deeper level has
// at least one index (regular or ordered).
func (l *LevelMessage) NeedGenAnyIndex() bool {
	return l.NeedGenIndex() || l.NeedGenOrderedIndex()
}

// NeedMapKeyForIndex checks if the map key variable at this level is needed
// by any deeper level's regular index's leveled containers.
// It finds the first level whose MapDepth > l.MapDepth+1 (i.e., at least 2 map
// levels deeper), then delegates to NeedGenIndex which recursively checks that
// level and all deeper levels for indexes.
func (l *LevelMessage) NeedMapKeyForIndex() bool {
	for lm := l.NextLevel; lm != nil; lm = lm.NextLevel {
		if lm.MapDepth > l.MapDepth {
			return lm.NeedGenIndex()
		}
	}
	return false
}

// NeedMapKeyForOrderedIndex checks if the map key variable at this level is
// needed by any deeper level's ordered index's leveled containers.
func (l *LevelMessage) NeedMapKeyForOrderedIndex() bool {
	for lm := l.NextLevel; lm != nil; lm = lm.NextLevel {
		if lm.MapDepth > l.MapDepth {
			return lm.NeedGenOrderedIndex()
		}
	}
	return false
}

type LevelIndex struct {
	*Index
	MD              protoreflect.MessageDescriptor
	ColFields       []*LevelField
	SortedColFields []*LevelField
	NameConflict    *Index
}

func (l *LevelIndex) Name() string {
	name := l.Index.Name
	if name == "" {
		name = string(l.MD.Name())
	}
	return name
}

func parseLevelMessage(fd protoreflect.FieldDescriptor, md protoreflect.MessageDescriptor, depth, mapDepth int) *LevelMessage {
	levelMsg := &LevelMessage{
		FD:       fd,
		Depth:    depth,
		MapDepth: mapDepth,
	}
	for i := 0; i < md.Fields().Len(); i++ {
		childFD := md.Fields().Get(i)
		if childFD.IsMap() && childFD.MapValue().Kind() == protoreflect.MessageKind {
			levelMsg.NextLevel = parseLevelMessage(childFD, childFD.MapValue().Message(), depth+1, mapDepth+1)
			return levelMsg
		} else if childFD.IsList() && childFD.Kind() == protoreflect.MessageKind {
			levelMsg.NextLevel = parseLevelMessage(childFD, childFD.Message(), depth+1, mapDepth)
			return levelMsg
		}
	}
	return levelMsg
}

// parseRecursively parses multi-column index related info.
func parseRecursively(index *Index, prefix string, md protoreflect.MessageDescriptor, lm *LevelMessage, ordered bool) {
	colFields := parseInSameLevel(index.Cols, prefix, md, nil)
	sortedColFields := parseInSameLevel(index.SortedCols, prefix, md, nil)
	if len(colFields) != 0 {
		// index belongs to current level
		levelIndex := &LevelIndex{
			Index:           index,
			MD:              md,
			ColFields:       colFields,
			SortedColFields: sortedColFields,
		}
		if ordered {
			lm.OrderedIndexes = append(lm.OrderedIndexes, levelIndex)
		} else {
			lm.Indexes = append(lm.Indexes, levelIndex)
		}
	} else if lm != nil && lm.NextLevel != nil {
		// index invalid or belongs to deeper level
		fd := lm.NextLevel.FD
		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := fdOpts.GetName()
		if fd.IsMap() {
			parseRecursively(index, prefix+fieldOptName, fd.MapValue().Message(), lm.NextLevel, ordered)
		} else {
			parseRecursively(index, prefix+fieldOptName, fd.Message(), lm.NextLevel, ordered)
		}
	}
}

func parseInSameLevel(cols []string, prefix string, md protoreflect.MessageDescriptor, leveledFDList []protoreflect.FieldDescriptor) []*LevelField {
	levelFieldMap := parseCols(cols, prefix, md, leveledFDList)
	var levelFields []*LevelField
	for _, columnName := range cols {
		if levelFieldMap[columnName] != nil {
			levelFields = append(levelFields, levelFieldMap[columnName])
		}
	}
	return levelFields
}

func parseCols(cols []string, prefix string, md protoreflect.MessageDescriptor, leveledFDList []protoreflect.FieldDescriptor) map[string]*LevelField {
	levelFields := map[string]*LevelField{} // column name -> level field
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
				levelFields[columnName] = field
				break
			} else if fd.Kind() == protoreflect.MessageKind &&
				!fd.IsMap() && !fd.IsList() &&
				strings.HasPrefix(columnName, prefix+fieldOptName) {
				subLevelFields := parseCols(
					cols, prefix+fieldOptName, fd.Message(),
					append(leveledFDList, fd),
				)
				for columnName, field := range subLevelFields {
					levelFields[columnName] = field
				}
			}
		}
	}
	return levelFields
}

func ParseIndexDescriptor(md protoreflect.MessageDescriptor) *IndexDescriptor {
	levelMessage := parseLevelMessage(nil, md, 0, 0)
	indexes, orderedIndexes := ParseWSOptionIndex(md)
	// parse indexes into level message
	for _, index := range indexes {
		if len(index.Cols) == 0 {
			continue
		}
		parseRecursively(index, "", md, levelMessage, false)
	}
	for _, index := range orderedIndexes {
		if len(index.Cols) == 0 {
			continue
		}
		parseRecursively(index, "", md, levelMessage, true)
	}
	// check duplicate index name
	indexNameMap := map[string]*Index{}
	for lm := levelMessage; lm != nil; lm = lm.NextLevel {
		allIndexes := append(lm.Indexes, lm.OrderedIndexes...)
		for _, index := range allIndexes {
			name := index.Name()
			if existingIndex, ok := indexNameMap[name]; ok {
				panic(fmt.Sprintf("duplicate index name on %v in %v: %v and %v", md.Name(), md.ParentFile().Path(), index, existingIndex))
			} else {
				indexNameMap[name] = index.Index
			}
		}
	}
	return &IndexDescriptor{
		LevelMessage: levelMessage.NextLevel,
	}
}
