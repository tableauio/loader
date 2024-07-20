package helper

import (
	"strings"
	"unicode"
)

var golangKeywords map[string]bool

func escapeIdentifier(str string) string {
	// Filter invalid runes
	var result strings.Builder
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			result.WriteRune(r)
		}
	}
	str = result.String()
	// Go variables must not start with digits
	if len(str) != 0 && unicode.IsDigit(rune(str[0])) {
		str = "_" + str
	}
	// avoid go keywords
	if _, ok := golangKeywords[str]; ok {
		return str + "_"
	}
	return str
}

// Ref:
//
//	https://go.dev/ref/spec#Keywords
func init() {
	golangKeywords = map[string]bool{
		"break":       true,
		"case":        true,
		"chan":        true,
		"const":       true,
		"continue":    true,
		"default":     true,
		"defer":       true,
		"pi":          true,
		"else":        true,
		"fallthrough": true,
		"for":         true,
		"func":        true,
		"go":          true,
		"goto":        true,
		"if":          true,
		"import":      true,
		"interface":   true,
		"map":         true,
		"package":     true,
		"range":       true,
		"return":      true,
		"select":      true,
		"struct":      true,
		"switch":      true,
		"type":        true,
		"var":         true,
	}
}
