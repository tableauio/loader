package helper

import (
	"testing"
)

func TestUnderscoresToCamelCase(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		capNextLetter  bool
		preservePeriod bool
		want           string
	}{
		// Empty / special input
		{
			name:           "empty string returns underscore",
			input:          "",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "_",
		},
		{
			name:           "only underscores returns underscore",
			input:          "___",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "_",
		},
		{
			name:           "only special characters returns underscore",
			input:          "!@#$%",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "_",
		},

		// Basic lowercase input with capNextLetter=true (PascalCase)
		{
			name:           "simple lowercase to PascalCase",
			input:          "hello",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "Hello",
		},
		{
			name:           "underscore separated to PascalCase",
			input:          "hello_world",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "HelloWorld",
		},
		{
			name:           "multiple underscores between words",
			input:          "hello__world",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "HelloWorld",
		},
		{
			name:           "leading underscore with capNextLetter true",
			input:          "_hello",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "Hello",
		},
		{
			name:           "trailing underscore",
			input:          "hello_",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "Hello",
		},

		// Basic lowercase input with capNextLetter=false (camelCase)
		{
			name:           "simple lowercase to camelCase",
			input:          "hello",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "hello",
		},
		{
			name:           "underscore separated to camelCase",
			input:          "hello_world",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "helloWorld",
		},

		// Uppercase input
		{
			name:           "uppercase start with capNextLetter false lowercases first char",
			input:          "Hello",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "hello",
		},
		{
			name:           "uppercase start with capNextLetter true preserves first char",
			input:          "Hello",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "Hello",
		},
		{
			name:           "all uppercase with capNextLetter true",
			input:          "ABC",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "ABC",
		},
		{
			name:           "all uppercase with capNextLetter false lowercases first",
			input:          "ABC",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "aBC",
		},

		// Digits
		{
			name:           "digits cause next letter to capitalize",
			input:          "field1name",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "field1Name",
		},
		{
			name:           "leading digit with capNextLetter true",
			input:          "1abc",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "1Abc",
		},
		{
			name:           "only digits",
			input:          "123",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "123",
		},
		{
			name:           "digit at end",
			input:          "abc2",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "Abc2",
		},
		{
			name:           "digits between words",
			input:          "field2name3value",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "Field2Name3Value",
		},

		// Period handling
		{
			name:           "period preserved when preservePeriod is true",
			input:          "my.package.name",
			capNextLetter:  true,
			preservePeriod: true,
			want:           "My.Package.Name",
		},
		{
			name:           "period stripped when preservePeriod is false",
			input:          "my.package.name",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "MyPackageName",
		},
		{
			name:           "period with camelCase and preserve",
			input:          "my.package",
			capNextLetter:  false,
			preservePeriod: true,
			want:           "my.Package",
		},

		// Mixed complex cases
		{
			name:           "mixed case with underscores and digits",
			input:          "my_field2_name",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "MyField2Name",
		},
		{
			name:           "protobuf style field name",
			input:          "repeated_field_name",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "RepeatedFieldName",
		},
		{
			name:           "single char lowercase capNextLetter true",
			input:          "a",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "A",
		},
		{
			name:           "single char lowercase capNextLetter false",
			input:          "a",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "a",
		},
		{
			name:           "single char uppercase capNextLetter true",
			input:          "A",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "A",
		},
		{
			name:           "single char uppercase capNextLetter false",
			input:          "A",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "a",
		},
		{
			name:           "underscore between digits",
			input:          "v1_2_3",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "v123",
		},
		{
			name:           "special chars between letters",
			input:          "hello-world",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "HelloWorld",
		},
		{
			name:           "multiple periods preserved",
			input:          "a.b.c.d",
			capNextLetter:  true,
			preservePeriod: true,
			want:           "A.B.C.D",
		},
		{
			name:           "period at start with preserve",
			input:          ".hello",
			capNextLetter:  true,
			preservePeriod: true,
			want:           ".Hello",
		},
		{
			name:           "period at end with preserve",
			input:          "hello.",
			capNextLetter:  true,
			preservePeriod: true,
			want:           "Hello.",
		},
		{
			name:           "consecutive periods preserved",
			input:          "a..b",
			capNextLetter:  true,
			preservePeriod: true,
			want:           "A..B",
		},
		{
			name:           "mixed special chars with period and no preserve",
			input:          "a.b_c-d",
			capNextLetter:  true,
			preservePeriod: false,
			want:           "ABCD",
		},
		{
			name:           "unicode-like but only ASCII handled",
			input:          "hello_world_foo",
			capNextLetter:  false,
			preservePeriod: false,
			want:           "helloWorldFoo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := underscoresToCamelCase(tt.input, tt.capNextLetter, tt.preservePeriod)
			if got != tt.want {
				t.Errorf("underscoresToCamelCase(%q, %v, %v) = %q, want %q",
					tt.input, tt.capNextLetter, tt.preservePeriod, got, tt.want)
			}
		})
	}
}

func TestEscapeIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Basic cases
		{
			name:  "simple camelCase",
			input: "myField",
			want:  "myField",
		},
		{
			name:  "PascalCase to camelCase",
			input: "MyField",
			want:  "myField",
		},
		{
			name:  "underscore separated",
			input: "my_field_name",
			want:  "myFieldName",
		},

		// Invalid characters filtered
		{
			name:  "special characters removed",
			input: "my-field!name",
			want:  "myfieldname",
		},
		{
			name:  "spaces removed",
			input: "my field",
			want:  "myfield",
		},
		{
			name:  "dots removed",
			input: "my.field",
			want:  "myfield",
		},

		// Digit handling
		{
			name:  "starts with digit gets underscore prefix",
			input: "1field",
			want:  "_1Field",
		},
		{
			name:  "digit in middle is fine",
			input: "field1Name",
			want:  "field1Name",
		},

		// C# keyword escaping
		{
			name:  "csharp keyword int gets underscore suffix",
			input: "int",
			want:  "int_",
		},
		{
			name:  "csharp keyword string gets underscore suffix",
			input: "string",
			want:  "string_",
		},
		{
			name:  "csharp keyword class gets underscore suffix",
			input: "class",
			want:  "class_",
		},
		{
			name:  "csharp keyword bool gets underscore suffix",
			input: "bool",
			want:  "bool_",
		},
		{
			name:  "csharp keyword null gets underscore suffix",
			input: "null",
			want:  "null_",
		},
		{
			name:  "csharp keyword return gets underscore suffix",
			input: "return",
			want:  "return_",
		},
		{
			name:  "csharp keyword void gets underscore suffix",
			input: "void",
			want:  "void_",
		},
		{
			name:  "csharp keyword while gets underscore suffix",
			input: "while",
			want:  "while_",
		},
		{
			name:  "non-keyword similar to keyword is not escaped",
			input: "integer",
			want:  "integer",
		},

		// Empty / edge cases
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only invalid characters",
			input: "!@#$%",
			want:  "",
		},
		{
			name:  "single letter",
			input: "a",
			want:  "a",
		},
		{
			name:  "single uppercase letter",
			input: "A",
			want:  "a",
		},
		{
			name:  "all digits",
			input: "123",
			want:  "_123",
		},
		{
			name:  "underscore only",
			input: "_",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("escapeIdentifier(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMapKeySlice_AddMapKey(t *testing.T) {
	tests := []struct {
		name     string
		initial  MapKeySlice
		newKey   MapKey
		wantName string
		wantLen  int
	}{
		{
			name:     "add first key with empty name gets default name",
			initial:  MapKeySlice{},
			newKey:   MapKey{Type: "int", Name: ""},
			wantName: "key1",
			wantLen:  1,
		},
		{
			name:     "add key with explicit name",
			initial:  MapKeySlice{},
			newKey:   MapKey{Type: "string", Name: "userId"},
			wantName: "userId",
			wantLen:  1,
		},
		{
			name: "add key with conflicting name gets rewritten",
			initial: MapKeySlice{
				{Type: "int", Name: "key1"},
			},
			newKey:   MapKey{Type: "string", Name: "key1"},
			wantName: "key12",
			wantLen:  2,
		},
		{
			name: "add key with empty name to existing slice",
			initial: MapKeySlice{
				{Type: "int", Name: "key1"},
			},
			newKey:   MapKey{Type: "string", Name: ""},
			wantName: "key2",
			wantLen:  2,
		},
		{
			name: "add key with non-conflicting name",
			initial: MapKeySlice{
				{Type: "int", Name: "key1"},
			},
			newKey:   MapKey{Type: "string", Name: "key2"},
			wantName: "key2",
			wantLen:  2,
		},
		{
			name: "add key with conflict among multiple existing keys",
			initial: MapKeySlice{
				{Type: "int", Name: "id"},
				{Type: "string", Name: "name"},
			},
			newKey:   MapKey{Type: "float", Name: "id"},
			wantName: "id3",
			wantLen:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.AddMapKey(tt.newKey)
			if len(result) != tt.wantLen {
				t.Errorf("AddMapKey() result length = %d, want %d", len(result), tt.wantLen)
			}
			lastKey := result[len(result)-1]
			if lastKey.Name != tt.wantName {
				t.Errorf("AddMapKey() last key name = %q, want %q", lastKey.Name, tt.wantName)
			}
		})
	}
}

