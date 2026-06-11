package main

import (
	"text/template"

	"github.com/tableauio/loader/cmd/protoc-gen-ts-tableau-loader/helper"
	"github.com/tableauio/loader/internal/xproto"
	"google.golang.org/protobuf/compiler/protogen"
)

var tpl = template.Must(template.New("").ParseFS(efs, "embed/templates/*"))

// importEntry describes a single ES module import in the generated hub.
type importEntry struct {
	Names  []string // messager class names imported from the module
	Module string   // module specifier (without leading "./" or trailing ".js")
}

// hubData is the template data for the generated hub file.
type hubData struct {
	Imports   []importEntry
	Messagers []string
}

// generateHub generates the hub file (registry + hub manager).
func generateHub(gen *protogen.Plugin) {
	filename := "hub.pc.ts"
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, nil, g, version)

	pfs := xproto.ParseProtoFiles(gen)
	var data hubData
	for _, pf := range pfs {
		if len(pf.Messagers) == 0 {
			continue
		}
		data.Imports = append(data.Imports, importEntry{
			Names:  pf.Messagers,
			Module: pf.Name + ".pc",
		})
		data.Messagers = append(data.Messagers, pf.Messagers...)
	}

	if err := tpl.Lookup("hub.pc.ts.tpl").Execute(g, data); err != nil {
		panic(err)
	}
}
