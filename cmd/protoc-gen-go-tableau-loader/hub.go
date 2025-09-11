package main

import (
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"toLowerCamel": strcase.ToLowerCamel,
}).ParseFS(efs, "embed/templates/*"))

// generateHub generates related hub files.
func generateHub(gen *protogen.Plugin) {
	entries, err := efs.ReadDir("embed/templates")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		g := gen.NewGeneratedFile(strings.TrimSuffix(entry.Name(), ".tpl"), "")
		generateCommonHeader(gen, g)
		g.P()
		g.P("package ", *pkg)
		g.P()
		if err := tpl.Lookup(entry.Name()).Execute(g, messagers); err != nil {
			panic(err)
		}
	}
}
