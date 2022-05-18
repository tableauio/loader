package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/check"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	formatPackage = protogen.GoImportPath("github.com/tableauio/tableau/format")
	loadPackage   = protogen.GoImportPath("github.com/tableauio/tableau/load")
	errors        = protogen.GoImportPath("github.com/pkg/errors")
)

// golbal container for record all proto filenames and messager names
var messagers []string
var errorsPackage protogen.GoImportPath
var codePackage protogen.GoImportPath

// generateMessager generates a protoconf file correponsing to the protobuf file.
// Each wrapped struct type implement the Messager interface.
func generateMessager(gen *protogen.Plugin, file *protogen.File) {
	errorsPackage = protogen.GoImportPath(string(file.GoImportPath) + "/" + *pkg + "/" + errPkg)
	codePackage = protogen.GoImportPath(string(file.GoImportPath) + "/" + *pkg + "/" + codePkg)

	filename := filepath.Join(file.GeneratedFilenamePrefix + "." + pcExt + ".go")
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

	g.P("// Messager is used to implement Checker interface.")
	g.P("func (x *", messagerName, ") Messager() Messager {")
	g.P("return x")
	g.P("}")
	g.P()

	g.P("// Check is used to implement Checker interface.")
	g.P("func (x *", messagerName, ") Check(hub *Hub) error {")

	levelInfos := check.ParseReferLevelInfo(*protoconfPkg, "", message.Desc)
	genCheckRefer(1, levelInfos, g, messagerName)

	g.P("return nil")
	g.P("}")
	g.P()

	g.P("func (x *", messagerName, ") Load(dir string, format ", formatPackage.Ident("Format"), " , options ...", loadPackage.Ident("Option"), ") error {")
	g.P("return ", loadPackage.Ident("Load"), "(&x.data, dir, format, options...)")
	g.P("}")
	g.P()

	// syntactic sugar for accessing map items
	genMapGetters(1, nil, messagerName, file, g, message)
}

func genCheckRefer(depth int, levelInfos []*check.LevelInfo, g *protogen.GeneratedFile, messagerName string) {
	if depth == 1 {
		g.P(`// refer check`)
	}
	for _, levelInfo := range levelInfos {
		accesser := levelInfo.Accesser
		itemName := fmt.Sprintf("item%d", depth)
		prevItemName := fmt.Sprintf("item%d", depth-1)
		fieldName := fmt.Sprintf("%s.%s", prevItemName, strcase.ToCamel(levelInfo.GoFieldName))
		if depth == 1 {
			if accesser != nil {
				fieldName = fmt.Sprintf("x.data.%s", strcase.ToCamel(levelInfo.GoFieldName))
			} else {
				g.P("for _, " + itemName + " := range x.data." + levelInfo.GoFieldName + "{")
			}
		} else {
			if levelInfo.FD == nil {
				g.P("for _, " + itemName + " := range " + prevItemName + "." + levelInfo.GoFieldName + "{")
			}
		}
		if accesser != nil {
			g.P(`if conf := hub.Get` + accesser.MessagerName + `(); conf != nil {`)
			g.P("    if _, ok := conf.Data()." + accesser.MapFieldName + "[" + accesser.MapKeyType + "(" + fieldName + ")]; !ok {")
			g.P(`        return `, errors.Ident("Errorf"), `("`, messagerName, ".", levelInfo.ColumnName, `(%v) not found in `, levelInfo.Refer, `", `, fieldName, `) `)
			g.P("    }")
			g.P("} else {")
			g.P(`    return `, errors.Ident("Errorf"), `("`+accesser.MessagerName+` not found")`)
			g.P("}")
		}

		genCheckRefer(depth+1, levelInfo.NextLevels, g, messagerName)
		if levelInfo.FD == nil {
			g.P("}")
		}
	}
}

func genMapGetters(depth int, params []string, messagerName string, file *protogen.File, g *protogen.GeneratedFile, message *protogen.Message) {
	for _, field := range message.Fields {
		fd := field.Desc
		if field.Desc.IsMap() {
			keyType := helper.ParseGoType(file, fd.MapKey())
			keyName := fmt.Sprintf("key%d", depth)
			params = append(params, keyName+" "+keyType)

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				g.P("func (x *", messagerName, ") Get", depth, "(", strings.Join(params, ", "), ") (*", getGoIdent(file, message, fd.MapValue()), ", error) {")
			} else {
				returnValType := helper.ParseGoType(file, fd.MapValue())
				g.P("func (x *", messagerName, ") Get", depth, "(", strings.Join(params, ", "), ") (", returnValType, ", error) {")
			}

			returnEmptyValue := helper.GetTypeEmptyValue(fd.MapValue())

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
				g.P(`return `, returnEmptyValue, `, err`)
				g.P("}")
				g.P()
			}

			g.P("d := ", container, ".", field.GoName)
			g.P("if d == nil {")
			g.P(`return `, returnEmptyValue, `, `, errorsPackage.Ident("Errorf"), `(`, codePackage.Ident("Nil"), `, "`, field.GoName, ` is nil")`)
			g.P("}")
			keyer := fmt.Sprintf("key%v", depth)
			g.P("if val, ok := d[", keyer, "]; !ok {")
			g.P(`return `, returnEmptyValue, `, `, errorsPackage.Ident("Errorf"), `(`, codePackage.Ident("NotFound"), `, "`, keyer, `(%v)not found", key1)`)
			g.P("} else {")
			g.P(`return val, nil`)
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
		GoName:       helper.ParseGoType(file, fd),
	}
}
