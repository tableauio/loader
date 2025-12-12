package main

import (
	"text/template"

	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/extensions"
	"google.golang.org/protobuf/compiler/protogen"
)

var tpl = template.Must(template.New("").ParseFS(efs, "embed/templates/*"))

// generateHub generates related hub files.
func generateHub(gen *protogen.Plugin) {
	filename := "Hub." + extensions.PC + ".cs"
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, nil, g, version)
	if err := tpl.Lookup(filename+".tpl").Execute(g, messagers); err != nil {
		panic(err)
	}
}
