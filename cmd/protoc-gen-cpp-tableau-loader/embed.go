package main

import (
	"embed"
	"path"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed embed/*
var efs embed.FS

const (
	includeNotes        = "// Auto-generated includes below\n"
	declarationNotes    = "// Auto-generated declarations below\n"
	specializationNotes = "// Auto-generated specializations below\n"
	initializationNotes = "// Auto-generated initializations below\n"
	fieldNotes          = "// Auto-generated fields below\n"
)

// generateEmbed generates related registry files.
func generateEmbed(gen *protogen.Plugin) {
	entries, err := efs.ReadDir("embed")
	if err != nil {
		panic(err)
	}
	protofiles, fileMessagers := getAllOrderedFilesAndMessagers(gen)
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
		file := string(content)
		switch entry.Name() {
		case "hub.pc.cc":
			// Auto-generated includes below
			impl := ""
			for _, proto := range protofiles {
				impl += `#include "` + proto + "." + pcExt + `.h"` + "\n"
			}
			file = strings.ReplaceAll(file, includeNotes, includeNotes+impl)
			// Auto-generated specializations below
			impl = ""
			for _, proto := range protofiles {
				for _, messager := range fileMessagers[proto] {
					impl += "template <>\n"
					impl += "const std::shared_ptr<" + messager + "> Hub::Get<" + messager + ">() const {\n"
					impl += "  return GetMessagerContainer()->" + strcase.ToSnake(messager) + "_;\n"
					impl += "}\n"
					impl += "\n"
				}
			}
			file = strings.ReplaceAll(file, specializationNotes, specializationNotes+impl)
			// Auto-generated initializations below
			impl = ""
			for _, proto := range protofiles {
				for _, messager := range fileMessagers[proto] {
					impl += "  " + strcase.ToSnake(messager) + "_ = std::dynamic_pointer_cast<" + messager + `>((*msger_map_)["` + messager + `"]);` + "\n"
				}
			}
			file = strings.ReplaceAll(file, initializationNotes, initializationNotes+impl)
		case "hub.pc.h":
			// Auto-generated declarations below
			impl := ""
			for _, proto := range protofiles {
				for _, messager := range fileMessagers[proto] {
					impl += "class " + messager + ";\n"
				}
			}
			file = strings.ReplaceAll(file, declarationNotes, declarationNotes+impl)
			// Auto-generated specializations below
			impl = ""
			for _, proto := range protofiles {
				for _, messager := range fileMessagers[proto] {
					impl += "template <>\n"
					impl += "const std::shared_ptr<" + messager + "> Hub::Get<" + messager + ">() const;\n"
					impl += "\n"
				}
			}
			file = strings.ReplaceAll(file, specializationNotes, specializationNotes+impl)
			// Auto-generated fields below
			impl = ""
			for _, proto := range protofiles {
				for _, messager := range fileMessagers[proto] {
					impl += "  std::shared_ptr<" + messager + "> " + strcase.ToSnake(messager) + "_;\n"
				}
			}
			file = strings.ReplaceAll(file, fieldNotes, fieldNotes+impl)
		}
		g.P(file)
	}
}
