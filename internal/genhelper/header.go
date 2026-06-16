// Package genhelper provides cross-language common helpers shared by all
// protoc-gen-*-tableau-loader plugins, such as generated-file header utilities.
package genhelper

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

// ProtocVersion returns the protoc compiler version string (e.g. "v3.19.3")
// extracted from the code generator request. Returns "(unknown)" if not
// available.
func ProtocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

// GenerateSourcePath writes the source path comment (or a deprecation notice)
// for the given file. It is a no-op if file is nil.
func GenerateSourcePath(file *protogen.File, g *protogen.GeneratedFile) {
	if file == nil {
		return
	}
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
}
