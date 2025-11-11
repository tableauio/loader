package main

import (
	"fmt"

	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	idx "github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/index"
	orderedindex "github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/ordered_index"
	orderedmap "github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/ordered_map"
	"github.com/tableauio/loader/internal/extensions"
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

	orderedMapGenerator := orderedmap.NewGenerator(gen, g, message)
	indexGenerator := idx.NewGenerator(gen, g, indexDescriptor, message)
	orderedIndexGenerator := orderedindex.NewGenerator(gen, g, indexDescriptor, message)

	// type definitions
	orderedMapGenerator.GenOrderedMapTypeDef()
	indexGenerator.GenIndexTypeDef()
	orderedIndexGenerator.GenOrderedIndexTypeDef()

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
	orderedMapGenerator.GenOrderedMapField()
	indexGenerator.GenIndexField()
	orderedIndexGenerator.GenOrderedIndexField()
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
	g.P("func (x *", messagerName, ") Load(dir string, format ", helper.FormatPackage.Ident("Format"), " , opts *", helper.LoadPackage.Ident("MessagerOptions"), ") error {")
	g.P("start := ", helper.TimePackage.Ident("Now"), "()")
	g.P("defer func ()  {")
	g.P("x.Stats.Duration = ", helper.TimePackage.Ident("Since"), "(start)")
	g.P("}()")
	g.P("x.data = &", message.GoIdent, "{}")
	g.P("err := ", helper.LoadPackage.Ident("LoadMessagerInDir"), "(x.data, dir, format, opts)")
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
	g.P("func (x *", messagerName, ") Store(dir string, format ", helper.FormatPackage.Ident("Format"), " , options ...", helper.StorePackage.Ident("Option"), ") error {")
	g.P("return ", helper.StorePackage.Ident("Store"), "(x.Data(), dir, format, options...)")
	g.P("}")
	g.P()

	g.P("// Message returns the ", messagerName, "'s inner message data.")
	g.P("func (x *", messagerName, ") Message() ", helper.ProtoPackage.Ident("Message"), " {")
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
		orderedMapGenerator.GenOrderedMapLoader()
		indexGenerator.GenIndexLoader()
		orderedIndexGenerator.GenOrderedIndexLoader()
		g.P("return nil")
		g.P("}")
		g.P()
	}

	// syntactic sugar for accessing map items
	genMapGetters(gen, g, message, 1, nil, messagerName)
	orderedMapGenerator.GenOrderedMapGetters()
	indexGenerator.GenIndexFinders()
	orderedIndexGenerator.GenOrderedIndexFinders()
}

func genMapGetters(gen *protogen.Plugin, g *protogen.GeneratedFile, message *protogen.Message, depth int, keys helper.MapKeys, messagerName string) {
	for _, field := range message.Fields {
		fd := field.Desc
		if fd.IsMap() {
			keys = keys.AddMapKey(helper.MapKey{
				Type: helper.ParseMapKeyType(fd.MapKey()),
				Name: helper.ParseMapFieldName(fd),
			})
			getter := fmt.Sprintf("Get%v", depth)
			g.P("// ", getter, " finds value in the ", depth, "-level map. It will return")
			g.P("// NotFound error if the key is not found.")
			g.P("func (x *", messagerName, ") ", getter, "(", keys.GenGetParams(), ") (", helper.ParseMapValueType(gen, g, fd), ", error) {")

			returnEmptyValue := helper.GetTypeEmptyValue(fd.MapValue())

			var container string
			if depth == 1 {
				container = "x.Data()"
			} else {
				container = "conf"
				prevKeys := keys[:len(keys)-1]
				prevGetter := fmt.Sprintf("Get%v", depth-1)
				g.P("conf, err := x.", prevGetter, "(", prevKeys.GenGetArguments(), ")")
				g.P("if err != nil {")
				g.P(`return `, returnEmptyValue, `, err`)
				g.P("}")
			}

			g.P("d := ", container, ".Get", field.GoName, "()")
			lastKeyName := keys[len(keys)-1].Name
			g.P("if val, ok := d[", lastKeyName, "]; !ok {")
			g.P(`return `, returnEmptyValue, `, `, helper.FmtPackage.Ident("Errorf"), `("`, lastKeyName, `(%v) %w", `, lastKeyName, `, ErrNotFound)`)
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
