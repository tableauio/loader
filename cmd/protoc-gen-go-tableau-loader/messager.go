package main

import (
	"fmt"

	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/extensions"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/options"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	formatPackage  = protogen.GoImportPath("github.com/tableauio/tableau/format")
	loadPackage    = protogen.GoImportPath("github.com/tableauio/tableau/load")
	storePackage   = protogen.GoImportPath("github.com/tableauio/tableau/store")
	errors         = protogen.GoImportPath("github.com/pkg/errors")
	treeMapPackage = protogen.GoImportPath("github.com/tableauio/loader/pkg/treemap")
	pairPackage    = protogen.GoImportPath("github.com/tableauio/loader/pkg/pair")
	timePackage    = protogen.GoImportPath("time")
	sortPackage    = protogen.GoImportPath("sort")
	fmtPackage     = protogen.GoImportPath("fmt")
	protoPackage   = protogen.GoImportPath("google.golang.org/protobuf/proto")
)

// golbal container for record all proto filenames and messager names
var messagers []string

// generateMessager generates a protoconf file corresponding to the protobuf file.
// Each wrapped struct type implement the Messager interface.
func generateMessager(gen *protogen.Plugin, file *protogen.File) {
	filename := file.GeneratedFilenamePrefix + "." + extensions.PC + ".go"
	g := gen.NewGeneratedFile(filename, "")
	generateFileHeader(gen, file, g)
	g.P()
	g.P("package ", *pkg)
	g.P()
	generateFileContent(gen, file, g)
}

// generateFileContent generates struct type definitions.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	var fileMessagers []string
	for _, message := range file.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			genMessage(gen, g, message)

			messagerName := string(message.Desc.Name())
			fileMessagers = append(fileMessagers, messagerName)
		}
	}
	messagers = append(messagers, fileMessagers...)
	generateRegister(fileMessagers, g)
}

func generateRegister(messagers []string, g *protogen.GeneratedFile) {
	// register messagers
	g.P("func init() {")
	for _, messager := range messagers {
		g.P(`Register(func() Messager {`)
		g.P("return new(", messager, ")")
		g.P("})")
	}
	g.P("}")
}

// genMessage generates a message definition.
func genMessage(gen *protogen.Plugin, g *protogen.GeneratedFile, message *protogen.Message) {
	messagerName := string(message.Desc.Name())
	indexDescriptor := index.ParseIndexDescriptor(message.Desc)

	// type definitions
	if options.NeedGenOrderedMap(message.Desc, options.LangGO) {
		genOrderedMapTypeDef(gen, g, message.Desc, 1, nil, messagerName)
	}
	if options.NeedGenIndex(message.Desc, options.LangGO) {
		genIndexTypeDef(gen, g, indexDescriptor, messagerName)
	}
	if options.NeedGenOrderedIndex(message.Desc, options.LangGO) {
		genOrderedIndexTypeDef(gen, g, indexDescriptor, messagerName)
	}

	g.P("// ", messagerName, " is a wrapper around protobuf message: ", message.GoIdent, ".")
	g.P("//")
	g.P("// It is designed for three goals:")
	g.P("//")
	g.P("//  1. Easy use: simple yet powerful accessers.")
	g.P("//  2. Elegant API: concise and clean functions.")
	g.P("//  3. Extensibility: Map, OrdererdMap, Index, OrderedIndex...")
	// messager definition
	g.P("type ", messagerName, " struct {")
	g.P("UnimplementedMessager")
	g.P("data, originalData *", message.GoIdent)
	if options.NeedGenOrderedMap(message.Desc, options.LangGO) {
		genOrderedMapField(g, message.Desc)
	}
	if options.NeedGenIndex(message.Desc, options.LangGO) {
		genIndexField(g, indexDescriptor, messagerName)
	}
	if options.NeedGenOrderedIndex(message.Desc, options.LangGO) {
		genOrderedIndexField(g, indexDescriptor, messagerName)
	}
	g.P("}")
	g.P()

	// messager methods
	g.P("// Name returns the ", messagerName, "'s message name.")
	g.P("func (x *", messagerName, ") Name() string {")
	g.P("if x != nil {")
	g.P("return string(x.data.ProtoReflect().Descriptor().Name())")
	g.P("}")
	g.P(`return ""`)
	g.P("}")
	g.P()

	g.P("// Data returns the ", messagerName, "'s inner message data.")
	g.P("func (x *", messagerName, ") Data() *", message.GoIdent, " {")
	g.P("if x != nil {")
	g.P("return x.data")
	g.P("}")
	g.P(`return nil`)
	g.P("}")
	g.P()

	g.P("// Load fills ", messagerName, "'s inner message from file in the specified directory and format.")
	g.P("func (x *", messagerName, ") Load(dir string, format ", formatPackage.Ident("Format"), " , opts *", loadPackage.Ident("MessagerOptions"), ") error {")
	g.P("start := ", timePackage.Ident("Now"), "()")
	g.P("defer func ()  {")
	g.P("x.Stats.Duration = ", timePackage.Ident("Since"), "(start)")
	g.P("}()")
	g.P("x.data = &", message.GoIdent, "{}")
	g.P("err := ", loadPackage.Ident("LoadMessagerInDir"), "(x.data, dir, format, opts)")
	g.P("if err != nil {")
	g.P("return err")
	g.P("}")
	g.P("if x.backup {")
	g.P("x.originalData = proto.Clone(x.data).(*", message.GoIdent, ")")
	g.P("}")
	g.P("return x.processAfterLoad()")
	g.P("}")
	g.P()

	g.P("// Store writes ", messagerName, "'s inner message to file in the specified directory and format.")
	g.P("// Available formats: JSON, Bin, and Text.")
	g.P("func (x *", messagerName, ") Store(dir string, format ", formatPackage.Ident("Format"), " , options ...", storePackage.Ident("Option"), ") error {")
	g.P("return ", storePackage.Ident("Store"), "(x.Data(), dir, format, options...)")
	g.P("}")
	g.P()

	g.P("// Message returns the ", messagerName, "'s inner message data.")
	g.P("func (x *", messagerName, ") Message() ", protoPackage.Ident("Message"), " {")
	g.P(`return x.Data()`)
	g.P("}")
	g.P()

	g.P("// Messager returns the current messager.")
	g.P("func (x *", messagerName, ") Messager() Messager {")
	g.P("return x")
	g.P("}")
	g.P()

	g.P("// originalMessage returns the ", messagerName, "'s original inner message.")
	g.P("func (x *", messagerName, ") originalMessage() proto.Message {")
	g.P("if x != nil {")
	g.P(`return x.originalData`)
	g.P("}")
	g.P(`return nil`)
	g.P("}")
	g.P()

	if options.NeedGenOrderedMap(message.Desc, options.LangGO) || options.NeedGenIndex(message.Desc, options.LangGO) || options.NeedGenOrderedIndex(message.Desc, options.LangGO) {
		g.P("// processAfterLoad runs after this messager is loaded.")
		g.P("func (x *", messagerName, ") processAfterLoad() error {")
		if options.NeedGenOrderedMap(message.Desc, options.LangGO) {
			genOrderedMapLoader(gen, g, message.Desc, 1, nil, messagerName, "")
		}
		if options.NeedGenIndex(message.Desc, options.LangGO) {
			genIndexLoader(gen, g, indexDescriptor, messagerName)
		}
		if options.NeedGenOrderedIndex(message.Desc, options.LangGO) {
			genOrderedIndexLoader(gen, g, indexDescriptor)
		}
		g.P("return nil")
		g.P("}")
		g.P()
	}

	// syntactic sugar for accessing map items
	genMapGetters(gen, g, message, 1, nil, messagerName)
	if options.NeedGenOrderedMap(message.Desc, options.LangGO) {
		genOrderedMapGetters(gen, g, message.Desc, 1, nil, messagerName)
	}
	if options.NeedGenIndex(message.Desc, options.LangGO) {
		genIndexFinders(gen, g, indexDescriptor, messagerName)
	}
	if options.NeedGenOrderedIndex(message.Desc, options.LangGO) {
		genOrderedIndexFinders(gen, g, indexDescriptor, messagerName)
	}
}

