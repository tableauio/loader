package indexes

import (
	"fmt"

	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Generator struct {
	gen        *protogen.Plugin
	g          *protogen.GeneratedFile
	descriptor *index.IndexDescriptor
	message    *protogen.Message

	// level message
	maxDepth int
	keys     helper.MapKeySlice
	mapFds   []protoreflect.FieldDescriptor
}

func NewGenerator(gen *protogen.Plugin, g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, message *protogen.Message) *Generator {
	generator := &Generator{
		gen:        gen,
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

func (x *Generator) messagerName() string {
	return string(x.message.Desc.Name())
}

func (x *Generator) levelKeyType(mapFd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf("%s_LevelIndex_%sKey", x.messagerName(), helper.ParseLeveledMapPrefix(x.message.Desc, mapFd))
}

func (x *Generator) mapValueType(index *index.LevelIndex) protogen.GoIdent {
	return helper.FindMessageGoIdent(x.gen, index.MD)
}

func (x *Generator) fieldGetter(fd protoreflect.FieldDescriptor) string {
	return fmt.Sprintf(".Get%s()", helper.ParseIndexFieldName(x.gen, fd))
}

func (x *Generator) parseKeyFieldNameAndSuffix(field *index.LevelField) (string, string) {
	var fieldName, suffix string
	for i, leveledFd := range field.LeveledFDList {
		fieldName += x.fieldGetter(leveledFd)
		if i == len(field.LeveledFDList)-1 && leveledFd.Message() != nil {
			switch leveledFd.Message().FullName() {
			case "google.protobuf.Timestamp", "google.protobuf.Duration":
				suffix = ".GetSeconds()"
			default:
			}
		}
	}
	return fieldName, suffix
}

func (x *Generator) GenIndexTypeDef() {
	if !x.NeedGenerate() {
		return
	}
	// The loop generates LevelIndex key types for upper-level map containers.
	//  - i < x.maxDepth-2: maxDepth is the 0-based MapDepth of the deepest level that has indexes.
	//    LevelIndex keys start from the 2nd map level (mapFds[1]) onward, so the count is maxDepth-2.
	//    This ensures we only generate key types up to the deepest level that actually has indexes.
	//  - i < len(x.mapFds)-1: defensive bound check to prevent out-of-range access on x.mapFds[i+1].
	//    In practice, maxDepth <= len(x.mapFds) always holds, so maxDepth-2 <= len(x.mapFds)-2 < len(x.mapFds)-1.
	for i := 0; i < x.maxDepth-2 && i < len(x.mapFds)-1; i++ {
		if i == 0 {
			x.g.P()
			x.g.P("// LevelIndex keys.")
		}
		fd := x.mapFds[i+1]
		keyType := x.levelKeyType(fd)
		keys := x.keys[:i+2]
		x.g.P("type ", keyType, " struct {")
		for _, key := range keys {
			x.g.P(key.FieldName, " ", key.Type)
		}
		x.g.P("}")
	}
	x.genIndexTypeDef()
	x.genOrderedIndexTypeDef()
}

func (x *Generator) GenIndexField() {
	x.genIndexField()
	x.genOrderedIndexField()
}

func (x *Generator) GenIndexLoader() {
	x.genIndexLoader()
	x.genOrderedIndexLoader()
}

func (x *Generator) GenIndexFinders() {
	x.genIndexFinders()
	x.genOrderedIndexFinders()
}

// needMapKeyForIndex checks if the map key variable at the given mapDepth
// is needed by any subsequent regular index's leveled containers.
func (x *Generator) needMapKeyForIndex(mapDepth int) bool {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		if len(levelMessage.Indexes) > 0 && levelMessage.NumLeveledContainers() > mapDepth {
			return true
		}
	}
	return false
}

// needMapKeyForOrderedIndex checks if the map key variable at the given mapDepth
// is needed by any subsequent ordered index's leveled containers.
func (x *Generator) needMapKeyForOrderedIndex(mapDepth int) bool {
	for levelMessage := x.descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		if len(levelMessage.OrderedIndexes) > 0 && levelMessage.NumLeveledContainers() > mapDepth {
			return true
		}
	}
	return false
}
