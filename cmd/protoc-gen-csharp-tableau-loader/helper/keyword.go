package helper

import (
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
)

var csharpKeywords map[string]bool

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
	// Csharp variables must not start with digits
	if len(str) != 0 && unicode.IsDigit(rune(str[0])) {
		str = "_" + str
	}
	// Avoid csharp keywords
	if _, ok := csharpKeywords[str]; ok {
		return str + "_"
	}
	return str
}

// Ref:
//
//	https://learn.microsoft.com/en-us/dotnet/csharp/language-reference/keywords
func init() {
	csharpKeywords = map[string]bool{
		"abstract":   true,
		"as":         true,
		"base":       true,
		"bool":       true,
		"break":      true,
		"byte":       true,
		"case":       true,
		"catch":      true,
		"char":       true,
		"checked":    true,
		"class":      true,
		"const":      true,
		"continue":   true,
		"decimal":    true,
		"default":    true,
		"delegate":   true,
		"do":         true,
		"double":     true,
		"else":       true,
		"enum":       true,
		"event":      true,
		"explicit":   true,
		"extern":     true,
		"false":      true,
		"finally":    true,
		"fixed":      true,
		"float":      true,
		"for":        true,
		"foreach":    true,
		"goto":       true,
		"if":         true,
		"implicit":   true,
		"in":         true,
		"int":        true,
		"interface":  true,
		"internal":   true,
		"is":         true,
		"lock":       true,
		"long":       true,
		"namespace":  true,
		"new":        true,
		"null":       true,
		"object":     true,
		"operator":   true,
		"out":        true,
		"override":   true,
		"params":     true,
		"private":    true,
		"protected":  true,
		"public":     true,
		"readonly":   true,
		"ref":        true,
		"return":     true,
		"sbyte":      true,
		"sealed":     true,
		"short":      true,
		"sizeof":     true,
		"stackalloc": true,
		"static":     true,
		"string":     true,
		"struct":     true,
		"switch":     true,
		"this":       true,
		"throw":      true,
		"true":       true,
		"try":        true,
		"typeof":     true,
		"uint":       true,
		"ulong":      true,
		"unchecked":  true,
		"unsafe":     true,
		"ushort":     true,
		"using":      true,
		"virtual":    true,
		"void":       true,
		"volatile":   true,
		"while":      true,
	}
}
