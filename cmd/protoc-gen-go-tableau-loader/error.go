package main

import (
	"path/filepath"

	"google.golang.org/protobuf/compiler/protogen"
)

const errPkg = "xerrors"

// generateError generates related error files.
func generateError(gen *protogen.Plugin) {
	filename := filepath.Join(errPkg, "xerrors."+pcExt+".go")
	g := gen.NewGeneratedFile(filename, "")
	generateCommonHeader(gen, g)
	g.P()
	g.P("package ", errPkg)
	g.P()
	g.P(`import (`)
	g.P(`"fmt"`)
	g.P(codePackage)
	g.P(")")
	g.P(staticErrorContent)
	g.P()
}

const staticErrorContent = `

type Error struct {
	code  code.Code
	cause error
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, cause: %+v", e.code, e.cause)
}

func Errorf(code code.Code, format string, args ...interface{}) error {
	return &Error{
		code:  code,
		cause: fmt.Errorf(format, args...),
	}
}

func Code(err error) code.Code {
	if err == nil {
		return code.Success
	}
	if ferr, ok := err.(*Error); ok && ferr != nil {
		return ferr.code
	}
	return code.Unknown
}

func Is(err error, code code.Code) bool {
	return Code(err) == code
}`
