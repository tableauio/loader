package main

import (
	"fmt"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/options"
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

	filename := filepath.Join(strcase.ToCamel(file.GeneratedFilenamePrefix) + ".cs")
	g := gen.NewGeneratedFile(filename, "")
	generateFileHeader(gen, file, g)
	generateFileContent(gen, file, g)
}

// generateFileContent generates struct type definitions.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	g.P(staticMessagerContent1)
	var fileMessagers []string
	for i, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			genMessage(gen, g, message)

			messagerName := string(message.Desc.Name())
			fileMessagers = append(fileMessagers, messagerName)
		}
		if i != len(file.Messages)-1 {
			g.P()
		}
	}
	messagers = append(messagers, fileMessagers...)
	g.P(staticMessagerContent2)
}

// genMessage generates a message definition.
func genMessage(gen *protogen.Plugin, g *protogen.GeneratedFile, message *protogen.Message) {
	messagerName := string(message.Desc.Name())
	messagerFullName := string(message.Desc.FullName())

	g.P("    public class ", messagerName, " : Messager, IMessagerName")
	g.P("    {")
	// type definitions
	if options.NeedGenOrderedMap(message.Desc, options.LangCS) {
		genOrderedMapTypeDef(gen, g, message.Desc, 1, nil, messagerFullName)
	}
	g.P("        private Protoconf.", messagerName, " Data_ = new Protoconf.", messagerName, "();")
	g.P()
	g.P("        public static string Name() => Protoconf.", messagerName, ".Descriptor.Name;")
	g.P()
	g.P("        public override bool Load(string dir, Format fmt, LoadOptions? options = null)")
	g.P("        {")
	g.P("            var start = DateTime.Now;")
	g.P("            bool loaded = LoadMessageByPath<Protoconf.", messagerName, ">(out var msg, dir, fmt, options);")
	g.P("            Data_ = msg;")
	g.P("            bool ok = loaded ? ProcessAfterLoad() : false;")
	g.P("            LoadStats.Duration = DateTime.Now - start;")
	g.P("            return ok;")
	g.P("        }")
	g.P()
	g.P("        public ref readonly Protoconf.", messagerName, " Data() => ref Data_;")

	if options.NeedGenOrderedMap(message.Desc, options.LangCS) || options.NeedGenIndex(message.Desc, options.LangCS) {
		g.P()
		g.P("        protected override bool ProcessAfterLoad()")
		g.P("        {")
		if options.NeedGenOrderedMap(message.Desc, options.LangCS) {
			genOrderedMapLoader(gen, g, message.Desc, 1, messagerFullName)
		}
		g.P("            return true;")
		g.P("        }")
	}

	// syntactic sugar for accessing map items
	genMapGetters(gen, g, message.Desc, 1, nil, messagerName)
	if options.NeedGenOrderedMap(message.Desc, options.LangCS) {
		genOrderedMapGetters(gen, g, message.Desc, 1, nil, messagerFullName)
	}
	g.P("    }")
}

func genMapGetters(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerName string) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			keys = helper.AddMapKey(gen, fd, keys)
			getter := fmt.Sprintf("Get%v", depth)
			g.P()
			g.P("        public ", parseMapValueType(fd), "? ", getter, "(", helper.GenGetParams(keys), ")")
			g.P("        {")

			lastKeyName := keys[len(keys)-1].Name
			if depth == 1 {
				g.P("            if (Data_.", strcase.ToCamel(string(fd.Name())), ".TryGetValue(", lastKeyName, ", out var val))")
			} else {
				prevKeys := keys[:len(keys)-1]
				prevGetter := fmt.Sprintf("Get%v", depth-1)
				g.P("            var conf = ", prevGetter, "(", helper.GenGetArguments(prevKeys), ");")
				g.P("            if (conf?.", strcase.ToCamel(string(fd.Name())), " != null && conf.", strcase.ToCamel(string(fd.Name())), ".TryGetValue(", lastKeyName, ", out var val))")
			}
			g.P("            {")
			g.P("                return val;")
			g.P("            }")
			g.P("            return null;")
			g.P("        }")

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				genMapGetters(gen, g, fd.MapValue().Message(), depth+1, keys, messagerName)
			}
			break
		}
	}
}

func getNextLevelMapFD(fd protoreflect.FieldDescriptor) protoreflect.FieldDescriptor {
	if fd.Kind() == protoreflect.MessageKind {
		md := fd.Message()
		for i := 0; i < md.Fields().Len(); i++ {
			fd := md.Fields().Get(i)
			if fd.IsMap() {
				return fd
			}
		}
	}
	return nil
}

func parseMapValueType(fd protoreflect.FieldDescriptor) string {
	return helper.ParseCsharpType(fd.MapValue())
}

const staticMessagerContent1 = `using System;
using System.Collections.Generic;
using Google.Protobuf;
using Google.Protobuf.Collections;

namespace Tableau
{`

const staticMessagerContent2 = `}`
