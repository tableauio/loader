package main

import (
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/internal/extensions"
	"google.golang.org/protobuf/compiler/protogen"
)

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"toLowerCamel": strcase.ToLowerCamel,
}).ParseFS(efs, "embed/templates/*"))

// generateHub generates related hub files.
func generateHub(gen *protogen.Plugin) {
	filename := "hub." + extensions.PC + ".go"
	g := gen.NewGeneratedFile(filename, "")
	generateCommonHeader(gen, g)
	g.P()
	g.P("package ", *pkg)
	g.P()
	if err := tpl.Lookup("hub.pc.go.tpl").Execute(g, messagers); err != nil {
		panic(err)
	}
}