func TestMapKeySlice_GenGetParams(t *testing.T) {
	tests := []struct {
		name  string
		slice MapKeySlice
		want  string
	}{
		{
			name:  "empty slice",
			slice: MapKeySlice{},
			want:  "",
		},
		{
			name: "single key",
			slice: MapKeySlice{
				{Type: "int", Name: "key1"},
			},
			want: "int key1",
		},
		{
			name: "multiple keys",
			slice: MapKeySlice{
				{Type: "int", Name: "key1"},
				{Type: "string", Name: "key2"},
			},
			want: "int key1, string key2",
		},
		{
			name: "three keys",
			slice: MapKeySlice{
				{Type: "int", Name: "id"},
				{Type: "string", Name: "name"},
				{Type: "float", Name: "score"},
			},
			want: "int id, string name, float score",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.slice.GenGetParams()
			if got != tt.want {
				t.Errorf("GenGetParams() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMapKeySlice_GenGetArguments(t *testing.T) {
	tests := []struct {
		name  string
		slice MapKeySlice
		want  string
	}{
		{
			name:  "empty slice",
			slice: MapKeySlice{},
			want:  "",
		},
		{
			name: "single key",
			slice: MapKeySlice{
				{Type: "int", Name: "key1"},
			},
			want: "key1",
		},
		{
			name: "multiple keys",
			slice: MapKeySlice{
				{Type: "int", Name: "key1"},
				{Type: "string", Name: "key2"},
			},
			want: "key1, key2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.slice.GenGetArguments()
			if got != tt.want {
				t.Errorf("GenGetArguments() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMapKeySlice_GenCustom(t *testing.T) {
	tests := []struct {
		name  string
		slice MapKeySlice
		fn    func(MapKey) string
		sep   string
		want  string
	}{
		{
			name:  "empty slice",
			slice: MapKeySlice{},
			fn:    func(key MapKey) string { return key.Name },
			sep:   ", ",
			want:  "",
		},
		{
			name: "custom format with pipe separator",
			slice: MapKeySlice{
				{Type: "int", Name: "a"},
				{Type: "string", Name: "b"},
			},
			fn:   func(key MapKey) string { return key.Type + ":" + key.Name },
			sep:  " | ",
			want: "int:a | string:b",
		},
		{
			name: "custom format with no separator",
			slice: MapKeySlice{
				{Type: "int", Name: "x"},
				{Type: "int", Name: "y"},
			},
			fn:   func(key MapKey) string { return key.Name },
			sep:  "",
			want: "xy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.slice.GenCustom(tt.fn, tt.sep)
			if got != tt.want {
				t.Errorf("GenCustom() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIndent(t *testing.T) {
	tests := []struct {
		name  string
		depth int
		want  string
	}{
		{
			name:  "zero depth",
			depth: 0,
			want:  "",
		},
		{
			name:  "depth 1",
			depth: 1,
			want:  "    ",
		},
		{
			name:  "depth 2",
			depth: 2,
			want:  "        ",
		},
		{
			name:  "depth 3",
			depth: 3,
			want:  "            ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Indent(tt.depth)
			if got != tt.want {
				t.Errorf("Indent(%d) = %q, want %q", tt.depth, got, tt.want)
			}
		})
	}
}
