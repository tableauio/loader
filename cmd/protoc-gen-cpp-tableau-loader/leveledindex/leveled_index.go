package leveledindex

import (
	"fmt"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Generator struct {
	g          *protogen.GeneratedFile
	descriptor *index.IndexDescriptor
	message    *protogen.Message

	maxDepth int
	Keys     helper.MapKeys
	MapFds   []protoreflect.FieldDescriptor
}

func NewGenerator(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, message *protogen.Message) *Generator {
	gen := &Generator{
		g:          g,
		descriptor: descriptor,
		message:    message,
	}
	gen.init()
	return gen
}

func (x *Generator) init() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		if fd := levelMessage.FD; fd != nil && fd.IsMap() {
			x.Keys = x.Keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			x.MapFds = append(x.MapFds, fd)
		}
		if len(levelMessage.Indexes) != 0 || len(levelMessage.OrderedIndexes) != 0 {
			x.maxDepth = levelMessage.Depth
		}
	}
}

func (x *Generator) NeedGenerate() bool {
	return options.NeedGenIndex(x.message.Desc, options.LangCPP) || options.NeedGenOrderedIndex(x.message.Desc, options.LangCPP)
}

func (x *Generator) KeyType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("LeveledIndex_%sKey", helper.ParseLeveledMapPrefix(x.message.Desc, mapFd))
}

func (x *Generator) GenHppLeveledIndexKeys() {
	if !x.NeedGenerate() {
		return
	}
	for i := 1; i <= x.maxDepth-3 && i <= len(x.MapFds)-1; i++ {
		if i == 1 {
			x.g.P()
			x.g.P(helper.Indent(1), "// LeveledIndex keys.")
			x.g.P(" public:")
		}
		fd := x.MapFds[i]
		keyType := x.KeyType(fd)
		x.g.P(helper.Indent(1), "struct ", keyType, " {")
		keys := x.Keys[:i+1]
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
}
