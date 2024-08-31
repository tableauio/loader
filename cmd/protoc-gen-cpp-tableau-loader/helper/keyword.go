package helper

import (
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
)

var cppKeywords map[string]bool

func EscapeIdentifier(str string) string {
	// Filter invalid runes
	var result strings.Builder
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			result.WriteRune(r)
		}
	}
	str = result.String()
	// To snake case
	str = strcase.ToSnake(str)
	// Cpp variables must not start with digits
	if len(str) != 0 && unicode.IsDigit(rune(str[0])) {
		str = "_" + str
	}
	// Avoid cpp keywords
	if _, ok := cppKeywords[str]; ok {
		return str + "_"
	}
	return str
}

// Ref:
//
//	https://en.cppreference.com/w/cpp/keyword
func init() {
	cppKeywords = map[string]bool{
		"alignas":          true,
		"alignof":          true,
		"and":              true,
		"and_eq":           true,
		"asm":              true,
		"atomic_cancel":    true,
		"atomic_commit":    true,
		"atomic_noexcept":  true,
		"auto":             true,
		"bitand":           true,
		"bitor":            true,
		"bool":             true,
		"break":            true,
		"case":             true,
		"catch":            true,
		"char":             true,
		"char8_t":          true,
		"char16_t":         true,
		"char32_t":         true,
		"class":            true,
		"compl":            true,
		"concept":          true,
		"const":            true,
		"consteval":        true,
		"constexpr":        true,
		"constinit":        true,
		"const_cast":       true,
		"continue":         true,
		"co_await":         true,
		"co_return":        true,
		"co_yield":         true,
		"decltype":         true,
		"default":          true,
		"delete":           true,
		"do":               true,
		"double":           true,
		"dynamic_cast":     true,
		"else":             true,
		"enum":             true,
		"explicit":         true,
		"export":           true,
		"extern":           true,
		"false":            true,
		"float":            true,
		"for":              true,
		"friend":           true,
		"goto":             true,
		"if":               true,
		"inline":           true,
		"int":              true,
		"long":             true,
		"mutable":          true,
		"namespace":        true,
		"new":              true,
		"noexcept":         true,
		"not":              true,
		"not_eq":           true,
		"nullptr":          true,
		"operator":         true,
		"or":               true,
		"or_eq":            true,
		"private":          true,
		"protected":        true,
		"public":           true,
		"register":         true,
		"reflexpr":         true,
		"reinterpret_cast": true,
		"requires":         true,
		"return":           true,
		"short":            true,
		"signed":           true,
		"sizeof":           true,
		"static":           true,
		"static_assert":    true,
		"static_cast":      true,
		"struct":           true,
		"switch":           true,
		"synchronized":     true,
		"template":         true,
		"this":             true,
		"thread_local":     true,
		"throw":            true,
		"true":             true,
		"try":              true,
		"typedef":          true,
		"typeid":           true,
		"typename":         true,
		"union":            true,
		"unsigned":         true,
		"using":            true,
		"virtual":          true,
		"void":             true,
		"volatile":         true,
		"wchar_t":          true,
		"while":            true,
		"xor":              true,
		"xor_eq":           true,
	}
}
