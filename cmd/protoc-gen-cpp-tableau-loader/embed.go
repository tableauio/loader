package main

import (
	"embed"
	"path"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed embed/*
var efs embed.FS

// generateEmbed generates related registry files.
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
		helper.GenerateCommonHeader(gen, g, version)
		g.P()
		// refer: [embed: embed path on different OS cannot open file](https://github.com/golang/go/issues/45230)
		content, err := efs.ReadFile(path.Join("embed", entry.Name()))
		if err != nil {
			panic(err)
		}
		g.P(string(content))
	}
}
