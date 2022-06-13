package main

import (
	"embed"
	"path/filepath"

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
		content, _ := efs.ReadFile(filepath.Join("embed", entry.Name()))
		g.P(string(content))
	}
}
