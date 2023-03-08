package helper

var golangKeywords map[string]bool

func KeywordEscape(str string) string {
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