func genMapGetters(gen *protogen.Plugin, g *protogen.GeneratedFile, message *protogen.Message, depth int, keys []helper.MapKey, messagerName string) {
	for _, field := range message.Fields {
		fd := field.Desc
		if fd.IsMap() {
			keys = helper.AddMapKey(gen, fd, keys)
			getter := fmt.Sprintf("Get%v", depth)
			g.P("// ", getter, " finds value in the ", depth, "-level map. It will return")
			g.P("// NotFound error if the key is not found.")
			g.P("func (x *", messagerName, ") ", getter, "(", helper.GenGetParams(keys), ") (", parseMapValueType(gen, g, fd), ", error) {")

			returnEmptyValue := helper.GetTypeEmptyValue(fd.MapValue())

			var container string
			if depth == 1 {
				container = "x.Data()"
			} else {
				container = "conf"
				prevKeys := keys[:len(keys)-1]
				prevGetter := fmt.Sprintf("Get%v", depth-1)
				g.P("conf, err := x.", prevGetter, "(", helper.GenGetArguments(prevKeys), ")")
				g.P("if err != nil {")
				g.P(`return `, returnEmptyValue, `, err`)
				g.P("}")
			}

			g.P("d := ", container, ".Get", field.GoName, "()")
			lastKeyName := keys[len(keys)-1].Name
			g.P("if val, ok := d[", lastKeyName, "]; !ok {")
			g.P(`return `, returnEmptyValue, `, `, fmtPackage.Ident("Errorf"), `("`, lastKeyName, `(%v) %w", `, lastKeyName, `, ErrNotFound)`)
			g.P("} else {")
			g.P(`return val, nil`)
			g.P("}")
			g.P("}")
			g.P()

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				msg := helper.FindMessage(gen, fd.MapValue().Message())
				if msg != nil {
					genMapGetters(gen, g, msg, depth+1, keys, messagerName)
				}
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

func parseMapValueType(gen *protogen.Plugin, g *protogen.GeneratedFile, fd protoreflect.FieldDescriptor) string {
	valueType := helper.ParseGoType(gen, fd.MapValue())
	var valueTypeStr string
	switch valueType := valueType.(type) {
	case string:
		valueTypeStr = valueType
	case protogen.GoIdent:
		valueTypeStr = g.QualifiedGoIdent(valueType)
	}
	if fd.MapValue().Kind() == protoreflect.MessageKind {
		return "*" + valueTypeStr
	}
	return valueTypeStr
}
