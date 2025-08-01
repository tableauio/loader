package main

import (
	"path/filepath"

	"google.golang.org/protobuf/compiler/protogen"
)

// generateError generates related error files.
func generateError(gen *protogen.Plugin) {
	filename := filepath.Join("errors." + pcExt + ".go")
	g := gen.NewGeneratedFile(filename, "")
	generateCommonHeader(gen, g)
	g.P()
	g.P("package ", *pkg)
	g.P()
	g.P(staticErrorContent)
	g.P()
}

const staticErrorContent = `
var ErrNotFound *errNotFound

type errNotFound struct{}

func (e *errNotFound) Error() string {
	return "not found"
}`
