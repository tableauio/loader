package leveledindex

import (
	"fmt"

	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Generator struct {
	gen        *protogen.Plugin
	g          *protogen.GeneratedFile
	descriptor *index.IndexDescriptor
	message    *protogen.Message

	maxDepth int
	Keys     helper.MapKeys
	MapFds   []protoreflect.FieldDescriptor
}

func NewGenerator(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, message *protogen.Message) *Generator {
	generator := &Generator{
		gen:        gen,
		g:          g,
		descriptor: descriptor,
		message:    message,
	}
	generator.init()
	return generator
}

func (x *Generator) init() {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		if fd := levelMessage.FD; fd != nil && fd.IsMap() {
			x.Keys = x.Keys.AddMapKey(helper.MapKey{
				Type:      helper.ParseMapKeyType(fd.MapKey()),
				Name:      helper.ParseMapFieldNameAsFuncParam(fd),
				FieldName: helper.ParseMapFieldNameAsKeyStructFieldName(fd),
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

func (x *Generator) messagerName() string {
	return string(x.message.Desc.Name())
}

func (x *Generator) KeyType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("%s_LeveledIndex_%sKey", x.messagerName(), helper.ParseLeveledMapPrefix(x.message.Desc, mapFd))
}

func (x *Generator) GenLeveledIndexTypeDef() {
	if !x.NeedGenerate() {
		return
	}
	for i := 1; i <= x.maxDepth-3 && i <= len(x.MapFds)-1; i++ {
		if i == 1 {
			x.g.P()
			x.g.P("// LeveledIndex keys.")
		}
		fd := x.MapFds[i]
		keyType := x.KeyType(fd)
		keys := x.Keys[:i+1]
		x.g.P("type ", keyType, " struct {")
		for _, key := range keys {
			x.g.P(key.FieldName, " ", key.Type)
		}
		x.g.P("}")
	}
}
