package helper

// tsReservedWords are TypeScript/JavaScript reserved words that cannot be used
// as bare identifiers for function parameters.
var tsReservedWords = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true, "const": true,
	"continue": true, "debugger": true, "default": true, "delete": true, "do": true,
	"else": true, "enum": true, "export": true, "extends": true, "false": true,
	"finally": true, "for": true, "function": true, "if": true, "import": true,
	"in": true, "instanceof": true, "new": true, "null": true, "return": true,
	"super": true, "switch": true, "this": true, "throw": true, "true": true,
	"try": true, "typeof": true, "var": true, "void": true, "while": true,
	"with": true, "let": true, "static": true, "yield": true, "await": true,
}

// escapeIdentifier escapes a TypeScript reserved word by appending an
// underscore, producing a valid identifier.
func escapeIdentifier(name string) string {
	if tsReservedWords[name] {
		return name + "_"
	}
	return name
}
