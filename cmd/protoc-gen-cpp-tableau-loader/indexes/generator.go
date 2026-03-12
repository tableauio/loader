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
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
				Fd:   fd,
			})
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
	// Generate LevelIndex key structs for intermediate map levels.
	//
	// x.keys holds one entry per effective map level (only levels that lead to
	// an index are included by initLevelMessage). For a 3-level map keyed by
	// (k1, k2, k3) with an index at the deepest level, x.keys = [k1, k2, k3].
	//
	// The deepest level's key (k3) is the direct lookup key and does not need
	// a LevelIndex struct. The top-level key (k1, i.e. keys[0]) is a plain map
	// key and also does not need one. Only the intermediate levels (keys[1] …
	// keys[len-2]) require a LevelIndex struct that bundles all ancestor keys:
	//
	//   keys = [k1, k2, k3]  →  one struct for keys[1]=k2: { k1, k2 }
	//   keys = [k1,k2,k3,k4] →  structs for keys[1]=k2: {k1,k2}
	//                                        and keys[2]=k3: {k1,k2,k3}
	//
	// Hence the loop runs len(x.keys)-2 times (0 times when len ≤ 2).
	for i := 2; i < len(x.keys); i++ {
		if i == 2 {
			x.g.P()
			x.g.P(helper.Indent(1), "// LevelIndex keys.")
			x.g.P(" public:")
		}
		fd := x.keys[i-1].Fd
		keyType := x.levelKeyType(fd)
		x.g.P(helper.Indent(1), "struct ", keyType, " {")
		keys := x.keys[:i]
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
