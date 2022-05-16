package index

import (
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const sep = "@"

// parse worksheet option index
func parseWSOptionIndex(md protoreflect.MessageDescriptor) map[string]string {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	if wsOpts == nil {
		return nil
	}
	indexes := map[string]string{} // columnName -> indexName
	for _, index := range wsOpts.Index {
		splits := strings.SplitN(index, sep, 2)
		if len(splits) == 0 {
			continue
		}

		columnName := ""
		indexName := ""
		if len(splits) == 1 {
			columnName = splits[0]
		} else {
			columnName = splits[0]
			indexName = splits[1]
		}
		indexes[columnName] = indexName
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
)

type IndexInfo struct {
	FullClassName  string
	IndexName      string
	IndexFieldType string
	IndexFieldName string

	LevelInfo *LevelInfo
}

// namespaced level info
type LevelInfo struct {
	FD protoreflect.FieldDescriptor   // index field descriptor
	MD protoreflect.MessageDescriptor // index field's parent message descriptor, not nil if found

	FieldName string
	Type      Type

	NextLevel *LevelInfo
}

func ParseIndexLevelInfo(indexColumnName string, prefix string, md protoreflect.MessageDescriptor) *LevelInfo {
	// fmt.Println("indexColumnName: ", indexColumnName)
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)

		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		fieldOptName := ""
		if fdOpts != nil {
			fieldOptName = fdOpts.Name
		}

		levelInfo := &LevelInfo{
			FieldName: string(fd.Name()),
		}
		if fd.IsMap() && fd.MapValue().Kind() == protoreflect.MessageKind {
			levelInfo.Type = TypeMap
			levelInfo.NextLevel = ParseIndexLevelInfo(indexColumnName, prefix+fieldOptName, fd.MapValue().Message())
		} else if fd.IsList() && fd.Kind() == protoreflect.MessageKind {
			levelInfo.Type = TypeList
			levelInfo.NextLevel = ParseIndexLevelInfo(indexColumnName, prefix+fieldOptName, fd.Message())
		} else if fd.Kind() == protoreflect.MessageKind {
			levelInfo.Type = TypeStruct
			levelInfo.NextLevel = ParseIndexLevelInfo(indexColumnName, prefix+fieldOptName, fd.Message())
		} else {
			// treated as scalar type
			if prefix+fieldOptName == indexColumnName {
				levelInfo.MD = md
				levelInfo.FD = fd
				return levelInfo
			}
		}

		if levelInfo.NextLevel != nil {
			return levelInfo
		}
	}
	return nil
}

func ParseIndexInfo(md protoreflect.MessageDescriptor) []*IndexInfo {
	indexInfos := []*IndexInfo{}
	indexes := parseWSOptionIndex(md)
	for columnName, indexName := range indexes {
		levelInfo := ParseIndexLevelInfo(columnName, "", md)
		if levelInfo == nil {
			continue
		}
		indexInfo := &IndexInfo{
			LevelInfo: levelInfo,
		}
		deepestLevelInfo := indexInfo.LevelInfo
		for tempIndexInfo := indexInfo.LevelInfo; tempIndexInfo != nil; {
			deepestLevelInfo = tempIndexInfo
			tempIndexInfo = tempIndexInfo.NextLevel
		}
		indexInfo.FullClassName = helper.ParseCppClassType(deepestLevelInfo.MD)
		indexInfo.IndexFieldType = helper.ParseCppType(deepestLevelInfo.FD)
		indexInfo.IndexFieldName = string(deepestLevelInfo.FD.Name())
		if indexName == "" {
			// use index field's parent message name if not set.
			indexInfo.IndexName = string(deepestLevelInfo.MD.Name())
		}
		indexInfos = append(indexInfos, indexInfo)
	}
	return indexInfos
}
