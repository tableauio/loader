package main

import (
	"fmt"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/indexes"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/orderedmap"
	"github.com/tableauio/loader/internal/extensions"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// golbal container for record all proto filenames and messager names
var messagers []string

// generateMessager generates a protoconf file corresponding to the protobuf file.
// Each wrapped struct type implement the Messager interface.
func generateMessager(gen *protogen.Plugin, file *protogen.File) {

	filename := filepath.Join(strcase.ToCamel(file.GeneratedFilenamePrefix) + "." + extensions.PC + ".cs")
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, file, g, version)
	generateFileContent(gen, file, g)
}

// generateFileContent generates struct type definitions.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P(staticMessagerContent1)
	var fileMessagers []string
	firstMessager := true
	for _, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			if !firstMessager {
				g.P()
			}
			firstMessager = false
			genMessage(gen, g, message)

			messagerName := string(message.Desc.Name())
			fileMessagers = append(fileMessagers, messagerName)
		}
	}
	messagers = append(messagers, fileMessagers...)
	g.P(staticMessagerContent2)
}

// genMessage generates a message definition.
func genMessage(gen *protogen.Plugin, g *protogen.GeneratedFile, message *protogen.Message) {
	messagerName := string(message.Desc.Name())
	indexDescriptor := index.ParseIndexDescriptor(message.Desc)

	orderedMapGenerator := orderedmap.NewGenerator(g, message)
	indexGenerator := indexes.NewGenerator(g, indexDescriptor, message)

	g.P(helper.Indent(1), "public class ", messagerName, " : Messager, IMessagerName")
	g.P(helper.Indent(1), "{")
	// type definitions
	orderedMapGenerator.GenOrderedMapTypeDef()
	indexGenerator.GenIndexTypeDef()
	g.P(helper.Indent(2), "private Protoconf.", messagerName, " _data = new();")
	g.P()
	g.P(helper.Indent(2), "public static string Name() => Protoconf.", messagerName, ".Descriptor.Name;")
	g.P()
	g.P(helper.Indent(2), "public override bool Load(string dir, Format fmt, in Load.MessagerOptions? options = null)")
	g.P(helper.Indent(2), "{")
	g.P(helper.Indent(3), "var start = DateTime.Now;")
	g.P(helper.Indent(3), "try")
	g.P(helper.Indent(3), "{")
	g.P(helper.Indent(4), "_data = (Protoconf.", messagerName, ")(")
	g.P(helper.Indent(5), "Tableau.Load.LoadMessagerInDir(Protoconf.", messagerName, ".Descriptor, dir, fmt, options)")
	g.P(helper.Indent(5), "?? throw new InvalidOperationException()")
	g.P(helper.Indent(4), ");")
	g.P(helper.Indent(3), "}")
	g.P(helper.Indent(3), "catch (Exception)")
	g.P(helper.Indent(3), "{")
	g.P(helper.Indent(4), "return false;")
	g.P(helper.Indent(3), "}")
	g.P(helper.Indent(3), "LoadStats.Duration = DateTime.Now - start;")
	g.P(helper.Indent(3), "return ProcessAfterLoad();")
	g.P(helper.Indent(2), "}")
	g.P()
	g.P(helper.Indent(2), "public ref readonly Protoconf.", messagerName, " Data() => ref _data;")
	g.P()
	g.P(helper.Indent(2), "public override pb::IMessage? Message() => _data;")

	if orderedMapGenerator.NeedGenerate() || indexGenerator.NeedGenerate() {
		g.P()
		g.P(helper.Indent(2), "protected override bool ProcessAfterLoad()")
		g.P(helper.Indent(2), "{")
		orderedMapGenerator.GenOrderedMapLoader()
		indexGenerator.GenIndexLoader()
		g.P(helper.Indent(3), "return true;")
		g.P(helper.Indent(2), "}")
	}

	// syntactic sugar for accessing map items
	genMapGetters(gen, g, message.Desc, 1, nil, messagerName)
	orderedMapGenerator.GenOrderedMapGetters()
	indexGenerator.GenIndexFinders()
	g.P(helper.Indent(1), "}")
}

func genMapGetters(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys helper.MapKeySlice, messagerName string) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			keys = keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldNameAsFuncParam(fd),
			})
			getter := fmt.Sprintf("Get%v", depth)
			g.P()

			lastKeyName := keys[len(keys)-1].Name
			if depth == 1 {
				g.P(helper.Indent(2), "public ", helper.ParseMapValueType(fd), "? ", getter, "(", keys.GenGetParams(), ") =>")
				g.P(helper.Indent(3), "_data.", strcase.ToCamel(string(fd.Name())), "?.TryGetValue(", lastKeyName, ", out var val) == true ? val : null;")
			} else {
				prevKeys := keys[:len(keys)-1]
				prevGetter := fmt.Sprintf("Get%v", depth-1)
				g.P(helper.Indent(2), "public ", helper.ParseMapValueType(fd), "? ", getter, "(", keys.GenGetParams(), ") =>")
				g.P(helper.Indent(3), prevGetter, "(", prevKeys.GenGetArguments(), ")?.", strcase.ToCamel(string(fd.Name())), "?.TryGetValue(", lastKeyName, ", out var val) == true ? val : null;")
			}

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genMapGetters(gen, g, fd.MapValue().Message(), depth+1, keys, messagerName)
			}
			break
		}
	}
}

const staticMessagerContent1 = `using pb = global::Google.Protobuf;
namespace Tableau
{`

const staticMessagerContent2 = `}`
