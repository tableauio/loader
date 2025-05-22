package main

import (
	"fmt"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
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
	messagerFullName := string(message.Desc.FullName())
	indexDescriptor := index.ParseIndexDescriptor(message.Desc)

	g.P(helper.Indent(1), "public class ", messagerName, " : Messager, IMessagerName")
	g.P(helper.Indent(1), "{")
	// type definitions
	if options.NeedGenOrderedMap(message.Desc, options.LangCS) {
		genOrderedMapTypeDef(gen, g, message.Desc, 1, nil, messagerFullName)
	}
	if options.NeedGenIndex(message.Desc, options.LangCS) {
		genIndexTypeDef(gen, g, indexDescriptor, messagerName)
	}
	g.P(helper.Indent(2), "private Protoconf.", messagerName, " Data_ = new Protoconf.", messagerName, "();")
	g.P()
	g.P(helper.Indent(2), "public static string Name() => Protoconf.", messagerName, ".Descriptor.Name;")
	g.P()
	g.P(helper.Indent(2), "public override bool Load(string dir, Format fmt, LoadOptions? options = null)")
	g.P(helper.Indent(2), "{")
	g.P(helper.Indent(3), "var start = DateTime.Now;")
	g.P(helper.Indent(3), "bool loaded = LoadMessageByPath<Protoconf.", messagerName, ">(out var msg, dir, fmt, options);")
	g.P(helper.Indent(3), "Data_ = msg;")
	g.P(helper.Indent(3), "bool ok = loaded ? ProcessAfterLoad() : false;")
	g.P(helper.Indent(3), "LoadStats.Duration = DateTime.Now - start;")
	g.P(helper.Indent(3), "return ok;")
	g.P(helper.Indent(2), "}")
	g.P()
	g.P(helper.Indent(2), "public ref readonly Protoconf.", messagerName, " Data() => ref Data_;")

	if options.NeedGenOrderedMap(message.Desc, options.LangCS) || options.NeedGenIndex(message.Desc, options.LangCS) {
		g.P()
		g.P(helper.Indent(2), "protected override bool ProcessAfterLoad()")
		g.P(helper.Indent(2), "{")
		if options.NeedGenOrderedMap(message.Desc, options.LangCS) {
			genOrderedMapLoader(gen, g, message.Desc, 1, messagerFullName)
		}
		if options.NeedGenIndex(message.Desc, options.LangCS) {
			genIndexLoader(gen, g, indexDescriptor, messagerName)
		}
		g.P(helper.Indent(3), "return true;")
		g.P(helper.Indent(2), "}")
	}

	// syntactic sugar for accessing map items
	genMapGetters(gen, g, message.Desc, 1, nil, messagerName)
	if options.NeedGenOrderedMap(message.Desc, options.LangCS) {
		genOrderedMapGetters(gen, g, message.Desc, 1, nil, messagerFullName)
	}
	if options.NeedGenIndex(message.Desc, options.LangCS) {
		genIndexFinders(gen, g, indexDescriptor, messagerName)
	}
	g.P(helper.Indent(1), "}")
}

func genMapGetters(gen *protogen.Plugin, g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys []helper.MapKey, messagerName string) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			keys = helper.AddMapKey(gen, fd, keys)
			getter := fmt.Sprintf("Get%v", depth)
			g.P()
			g.P(helper.Indent(2), "public ", parseMapValueType(fd), "? ", getter, "(", helper.GenGetParams(keys), ")")
			g.P(helper.Indent(2), "{")

			lastKeyName := keys[len(keys)-1].Name
			if depth == 1 {
				g.P(helper.Indent(3), "if (Data_.", strcase.ToCamel(string(fd.Name())), ".TryGetValue(", lastKeyName, ", out var val))")
			} else {
				prevKeys := keys[:len(keys)-1]
				prevGetter := fmt.Sprintf("Get%v", depth-1)
				g.P(helper.Indent(3), "var conf = ", prevGetter, "(", helper.GenGetArguments(prevKeys), ");")
				g.P(helper.Indent(3), "if (conf?.", strcase.ToCamel(string(fd.Name())), " != null && conf.", strcase.ToCamel(string(fd.Name())), ".TryGetValue(", lastKeyName, ", out var val))")
			}
			g.P(helper.Indent(3), "{")
			g.P(helper.Indent(4), "return val;")
			g.P(helper.Indent(3), "}")
			g.P(helper.Indent(3), "return null;")
			g.P(helper.Indent(2), "}")

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
