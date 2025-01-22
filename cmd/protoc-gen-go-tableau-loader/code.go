package main

import (
	"path/filepath"

	"google.golang.org/protobuf/compiler/protogen"
)

const codePkg = "code"

// generateCode generates related code files.
func generateCode(gen *protogen.Plugin) {
	filename := filepath.Join(codePkg, "code."+pcExt+".go")
	g := gen.NewGeneratedFile(filename, "")
	generateCommonHeader(gen, g)
	g.P()
	g.P("package ", codePkg)
	g.P()
	g.P(staticCodeContent)
	g.P()
}

const staticCodeContent = `type Code int

const (
	Success Code = iota
	NotFound
	Unknown
)`
