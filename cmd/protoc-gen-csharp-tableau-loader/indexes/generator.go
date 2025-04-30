package indexes

import (
	"fmt"

	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Generator struct {
	g          *protogen.GeneratedFile
	descriptor *index.IndexDescriptor
	message    *protogen.Message

	// level message
	maxDepth int
	keys     helper.MapKeySlice
	mapFds   []protoreflect.FieldDescriptor
}

func NewGenerator(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, message *protogen.Message) *Generator {
	generator := &Generator{
		g:          g,
		descriptor: descriptor,
		message:    message,
	}
	generator.initLevelMessage()
	return generator
}

func (x *Generator) initLevelMessage() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		if fd := levelMessage.FD; fd != nil && fd.IsMap() {
			x.keys = x.keys.AddMapKey(helper.MapKey{
				Type:      helper.ParseMapKeyType(fd.MapKey()),
				Name:      helper.ParseMapFieldNameAsFuncParam(fd),
				FieldName: helper.ParseMapFieldNameAsKeyStructFieldName(fd),
			})
			x.mapFds = append(x.mapFds, fd)
		}
		if len(levelMessage.Indexes) != 0 || len(levelMessage.OrderedIndexes) != 0 {
			x.maxDepth = levelMessage.MapDepth
		}
	}
}

func (x *Generator) NeedGenerate() bool {
	return x.needGenerateIndex() || x.needGenerateOrderedIndex()
}

func (x *Generator) levelKeyType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("LevelIndex_%sKey", helper.ParseLeveledMapPrefix(x.message.Desc, mapFd))
}

func (x *Generator) mapValueType(index *index.LevelIndex) string {
	return helper.ParseCsharpClassType(index.MD)
}

func (x *Generator) fieldGetter(fd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf(".%s", helper.ParseIndexFieldName(fd))
}

func (x *Generator) parseKeyFieldNameAndSuffix(field *index.LevelField) (string, string) {
	var fieldName, suffix string
	needEmptyValue := len(field.LeveledFDList) > 1
	for i, leveledFd := range field.LeveledFDList {
		if i != 0 {
			fieldName += "?"
		}
		fieldName += x.fieldGetter(leveledFd)
		if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
			switch leveledFd.Message().FullName() {
			case "google.protobuf.Timestamp", "google.protobuf.Duration":
				suffix = "?.Seconds ?? 0"
				needEmptyValue = false
			default:
			}
		}
	}
	if field.FD.IsList() {
		fieldName += " ?? Enumerable.Empty<" + helper.ParseCsharpType(field.FD) + ">()"
	} else if needEmptyValue {
		fieldName += " ?? " + helper.GetTypeEmptyValue(field.FD)
	}
	return fieldName, suffix
}

func (x *Generator) GenIndexTypeDef() {
	if !x.NeedGenerate() {
		return
	}
	for i := 1; i <= x.maxDepth-3 && i <= len(x.mapFds)-1; i++ {
		if i == 1 {
			x.g.P()
			x.g.P(helper.Indent(2), "// LevelIndex keys.")
		}
		fd := x.mapFds[i]
		keyType := x.levelKeyType(fd)
		x.g.P(helper.Indent(2), "public readonly struct ", keyType, " : IEquatable<", keyType, ">")
		x.g.P(helper.Indent(2), "{")
		keys := x.keys[:i+1]
		for _, key := range keys {
			x.g.P(helper.Indent(3), "public ", key.Type, " ", key.FieldName, " { get; }")
		}
		x.g.P()
		x.g.P(helper.Indent(3), "public ", keyType, "(", keys.GenGetParams(), ")")
		x.g.P(helper.Indent(3), "{")
		for _, key := range keys {
			x.g.P(helper.Indent(4), key.FieldName, " = ", key.Name, ";")
		}
		x.g.P(helper.Indent(3), "}")
		x.g.P()
		x.g.P(helper.Indent(3), "public bool Equals(", keyType, " other) =>")
		x.g.P(helper.Indent(4), "(", keys.GenCustom(func(key helper.MapKey) string { return key.FieldName }, ", "), ").Equals((", keys.GenCustom(func(key helper.MapKey) string { return "other." + key.FieldName }, ", "), "));")
		x.g.P()
		x.g.P(helper.Indent(3), "public override int GetHashCode() =>")
		x.g.P(helper.Indent(4), "(", keys.GenCustom(func(key helper.MapKey) string { return key.FieldName }, ", "), ").GetHashCode();")
		x.g.P(helper.Indent(2), "}")
		x.g.P()
	}
	x.genIndexTypeDef()
	x.genOrderedIndexTypeDef()
}

func (x *Generator) GenIndexLoader() {
	x.genIndexLoader()
	x.genOrderedIndexLoader()
}

func (x *Generator) GenIndexFinders() {
	x.genIndexFinders()
	x.genOrderedIndexFinders()
}
