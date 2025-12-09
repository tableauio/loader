package indexes

import (
	"fmt"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
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
	keys     helper.MapKeys
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
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
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

func (x *Generator) messagerName() string {
	return string(x.message.Desc.Name())
}

func (x *Generator) levelKeyType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("LevelIndex_%sKey", helper.ParseLeveledMapPrefix(x.message.Desc, mapFd))
}

func (x *Generator) mapValueType(index *index.LevelIndex) string {
	return helper.ParseCppClassType(index.MD)
}

func (x *Generator) fieldGetter(fd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf(".%s()", helper.ParseIndexFieldName(fd))
}

func (x *Generator) parseKeyFieldNameAndSuffix(field *index.LevelField) (string, string) {
	var fieldName, suffix string
	for i, leveledFd := range field.LeveledFDList {
		fieldName += x.fieldGetter(leveledFd)
		if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
			switch leveledFd.Message().FullName() {
			case "google.protobuf.Timestamp", "google.protobuf.Duration":
				suffix = ".seconds()"
			default:
			}
		}
	}
	return fieldName, suffix
}

func (x *Generator) GenHppIndexFinders() {
	if !x.NeedGenerate() {
		return
	}
	for i := 1; i <= x.maxDepth-3 && i <= len(x.mapFds)-1; i++ {
		if i == 1 {
			x.g.P()
			x.g.P(helper.Indent(1), "// LevelIndex keys.")
			x.g.P(" public:")
		}
		fd := x.mapFds[i]
		keyType := x.levelKeyType(fd)
		x.g.P(helper.Indent(1), "struct ", keyType, " {")
		keys := x.keys[:i+1]
		for _, key := range keys {
			x.g.P(helper.Indent(2), key.Type, " ", key.Name, ";")
		}
		x.g.P("#if __cplusplus >= 202002L")
		x.g.P(helper.Indent(2), "bool operator==(const ", keyType, "& other) const = default;")
		x.g.P("#else")
		x.g.P(helper.Indent(2), "bool operator==(const ", keyType, "& other) const {")
		x.g.P(helper.Indent(3), "return std::tie(", keys.GenGetArguments(), ") == std::tie(", keys.GenOtherArguments("other"), ");")
		x.g.P(helper.Indent(2), "}")
		x.g.P("#endif")
		x.g.P(helper.Indent(1), "};")
		x.g.P(helper.Indent(1), "struct ", keyType, "Hasher {")
		x.g.P(helper.Indent(2), "std::size_t operator()(const ", keyType, "& key) const {")
		x.g.P(helper.Indent(3), "return util::SugaredHashCombine(", keys.GenOtherArguments("key"), ");")
		x.g.P(helper.Indent(2), "}")
		x.g.P(helper.Indent(1), "};")
	}
	x.genHppIndexFinders()
	x.genHppOrderedIndexFinders()
}

func (x *Generator) GenIndexLoader() {
	x.genIndexLoader()
	x.genOrderedIndexLoader()
}

func (x *Generator) GenCppIndexFinders() {
	x.genCppIndexFinders()
	x.genCppOrderedIndexFinders()
}
