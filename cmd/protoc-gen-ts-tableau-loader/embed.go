package main

import (
	"embed"
	"path"
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-ts-tableau-loader/helper"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed embed/*
var efs embed.FS

// generateEmbed generates the runtime library files (util/load/messager/...)
// by copying the embedded TypeScript sources verbatim. The templates
// subdirectory is skipped since it is consumed by the hub generator.
func generateEmbed(gen *protogen.Plugin) {
	entries, err := efs.ReadDir("embed")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		g := gen.NewGeneratedFile(entry.Name(), "")
		helper.GenerateFileHeader(gen, nil, g, version)
		// refer: [embed: embed path on different OS cannot open file](https://github.com/golang/go/issues/45230)
		content, err := efs.ReadFile(path.Join("embed", entry.Name()))
		if err != nil {
			panic(err)
		}
		// Rewrite the tableau extension descriptor import to honor pb_path: the
		// runtime sources are authored assuming the base output is one level up
		// ("../tableau/protobuf/..."); rebase it onto the configured pb_path.
		text := strings.ReplaceAll(string(content),
			`"../tableau/protobuf/tableau_pb.js"`,
			`"`+pbImportPath+`/tableau/protobuf/tableau_pb.js"`)
		g.P(text)
	}
}
