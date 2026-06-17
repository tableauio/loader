package genhelper

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// MapKey represents a single key component of a (possibly nested / composite)
// map getter or index finder. It is shared by all protoc-gen-*-tableau-loader
// plugins: the Go / C++ / C# loaders alias it directly (type MapKey =
// genhelper.MapKey), while the TypeScript loader embeds it to add
// language-specific fields, so every loader's MapKey shares the same base.
type MapKey struct {
	// Type is the generated-language type string of the key (e.g. "int32" for
	// Go, "int32_t" for C++, "int" for C#, "number" for TypeScript).
	Type string
	// Name is the parameter/variable name in generated code (deduplicated
	// across nested levels, e.g. "id" → "id3").
	Name string
	// FieldName is the key struct field name, used by multi-column indexes that
	// generate LevelIndex key structs (may be deduplicated, e.g. "Id" → "Id3").
	// Empty for single-column indexes.
	FieldName string
	// OrigFieldName is the original FieldName before deduplication (empty if not
	// renamed).
	OrigFieldName string
	// Fd is the map field descriptor this key belongs to.
	Fd protoreflect.FieldDescriptor
}

// ParamFormatter renders a single MapKey as a function-parameter declaration in
// a target language. It is the ONLY piece of MapKeySlice behaviour that differs
// per language, e.g.:
//
//	Go     "id int32"        (Name + " " + Type)
//	C#     "int id"          (Type + " " + Name)
//	C++    "int32_t id"      (ToConstRefType(Type) + " " + Name)
//	TS     "id: number"      (Name + ": " + Type)
//
// Each loader provides a zero-size implementation and wires it into MapKeySlice
// via the type parameter below, so every loader shares the exact same slice
// type and methods while only "overriding" parameter formatting.
type ParamFormatter interface {
	FormatParam(MapKey) string
}

// MapKeySlice is the single, cross-language ordered collection of MapKey shared
// by all protoc-gen-*-tableau-loader plugins. The type parameter F injects the
// language-specific parameter formatting used by GenGetParams; every other
// method (AddMapKey / GenGetArguments / GenCustom / GenOtherArguments) is fully
// shared. Loaders alias a concrete instantiation, e.g.:
//
//	type MapKeySlice = genhelper.MapKeySlice[goParamFormatter]
type MapKeySlice[F ParamFormatter] []MapKey

// AddMapKey appends newKey to s, automatically deduplicating both Name (used as
// function parameter names) and FieldName (used as struct field names in
// LevelIndex key structs).
//
// Deduplication is needed because different map levels may share the same key
// name. For example, given the following nested proto maps where country_map
// and item_map both use "ID" as their key name:
//
//	message Fruit4Conf {
//	    map<int32, Fruit> fruit_map = 1;          // key field: "FruitType"
//	    message Fruit {
//	        map<int32, Country> country_map = 2;  // key field: "ID"
//	        message Country {
//	            map<int32, Item> item_map = 3;    // key field: "ID"  ← same name!
//	        }
//	    }
//	}
//
// Without dedup, the generated LevelIndex key struct would have duplicate field
// names, causing a compile error. With dedup, the conflicting name gets a
// numeric suffix (the 1-based position of the new key in the slice), producing
// valid code, e.g. "FruitType", "Id", "Id3".
func (s MapKeySlice[F]) AddMapKey(newKey MapKey) MapKeySlice[F] {
	if newKey.Name == "" {
		newKey.Name = fmt.Sprintf("key%d", len(s)+1)
	}
	// Deduplicate Name (used as function parameter, e.g., "id" → "id3").
	for _, key := range s {
		if key.Name == newKey.Name {
			newKey.Name = fmt.Sprintf("%s%d", newKey.Name, len(s)+1)
			break
		}
	}
	// Deduplicate FieldName (used as struct field, e.g., "Id" → "Id3").
	// This is only relevant for multi-column indexes that generate LevelIndex
	// key structs; single-column indexes leave FieldName empty.
	if newKey.FieldName != "" {
		for _, key := range s {
			if key.FieldName == newKey.FieldName {
				newKey.OrigFieldName = newKey.FieldName
				newKey.FieldName = fmt.Sprintf("%s%d", newKey.FieldName, len(s)+1)
				break
			}
		}
	}
	return append(s, newKey)
}

// GenGetParams generates the function parameter list (declarations), formatted
// for the language carried by F (e.g. "id int32, name string").
func (s MapKeySlice[F]) GenGetParams() string {
	var f F
	return s.GenCustom(f.FormatParam, ", ")
}

// GenGetArguments generates the call argument list (the key Names joined by ", ").
func (s MapKeySlice[F]) GenGetArguments() string {
	return s.GenCustom(func(key MapKey) string { return key.Name }, ", ")
}

// GenCustom builds a string by applying fn to each MapKey and joining the
// results with sep. Returns an empty string for an empty slice.
func (s MapKeySlice[F]) GenCustom(fn func(MapKey) string, sep string) string {
	var params []string
	for _, key := range s {
		params = append(params, fn(key))
	}
	return strings.Join(params, sep)
}

// GenOtherArguments generates arguments that access each key by Name on another
// object (e.g. "other.id, other.name"), used by C++ std::tie / hash combine.
func (s MapKeySlice[F]) GenOtherArguments(other string) string {
	return s.GenCustom(func(key MapKey) string { return other + "." + key.Name }, ", ")
}
