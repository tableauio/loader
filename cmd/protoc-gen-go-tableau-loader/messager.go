package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	errorsPackage = protogen.GoImportPath("errors")
	fmtPackage    = protogen.GoImportPath("fmt")
	formatPackage = protogen.GoImportPath("github.com/tableauio/tableau/format")
	loadPackage   = protogen.GoImportPath("github.com/tableauio/tableau/load")
)

// golbal container for record all proto filenames and messager names
var messagers []string

// generateMessager generates a protoconf file correponsing to the protobuf file.
// Each wrapped struct type implement the Messager interface.
func generateMessager(gen *protogen.Plugin, file *protogen.File) {
	filename := filepath.Join(*pkg, file.GeneratedFilenamePrefix+"."+pcExt+".go")
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
			genMessage(gen, file, g, message)

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
		g.P(`register("`, messager, `", func() Messager {`)
		g.P("return &", messager, "{}")
		g.P("})")
	}
	g.P("}")
}

// genMessage generates a message definition.
func genMessage(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, message *protogen.Message) {
	messagerName := string(message.Desc.Name())

	// messager definition
	g.P("type ", messagerName, " struct {")
	g.P("data ", file.GoImportPath.Ident(messagerName))
	g.P("}")
	g.P()

	// messager methods
	g.P("func (x *", messagerName, ") Name() string {")
	g.P("return string((&x.data).ProtoReflect().Descriptor().Name())")
	g.P("}")
	g.P()

	g.P("func (x *", messagerName, ") Data() *", file.GoImportPath.Ident(messagerName), " {")
	g.P("return &x.data")
	g.P("}")
	g.P()

	g.P("func (x *", messagerName, ") Load(dir string, format ", formatPackage.Ident("Format"), ") error {")
	g.P("return ", loadPackage.Ident("Load"), "(&x.data, dir, format)")
	g.P("}")
	g.P()

	// syntactic sugar for accessing map items
	genMapGetters(1, nil, messagerName, file, g, message)
}

func genMapGetters(depth int, params []string, messagerName string, file *protogen.File, g *protogen.GeneratedFile, message *protogen.Message) {
	for _, field := range message.Fields {
		fd := field.Desc
		if field.Desc.IsMap() {
			keyType := parseGoType(file, fd.MapKey())
			keyName := fmt.Sprintf("key%d", depth)
			params = append(params, keyName+" "+keyType)

			g.P("func (x *", messagerName, ") Get", depth, "(", strings.Join(params, ", "), ") (*", getGoIdent(file, message, fd.MapValue()), ", error) {")

			var container string
			if depth == 1 {
				container = "x.data"
			} else {
				container = "conf"
				var findParams []string
				for i := 1; i < depth; i++ {
					findParams = append(findParams, fmt.Sprintf("key%d", i))
				}
				getter := fmt.Sprintf("Get%v", depth-1)
				g.P("conf, err := x.", getter, "(", strings.Join(findParams, ", "), ")")
				g.P("if err != nil {")
				g.P(`return nil, `, fmtPackage.Ident("Errorf"), `("`, getter, ` failed: %v", err)`)
				g.P("}")
				g.P()
			}

			g.P("d := ", container, ".", field.GoName)
			g.P("if d == nil {")
			g.P(`return nil, `, errorsPackage.Ident("New"), `("`, field.GoName, ` is nil")`)
			g.P("}")
			keyer := fmt.Sprintf("key%v", depth)
			g.P("if val, ok := d[", keyer, "]; !ok {")
			g.P(`return nil, `, fmtPackage.Ident("Errorf"), `("`, keyer, `(%v)not found", key1)`)
			g.P("} else {")
			g.P("return val, nil")
			g.P("}")
			g.P("}")
			g.P()

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				msg := getMessage(file.Messages, fd.MapValue().Message())
				if msg != nil {
					genMapGetters(depth+1, params, messagerName, file, g, msg)
				}
			}

			break
		}
	}
}

func getMessage(messages []*protogen.Message, md protoreflect.MessageDescriptor) *protogen.Message {
	if len(messages) != 0 {
		for _, msg := range messages {
			if msg.Desc.FullName() == md.FullName() {
				return msg
			} else {
				if m := getMessage(msg.Messages, md); m != nil {
					return m
				}
			}
		}
	}
	return nil
}

func getGoIdent(file *protogen.File, message *protogen.Message, fd protoreflect.FieldDescriptor) protogen.GoIdent {
	// TODO: optimize
	for _, field := range message.Fields {
		if field.Desc.FullName() == fd.FullName() {
			return field.GoIdent
		}
	}
	return protogen.GoIdent{
		GoImportPath: file.GoImportPath,
		GoName:       parseGoType(file, fd),
	}
}

// parseGoType converts a FieldDescriptor to C++ type string.
func parseGoType(file *protogen.File, fd protoreflect.FieldDescriptor) string {
	file.Desc.Package()
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	// case protoreflect.EnumKind:
	// 	protoFullName := string(fd.Message().FullName())
	// 	return strings.ReplaceAll(protoFullName, ".", "_")
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind:
		fullName := string(fd.Message().FullName())
		pkg := string(file.Desc.Package())
		protoName := fullName
		if strings.HasPrefix(fullName, pkg) {
			// defined at the same package
			protoName = fullName[len(pkg)+1:]
		}
		goName := strings.ReplaceAll(protoName, ".", "_")
		return string(file.GoImportPath.Ident(goName).GoName)
	// case protoreflect.GroupKind:
	// 	return "group"
	default:
		return fmt.Sprintf("<unknown:%d>", fd.Kind())
	}
}
