package check

import (
	"log"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/firstpass"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const referSep = "."

// namespaced level info
type LevelInfo struct {
	GoFieldName string

	FD protoreflect.FieldDescriptor   // index field descriptor
	MD protoreflect.MessageDescriptor // index field's parent message descriptor, not nil if found

	Accesser *ReferedAccesserInfo

	Refer      string
	ColumnName string

	NextLevels []*LevelInfo
}

func ParseReferLevelInfo(protoconfPkg string, prefix string, md protoreflect.MessageDescriptor) []*LevelInfo {
	var levelInfos []*LevelInfo
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		opts := fd.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		if fdOpts == nil {
			continue
		}
		levelInfo := &LevelInfo{
			GoFieldName: strcase.ToCamel(string(fd.Name())),
		}
		if fd.IsMap() && fd.MapValue().Kind() == protoreflect.MessageKind {
			levelInfo.NextLevels = ParseReferLevelInfo(protoconfPkg, prefix+fdOpts.Name, fd.MapValue().Message())
			if len(levelInfo.NextLevels) != 0 {
				levelInfos = append(levelInfos, levelInfo)
			}
		} else if fd.IsList() && fd.Kind() == protoreflect.MessageKind {
			levelInfo.NextLevels = ParseReferLevelInfo(protoconfPkg, prefix+fdOpts.Name, fd.Message())
			if len(levelInfo.NextLevels) != 0 {
				levelInfos = append(levelInfos, levelInfo)
			}
		} else if fd.Kind() == protoreflect.MessageKind {
			levelInfo.NextLevels = ParseReferLevelInfo(protoconfPkg, prefix+fdOpts.Name, fd.Message())
			if len(levelInfo.NextLevels) != 0 {
				levelInfos = append(levelInfos, levelInfo)
			}
		} else {
			// treated as scalar type
			if fdOpts.Prop == nil || fdOpts.Prop.Refer == "" {
				continue
			}
			splits := strings.SplitN(fdOpts.Prop.Refer, referSep, 2)
			if len(splits) != 2 {
				log.Panicf("%v.%v prop.refer is illegal, pattern should be: <SheetName>.<ColumnName>", md.FullName(), prefix+fdOpts.Name)
			}
			msgerName, columnName := splits[0], splits[1]
			fullMsgName := protoconfPkg + "." + msgerName
			msg, ok := firstpass.AllTopMessages[fullMsgName]
			if !ok {
				log.Panicf("refer: %v, message %s not found", fdOpts.Prop.Refer, fullMsgName)
			}
			// log.Println("fuck ", md.FullName(), fdOpts.Prop.Refer)
			levelInfo.Refer = fdOpts.Prop.Refer
			levelInfo.ColumnName = prefix + fdOpts.Name
			levelInfo.MD = md
			levelInfo.FD = fd
			levelInfo.Accesser = ParseReferedMapAccesserInfo(columnName, "", msg)
			if levelInfo.Accesser == nil {
				log.Panicf("refer: %v, %s is not the key of map in %s", fdOpts.Prop.Refer, columnName, msgerName)
			}
			levelInfos = append(levelInfos, levelInfo)
		}
	}
	return levelInfos
}

type ReferedAccesserInfo struct {
	MessagerName string
	MapFieldName string
	MapKeyType   string
}

func ParseReferedMapAccesserInfo(columnName string, prefix string, msg *firstpass.MessageInfo) *ReferedAccesserInfo {
	for _, field := range msg.Msg.Fields {
		opts := field.Desc.Options().(*descriptorpb.FieldOptions)
		fdOpts := proto.GetExtension(opts, tableaupb.E_Field).(*tableaupb.FieldOptions)
		if fdOpts == nil {
			continue
		}
		if field.Desc.IsMap() && prefix+fdOpts.Name+fdOpts.Key == columnName {
			return &ReferedAccesserInfo{
				MessagerName: msg.Msg.GoIdent.GoName,
				MapFieldName: field.GoName,
				MapKeyType:   helper.ParseGoType(msg.File, field.Desc.MapKey()),
			}
		}

	}
	return nil
}

// func ParseIndexInfo(md protoreflect.MessageDescriptor) []*IndexInfo {
// 	indexInfos := []*IndexInfo{}
// 	indexes := parseWSOptionIndex(md)
// 	for columnName, indexName := range indexes {
// 		levelInfo := ParseIndexLevelInfo(columnName, "", md)
// 		if levelInfo == nil {
// 			continue
// 		}
// 		indexInfo := &IndexInfo{
// 			LevelInfo: levelInfo,
// 		}
// 		deepestLevelInfo := indexInfo.LevelInfo
// 		for tempIndexInfo := indexInfo.LevelInfo; tempIndexInfo != nil; {
// 			deepestLevelInfo = tempIndexInfo
// 			tempIndexInfo = tempIndexInfo.NextLevel
// 		}
// 		indexInfo.FullClassName = helper.ParseCppClassType(deepestLevelInfo.MD)
// 		indexInfo.IndexFieldType = helper.ParseCppType(deepestLevelInfo.FD)
// 		indexInfo.IndexFieldName = string(deepestLevelInfo.FD.Name())
// 		indexInfo.IndexName = indexName
// 		if indexInfo.IndexName == "" {
// 			// use index field's parent message name if not set.
// 			indexInfo.IndexName = string(deepestLevelInfo.MD.Name())
// 		}
// 		indexInfos = append(indexInfos, indexInfo)
// 	}
// 	return indexInfos
// }
