package main

import (
	"github.com/tableauio/loader/internal/extensions"
	"google.golang.org/protobuf/compiler/protogen"
)

// generateError generates related error files.
func generateError(gen *protogen.Plugin) {
	filename := "errors." + extensions.PC + ".go"
	g := gen.NewGeneratedFile(filename, "")
	generateCommonHeader(gen, g)
	g.P()
	g.P("package ", *pkg)
	g.P()
	g.P(staticErrorContent)
	g.P()
}

const staticErrorContent = `
import (
	"errors"
)

var ErrNotFound = errors.New("not found")`
