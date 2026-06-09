package helper

import (
	"go/token"
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
)

// escapeIdentifier converts a raw string into a valid Go lowerCamelCase
// identifier, escaping Go reserved words the same way protoc-gen-go's
// GoSanitized does: it consults go/token's keyword table via
// token.Lookup(...).IsKeyword(). On collision a trailing "_" is appended.
//
// Ref:
//
//	https://github.com/protocolbuffers/protobuf-go/blob/master/internal/strs/strings.go (GoSanitized)
//	https://pkg.go.dev/go/token#Lookup
func escapeIdentifier(str string) string {
	// Filter invalid runes
	var result strings.Builder
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			result.WriteRune(r)
		}
	}
	str = result.String()
	// To camel case
	str = strcase.ToLowerCamel(str)
	// Go variables must not start with digits
	if len(str) != 0 && unicode.IsDigit(rune(str[0])) {
		str = "_" + str
	}
	// Avoid Go keywords, plus the loader-specific reserved name "x": the
	// generated Go code uses "x" as the method receiver name (e.g.,
	// `func (x *FooConf) FindIndex1(...)`). If a proto field named "X" is used as
	// an index key, escapeIdentifier converts it to "x" (lowerCamelCase), which
	// would shadow the receiver and cause a compile error, so it is escaped too.
	if token.Lookup(str).IsKeyword() || str == "x" {
		return str + "_"
	}
	return str
}
