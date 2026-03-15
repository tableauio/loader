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
	keys helper.MapKeySlice
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
	for lm := x.descriptor.LevelMessage; lm != nil; lm = lm.NextLevel {
		if fd := lm.FD; fd != nil && fd.IsMap() {
			// Only collect map keys/fds when a deeper level has an index or ordered index,
			// because these keys are used solely for building upper-level (leveled) containers.
			if !lm.NextLevel.NeedGenAnyIndex() {
				break
			}
			x.keys = x.keys.AddMapKey(helper.MapKey{
				Type:      helper.ParseMapKeyType(fd.MapKey()),
				Name:      helper.ParseMapFieldNameAsFuncParam(fd),
				FieldName: helper.ParseMapFieldNameAsKeyStructFieldName(fd),
				Fd:        fd,
			})
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
	// Generate LevelIndex key structs for intermediate map levels.
	//
	// x.keys holds one entry per map level whose next level still needs an
	// index (populated by initLevelMessage). For a 3-level map keyed by
	// (k1, k2, k3) with an index at the deepest level, x.keys = [k1, k2, k3].
	//
	// Level containers at depth 1 are keyed by a single scalar (k1), so no
	// composite key struct is needed. The deepest level (depth = len(keys))
	// also does not need one, because its full key combination (all keys)
	// is already represented by the index's own key struct generated
	// separately. Only intermediate depths (2 ≤ depth < len(keys)) require
	// a LevelIndex struct that bundles all ancestor keys up to that depth:
	//
	//   keys = [k1, k2, k3]     → struct for depth 2: {k1, k2}
	//   keys = [k1, k2, k3, k4] → struct for depth 2: {k1, k2}
	//                              struct for depth 3: {k1, k2, k3}
	//
	// The loop starts at i=2 (depth 2) and creates a struct from keys[:i].
	// It runs len(x.keys)-2 times (0 times when len ≤ 2).
	for i := 2; i < len(x.keys); i++ {
		if i == 2 {
			x.g.P()
			x.g.P(helper.Indent(2), "// LevelIndex keys.")
		}
		fd := x.keys[i-1].Fd
		keyType := x.levelKeyType(fd)
		x.g.P(helper.Indent(2), "public readonly struct ", keyType, " : IEquatable<", keyType, ">")
		x.g.P(helper.Indent(2), "{")
		keys := x.keys[:i]
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
