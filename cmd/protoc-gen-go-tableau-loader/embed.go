package main

import (
	"embed"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/helper"
	"github.com/tableauio/loader/internal/xproto"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed embed/*
var efs embed.FS

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"toLowerCamel": strcase.ToLowerCamel,
}).ParseFS(efs, "embed/templates/*"))

type embedData struct {
	Messagers []string
	Go127     bool
}

// generateEmbed generates related registry files.
func generateEmbed(gen *protogen.Plugin) {
	entries, err := efs.ReadDir("embed/templates")
	if err != nil {
		panic(err)
	}
	data := embedData{
		Messagers: xproto.ParseProtoFiles(gen).FlatMessagers(),
		Go127:     *go127,
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		g := gen.NewGeneratedFile(strings.TrimSuffix(entry.Name(), ".tpl"), "")
		helper.GenerateCommonHeader(gen, g, version)
		g.P()
		g.P("package ", *pkg)
		g.P()
		if err := tpl.Lookup(entry.Name()).Execute(g, data); err != nil {
			panic(err)
		}
	}
}
