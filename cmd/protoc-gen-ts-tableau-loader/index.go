package main

import (
	"fmt"
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-ts-tableau-loader/helper"
	"github.com/tableauio/loader/internal/genhelper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/loadutil"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// indexGen generates index / ordered index / ordered map support for a single
// worksheet messager.
//
// It mirrors the leveled-container model of the Go / C# / C++ generators: an
// index may be declared on a column that lives at any container level, and for
// every map ancestor of that level a "leveled" container is generated so the
// index can be queried scoped to a specific upper map key (e.g. findItem1).
//
// JavaScript Maps compare object keys by reference, so a single upper key
// (k1) buckets into a native Map keyed by its raw value, while composite upper
// keys (k1..ki, i>=2) bucket into a TupleKeyMap keyed by the upper key tuple
// (which serializes it internally via makeIndexKey).
type indexGen struct {
	g          *protogen.GeneratedFile
	descriptor *index.IndexDescriptor
	message    *protogen.Message
	pkgs       *pbPackages

	// keys holds, in order, the map key of every ancestor map level whose
	// deeper level still needs an index. They are the upper keys used by
	// leveled containers/finders. Names are deduplicated (e.g. id -> id3).
	keys helper.MapKeySlice
}

func newIndexGen(g *protogen.GeneratedFile, descriptor *index.IndexDescriptor, message *protogen.Message, pkgs *pbPackages) *indexGen {
	x := &indexGen{g: g, descriptor: descriptor, message: message, pkgs: pkgs}
	x.initLevelKeys()
	return x
}

// initLevelKeys collects the upper map keys used to build leveled containers.
func (x *indexGen) initLevelKeys() {
	for lm := x.descriptor.LevelMessage; lm != nil; lm = lm.NextLevel {
		if fd := lm.FD; fd != nil && fd.IsMap() {
			// Only collect keys when a deeper level needs an index, because the
			// keys are used solely for building leveled (upper-level) containers.
			if !lm.NextLevel.NeedGenAnyIndex() {
				break
			}
			paramName := helper.ParseMapFieldNameAsFuncParam(fd)
			key := helper.ParseMapKey(fd.MapKey(), paramName)
			key.Fd = fd
			x.keys = x.keys.AddMapKey(key)
		}
	}
}

// firstMapField returns the first map field of a message descriptor, or nil if
// it has none. The ordered map walks the message tree along this first-map-field
// chain (independently of indexes), mirroring the Go / C++ / C# loaders.
func firstMapField(md protoreflect.MessageDescriptor) protoreflect.FieldDescriptor {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsMap() {
			return fd
		}
	}
	return nil
}

// nextLevelMapField returns the first map field of a map field's value message
// (the next ordered-map level down), or nil if the value is a scalar/enum or a
// message that holds no map (i.e. fd is the leaf level).
func nextLevelMapField(fd protoreflect.FieldDescriptor) protoreflect.FieldDescriptor {
	if fd.MapValue().Kind() == protoreflect.MessageKind {
		return firstMapField(fd.MapValue().Message())
	}
	return nil
}

func (x *indexGen) needIndex() bool {
	return x.descriptor.LevelMessage.NeedGenIndex() && options.NeedGenIndex(x.message.Desc, options.LangTS)
}

func (x *indexGen) needOrderedIndex() bool {
	return x.descriptor.LevelMessage.NeedGenOrderedIndex() && options.NeedGenOrderedIndex(x.message.Desc, options.LangTS)
}

func (x *indexGen) needOrderedMap() bool {
	return firstMapField(x.message.Desc) != nil && options.NeedGenOrderedMap(x.message.Desc, options.LangTS)
}

// needNestedOrderedMap reports whether the ordered map has 2+ levels, i.e. its
// value nodes are OrderedMapValue pairs (sub-map + value) rather than plain
// leaf values. Used to decide whether the OrderedMapValue runtime type must be
// imported.
func (x *indexGen) needNestedOrderedMap() bool {
	return x.needOrderedMap() && nextLevelMapField(firstMapField(x.message.Desc)) != nil
}

// NeedGenerate reports whether any index/ordered-index/ordered-map code is emitted.
func (x *indexGen) NeedGenerate() bool {
	return x.needIndex() || x.needOrderedIndex() || x.needOrderedMap()
}

// ----------------------------------------------------------------------------
// Type / name helpers
// ----------------------------------------------------------------------------

// valueType returns the package-qualified TypeScript type of an index's value
// message (the map/list value at the index's level).
func (x *indexGen) valueType(idx *index.LevelIndex) string {
	return x.pkgs.aliasOf(idx.MD) + "." + helper.LocalTypeName(idx.MD)
}

// keyTSType returns the TypeScript type used as a single-column index key.
func (x *indexGen) keyTSType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.EnumKind:
		return x.pkgs.aliasOf(fd.Enum()) + "." + helper.LocalTypeName(fd.Enum())
	case protoreflect.MessageKind, protoreflect.GroupKind:
		// Only Timestamp/Duration are valid index keys here; keyed by seconds.
		return "bigint"
	default:
		s, _ := helper.ScalarTSType(fd.Kind())
		return s
	}
}

// indexKeyType returns the TypeScript Map key type for an index (single-column
// uses the column's scalar/enum type; multi-column uses a serialized string).
func (x *indexGen) indexKeyType(idx *index.LevelIndex) string {
	if len(idx.ColFields) == 1 {
		return x.keyTSType(idx.ColFields[0].FD)
	}
	return "string"
}

// params builds the finder parameter slice for an index's columns.
func (x *indexGen) params(idx *index.LevelIndex) helper.MapKeySlice {
	var keys helper.MapKeySlice
	for _, field := range idx.ColFields {
		keys = keys.AddMapKey(helper.MapKey{
			MapKey: genhelper.MapKey{
				Type: x.keyTSType(field.FD),
				Name: helper.IndexFieldNameAsFuncParam(field.FD),
			},
		})
	}
	return keys
}

// containerField returns the private container field name for index idx at
// leveled depth i (i==0 is the global container).
func (x *indexGen) containerField(idx *index.LevelIndex, ordered bool, i int) string {
	base := "index" + idx.Name() + "Map"
	if ordered {
		base = "orderedIndex" + idx.Name() + "Map"
	}
	if i == 0 {
		return "#" + base
	}
	return fmt.Sprintf("#%s%d", base, i)
}

const orderedMapField = "#orderedMap"

// ----------------------------------------------------------------------------
// Type aliases
//
// Mirroring the named container types of the C#/C++/Go loaders, every index /
// ordered index / ordered map container gets a readable type alias declared in
// a namespace merged with the messager class (e.g. ItemConf.Index_AwardItemMap).
// The namespace only contains `type` members, so it is fully erased at compile
// time (no runtime overhead) while making the finder signatures self-documenting.
// ----------------------------------------------------------------------------

// aliasName returns the bare type-alias name for an index's container (e.g.
// Index_AwardItemMap / OrderedIndex_ParamExtTypeMap), as declared inside the
// messager namespace.
func (x *indexGen) aliasName(idx *index.LevelIndex, ordered bool) string {
	if ordered {
		return "OrderedIndex_" + idx.Name() + "Map"
	}
	return "Index_" + idx.Name() + "Map"
}

// indexType returns the messager-qualified alias used inside the class (e.g.
// ItemConf.Index_AwardItemMap), so field declarations and finder signatures
// reference the readable named type rather than an inline Map/TupleKeyMap.
func (x *indexGen) indexType(idx *index.LevelIndex, ordered bool) string {
	return string(x.message.Desc.Name()) + "." + x.aliasName(idx, ordered)
}

// keyAliasName returns the bare composite-key tuple alias name for a
// multi-column index (e.g. Index_AwardItemKey / OrderedIndex_ParamExtTypeKey),
// declared inside the messager namespace.
func (x *indexGen) keyAliasName(idx *index.LevelIndex, ordered bool) string {
	if ordered {
		return "OrderedIndex_" + idx.Name() + "Key"
	}
	return "Index_" + idx.Name() + "Key"
}

// keyAliasType returns the messager-qualified composite-key alias (e.g.
// ItemConf.Index_AwardItemKey), used to explicitly parameterize TupleKeyMap at
// its (invariant-in-K) construction sites.
func (x *indexGen) keyAliasType(idx *index.LevelIndex, ordered bool) string {
	return string(x.message.Desc.Name()) + "." + x.keyAliasName(idx, ordered)
}

// keyTupleType returns the labeled readonly tuple type of a multi-column
// index's key columns (e.g. "readonly [id: number, name: string]"). The labels
// are purely for editor hints; assignability ignores them.
func (x *indexGen) keyTupleType(idx *index.LevelIndex) string {
	var parts []string
	for _, field := range idx.ColFields {
		parts = append(parts, helper.IndexFieldNameAsFuncParam(field.FD)+": "+x.keyTSType(field.FD))
	}
	return "readonly [" + strings.Join(parts, ", ") + "]"
}

// upperKeyTupleType returns the labeled readonly tuple type of the first i
// upper-level map keys (e.g. "readonly [activityId: bigint, chapterId:
// number]"), used as the right-hand side of the leveled-container key alias.
func (x *indexGen) upperKeyTupleType(i int) string {
	var parts []string
	for _, k := range x.keys[:i] {
		parts = append(parts, k.Name+": "+k.Type)
	}
	return "readonly [" + strings.Join(parts, ", ") + "]"
}

// upperKeyAliasName returns the bare composite upper-key tuple alias name for
// the depth-i leveled containers (e.g. LevelIndex_Activity_ChapterKey),
// declared inside the messager namespace. It mirrors the Go loader's
// <Messager>_LevelIndex_<prefix>Key struct, the messager prefix being supplied
// by the enclosing namespace. Only meaningful for i>=2 (composite upper keys).
func (x *indexGen) upperKeyAliasName(i int) string {
	return "LevelIndex_" + helper.ParseLeveledMapPrefix(x.message.Desc, x.keys[i-1].Fd) + "Key"
}

// upperKeyAliasType returns the messager-qualified composite upper-key alias
// (e.g. ActivityConf.LevelIndex_Activity_ChapterKey), used to type and
// construct the i>=2 leveled TupleKeyMap containers.
func (x *indexGen) upperKeyAliasType(i int) string {
	return string(x.message.Desc.Name()) + "." + x.upperKeyAliasName(i)
}

// newLeafContainer returns the construction expression for a multi-column
// index's leaf TupleKeyMap (key tuple -> value list), explicitly parameterized
// because TupleKeyMap is invariant in K and constructed empty.
func (x *indexGen) newLeafContainer(idx *index.LevelIndex, ordered bool) string {
	return "new TupleKeyMap<" + x.keyAliasType(idx, ordered) + ", " + x.valueType(idx) + "[]>()"
}

// orderedMapAliasOf returns the bare ordered-map alias name for a map field
// (e.g. OrderedMap_ActivityMap / OrderedMap_protoconf_SectionMap /
// OrderedMap_int32Map), using the same leveled-map prefix as the Go / C++ / C#
// loaders so the cross-language container names line up.
func (x *indexGen) orderedMapAliasOf(mapFd protoreflect.FieldDescriptor) string {
	return "OrderedMap_" + helper.ParseLeveledMapPrefix(x.message.Desc, mapFd) + "Map"
}

// orderedMapValueAliasOf returns the bare ordered-map value (pair) alias name
// for an intermediate map field (e.g. OrderedMap_ActivityValue), used for the
// OrderedMapValue<subMap, value> node alias of a non-leaf level.
func (x *indexGen) orderedMapValueAliasOf(mapFd protoreflect.FieldDescriptor) string {
	return "OrderedMap_" + helper.ParseLeveledMapPrefix(x.message.Desc, mapFd) + "Value"
}

// orderedMapTypeOf returns the messager-qualified ordered-map alias for a map
// field (e.g. ActivityConf.OrderedMap_Activity_ChapterMap).
func (x *indexGen) orderedMapTypeOf(mapFd protoreflect.FieldDescriptor) string {
	return string(x.message.Desc.Name()) + "." + x.orderedMapAliasOf(mapFd)
}

// orderedMapType returns the messager-qualified top-level (1st-level) ordered-map
// alias.
func (x *indexGen) orderedMapType() string {
	return x.orderedMapTypeOf(firstMapField(x.message.Desc))
}

// genOrderedMapAliases recursively emits the nested ordered-map type aliases for
// the first-map-field chain rooted at md. Each level is a Map keyed in sorted
// order; an intermediate level's value is an OrderedMapValue pair (next-level
// sub-map + this level's message value), while the leaf level stores its plain
// value directly — mirroring the TreeMap<key, Pair<subMap, value>> / leaf
// TreeMap<key, value> shape of the Go / C++ / C# loaders. Deeper aliases are
// emitted first so the file reads inner-to-outer like the Go loader (all
// aliases share one namespace, so referencing order does not matter).
func (x *indexGen) genOrderedMapAliases(md protoreflect.MessageDescriptor, depth int) {
	fd := firstMapField(md)
	if fd == nil {
		return
	}
	nextFd := nextLevelMapField(fd)
	if fd.MapValue().Kind() == protoreflect.MessageKind {
		x.genOrderedMapAliases(fd.MapValue().Message(), depth+1)
	}
	k := helper.ParseMapKey(fd.MapKey(), "").Type
	alias := x.orderedMapAliasOf(fd)
	ordinal := loadutil.Ordinal(depth)
	if nextFd != nil {
		valueAlias := x.orderedMapValueAliasOf(fd)
		nextMap := x.orderedMapAliasOf(nextFd)
		curr := mapValueType(fd, x.pkgs)
		x.g.P(helper.Indent(1), "/** ", valueAlias, " is a ", ordinal, "-level node: its next-level sub-map (first) plus this level's value (second). */")
		x.g.P(helper.Indent(1), "export type ", valueAlias, " = OrderedMapValue<", nextMap, ", ", curr, ">;")
		x.g.P(helper.Indent(1), "/** ", alias, " is the ", ordinal, "-level ordered map: key -> node (sorted by key). */")
		x.g.P(helper.Indent(1), "export type ", alias, " = Map<", k, ", ", valueAlias, ">;")
		return
	}
	v := mapValueType(fd, x.pkgs)
	x.g.P(helper.Indent(1), "/** ", alias, " is the ", ordinal, "-level (leaf) ordered map: key -> value (sorted by key). */")
	x.g.P(helper.Indent(1), "export type ", alias, " = Map<", k, ", ", v, ">;")
}

// GenTypeAliases emits the messager namespace holding the index / ordered index
// / ordered map type aliases. Declaration-merged with the class, it is purely
// type-level and erased at runtime.
func (x *indexGen) GenTypeAliases() {
	if !x.NeedGenerate() {
		return
	}
	name := string(x.message.Desc.Name())
	x.g.P()
	x.g.P("// Type aliases for the index / ordered index / ordered map containers,")
	x.g.P("// mirroring the named container types of the other-language loaders so the")
	x.g.P("// finder signatures read clearly (e.g. ", name, ".Index_XxxMap).")
	x.g.P("export namespace ", name, " {")
	if x.needOrderedMap() {
		x.genOrderedMapAliases(x.message.Desc, 1)
	}
	// Composite upper-key tuple aliases shared by every depth-i (i>=2) leveled
	// index / ordered-index container, mirroring the Go loader's
	// <Messager>_LevelIndex_<prefix>Key structs. A single upper key (depth 1)
	// buckets into a native Map by its raw value, so no alias is needed there.
	for i := 2; i <= len(x.keys); i++ {
		alias := x.upperKeyAliasName(i)
		x.g.P(helper.Indent(1), "/** ", alias, " is the composite upper map key (k1..k", i, ") of the ", loadutil.Ordinal(i), "-level leveled containers. */")
		x.g.P(helper.Indent(1), "export type ", alias, " = ", x.upperKeyTupleType(i), ";")
	}
	if x.needIndex() {
		x.eachIndex(false, func(lm *index.LevelMessage, idx *index.LevelIndex) {
			x.genAliasDecl(idx, false)
		})
	}
	if x.needOrderedIndex() {
		x.eachIndex(true, func(lm *index.LevelMessage, idx *index.LevelIndex) {
			x.genAliasDecl(idx, true)
		})
	}
	x.g.P("}")
}

// genAliasDecl emits one index/ordered-index alias. Single-column indexes alias
// the native Map; multi-column indexes first declare a labeled readonly tuple
// for the composite key (e.g. Index_AwardItemKey = readonly [id: number,
// name: string]) and alias a TupleKeyMap parameterized by it, so the finders'
// keys are fully typed instead of an opaque unknown[]. TupleKeyMap is the
// owning multi-column container (key tuple -> value list), keyed internally by
// the opaque serialized composite key, which is an internal serialization
// detail.
func (x *indexGen) genAliasDecl(idx *index.LevelIndex, ordered bool) {
	v := x.valueType(idx)
	alias := x.aliasName(idx, ordered)
	label := "index"
	if ordered {
		label = "ordered index"
	}
	if len(idx.ColFields) > 1 {
		keyAlias := x.keyAliasName(idx, ordered)
		x.g.P(helper.Indent(1), "/** ", keyAlias, " is the composite key of ", label, ": key(", idx.Index, "). */")
		x.g.P(helper.Indent(1), "export type ", keyAlias, " = ", x.keyTupleType(idx), ";")
		x.g.P(helper.Indent(1), "/** ", alias, " is the ", label, " map: key(", idx.Index, ") -> values. */")
		x.g.P(helper.Indent(1), "export type ", alias, " = TupleKeyMap<", keyAlias, ", ", v, "[]>;")
		return
	}
	x.g.P(helper.Indent(1), "/** ", alias, " is the ", label, " map: key(", idx.Index, ") -> values. */")
	x.g.P(helper.Indent(1), "export type ", alias, " = Map<", x.indexKeyType(idx), ", ", v, "[]>;")
}

// eachIndex iterates every index (or ordered index) across all levels.
func (x *indexGen) eachIndex(ordered bool, fn func(lm *index.LevelMessage, idx *index.LevelIndex)) {
	for lm := x.descriptor.LevelMessage; lm != nil; lm = lm.NextLevel {
		indexes := lm.Indexes
		if ordered {
			indexes = lm.OrderedIndexes
		}
		for _, idx := range indexes {
			fn(lm, idx)
		}
	}
}

// ----------------------------------------------------------------------------
// Declarations
// ----------------------------------------------------------------------------

// GenDecls emits the private container field declarations.
func (x *indexGen) GenDecls() {
	if x.needOrderedMap() {
		x.g.P(helper.Indent(1), orderedMapField, ": ", x.orderedMapType(), " = new Map();")
	}
	if x.needIndex() {
		x.eachIndex(false, func(lm *index.LevelMessage, idx *index.LevelIndex) {
			x.genContainerDecls(lm, idx, false)
		})
	}
	if x.needOrderedIndex() {
		x.eachIndex(true, func(lm *index.LevelMessage, idx *index.LevelIndex) {
			x.genContainerDecls(lm, idx, true)
		})
	}
}

func (x *indexGen) genContainerDecls(lm *index.LevelMessage, idx *index.LevelIndex, ordered bool) {
	at := x.indexType(idx, ordered)
	multi := len(idx.ColFields) > 1
	// The global (0-level) container: single-column indexes use the native Map
	// alias; multi-column indexes use the owning leaf TupleKeyMap (key tuple ->
	// value list), which holds the serialized-key -> tuple side table
	// internally. The field's alias annotation (at) resolves to the fully
	// parameterized TupleKeyMap, so the constructor's type arguments are
	// inferred from it contextually and need not be repeated on the `new` side.
	if multi {
		x.g.P(helper.Indent(1), x.containerField(idx, ordered, 0), ": ", at, " = new TupleKeyMap();")
	} else {
		x.g.P(helper.Indent(1), x.containerField(idx, ordered, 0), ": ", at, " = new Map();")
	}
	// Leveled containers bucket the global container alias under the upper map
	// key(s): a single upper key (i==1) uses a native Map keyed by its raw
	// value; composite upper keys (i>=2) use a TupleKeyMap keyed by the upper
	// key tuple alias (serialized internally), so no makeIndexKey is needed at
	// the call sites.
	for i := 1; i < lm.LeveledContainerDepth(); i++ {
		if i == 1 {
			x.g.P(helper.Indent(1), x.containerField(idx, ordered, i), ": Map<", x.keys[0].Type, ", ", at, "> = new Map();")
		} else {
			ut := x.upperKeyAliasType(i)
			// The field carries the full TupleKeyMap<ut, at> annotation, so the
			// constructor's type arguments are inferred from it contextually;
			// repeating them on the `new` side would be redundant.
			x.g.P(helper.Indent(1), x.containerField(idx, ordered, i), ": TupleKeyMap<", ut, ", ", at, "> = new TupleKeyMap();")
		}
	}
}

// ----------------------------------------------------------------------------
// processAfterLoad body
// ----------------------------------------------------------------------------

// GenProcessAfterLoadBody emits the body of the processAfterLoad override
// (statements at indent level 2).
func (x *indexGen) GenProcessAfterLoadBody() {
	if x.needOrderedMap() {
		x.genOrderedMapLoader()
	}
	if x.needIndex() {
		x.genIndexSection(false)
	}
	if x.needOrderedIndex() {
		x.genIndexSection(true)
	}
}

func (x *indexGen) genOrderedMapLoader() {
	x.g.P(helper.Indent(2), "// OrderedMap init.")
	x.g.P(helper.Indent(2), "const orderedMap: ", x.orderedMapType(), " = new Map();")
	x.genOrderedMapLoaderLevel(x.message.Desc, 1, "orderedMap", "this.#data", 2)
	x.g.P(helper.Indent(2), "this.", orderedMapField, " = orderedMap;")
}

// genOrderedMapLoaderLevel recursively emits the nested loops that fill one
// ordered-map level. It converts each string object-key, and at an intermediate
// level builds the next-level sub-map (recursing) then stores the
// {first: subMap, second: value} node; at the leaf level it stores the plain
// value. Each level is sorted by key once fully built, so iteration is
// ascending at every depth.
func (x *indexGen) genOrderedMapLoaderLevel(md protoreflect.MessageDescriptor, depth int, targetMap, sourceExpr string, indent int) {
	fd := firstMapField(md)
	field := helper.FieldLocalName(fd)
	keyStr := fmt.Sprintf("k%dStr", depth)
	keyVar := fmt.Sprintf("k%d", depth)
	valVar := fmt.Sprintf("v%d", depth)
	nextFd := nextLevelMapField(fd)
	x.g.P(helper.Indent(indent), "for (const [", keyStr, ", ", valVar, "] of Object.entries(", sourceExpr, ".", field, ")) {")
	x.g.P(helper.Indent(indent+1), "const ", keyVar, " = ", x.mapKeyConv(fd.MapKey(), keyStr), ";")
	if nextFd != nil {
		subMap := fmt.Sprintf("orderedMap%d", depth+1)
		x.g.P(helper.Indent(indent+1), "const ", subMap, ": ", x.orderedMapTypeOf(nextFd), " = new Map();")
		x.genOrderedMapLoaderLevel(fd.MapValue().Message(), depth+1, subMap, valVar, indent+1)
		x.g.P(helper.Indent(indent+1), targetMap, ".set(", keyVar, ", { first: ", subMap, ", second: ", valVar, " });")
	} else {
		x.g.P(helper.Indent(indent+1), targetMap, ".set(", keyVar, ", ", valVar, ");")
	}
	x.g.P(helper.Indent(indent), "}")
	x.g.P(helper.Indent(indent), "sortMapByKey(", targetMap, ", compareValues);")
}

// mapKeyConv converts a string object-key (from Object.entries) into the proper
// map key type.
func (x *indexGen) mapKeyConv(keyFd protoreflect.FieldDescriptor, v string) string {
	switch helper.ParseMapKey(keyFd, "").Type {
	case "number":
		return "Number(" + v + ")"
	case "bigint":
		return "BigInt(" + v + ")"
	case "boolean":
		return v + ` === "true"`
	default:
		return v
	}
}

func (x *indexGen) genIndexSection(ordered bool) {
	label := "Index"
	if ordered {
		label = "OrderedIndex"
	}
	x.g.P(helper.Indent(2), "// ", label, " init.")
	// Clear all containers.
	x.eachIndex(ordered, func(lm *index.LevelMessage, idx *index.LevelIndex) {
		x.g.P(helper.Indent(2), "this.", x.containerField(idx, ordered, 0), ".clear();")
		for i := 1; i < lm.LeveledContainerDepth(); i++ {
			x.g.P(helper.Indent(2), "this.", x.containerField(idx, ordered, i), ".clear();")
		}
	})
	// Build containers via nested level traversal.
	x.genLevelLoop(x.descriptor.LevelMessage, ordered, "this.#data", 2)
	// Sort value lists by sorted columns.
	x.eachIndex(ordered, func(lm *index.LevelMessage, idx *index.LevelIndex) {
		x.genValueListSorter(lm, idx, ordered)
	})
	// Rebuild ordered maps sorted by key.
	if ordered {
		x.eachIndex(true, func(lm *index.LevelMessage, idx *index.LevelIndex) {
			x.genOrderedRebuild(lm, idx)
		})
	}
}

// genLevelLoop recursively emits the nested for-loops that walk the level
// hierarchy and build the index containers.
func (x *indexGen) genLevelLoop(lm *index.LevelMessage, ordered bool, parentData string, indent int) {
	if lm == nil {
		return
	}
	need := lm.NeedGenIndex()
	if ordered {
		need = lm.NeedGenOrderedIndex()
	}
	if !need {
		return
	}
	field := helper.FieldLocalName(lm.FD)
	itemVar := fmt.Sprintf("item%d", lm.Depth)
	if lm.FD.IsMap() {
		needKey := lm.NeedMapKeyForIndex()
		if ordered {
			needKey = lm.NeedMapKeyForOrderedIndex()
		}
		if needKey {
			keyStr := fmt.Sprintf("k%dStr", lm.MapDepth)
			x.g.P(helper.Indent(indent), "for (const [", keyStr, ", ", itemVar, "] of Object.entries(", parentData, ".", field, ")) {")
			x.g.P(helper.Indent(indent+1), "const k", lm.MapDepth, " = ", x.mapKeyConv(lm.FD.MapKey(), keyStr), ";")
		} else {
			x.g.P(helper.Indent(indent), "for (const ", itemVar, " of Object.values(", parentData, ".", field, ")) {")
		}
	} else {
		x.g.P(helper.Indent(indent), "for (const ", itemVar, " of ", parentData, ".", field, " ?? []) {")
	}
	indexes := lm.Indexes
	if ordered {
		indexes = lm.OrderedIndexes
	}
	for _, idx := range indexes {
		x.genOneIndexLoader(lm, idx, ordered, indent+1, itemVar)
	}
	x.genLevelLoop(lm.NextLevel, ordered, itemVar, indent+1)
	x.g.P(helper.Indent(indent), "}")
}

func (x *indexGen) genOneIndexLoader(lm *index.LevelMessage, idx *index.LevelIndex, ordered bool, indent int, itemVar string) {
	label := "Index"
	if ordered {
		label = "OrderedIndex"
	}
	x.g.P(helper.Indent(indent), "{")
	x.g.P(helper.Indent(indent+1), "// ", label, ": ", idx.Index)
	if len(idx.ColFields) == 1 {
		field := idx.ColFields[0]
		access, isList := x.fieldAccess(itemVar, field)
		if isList {
			x.g.P(helper.Indent(indent+1), "for (const elem of ", access, ") {")
			x.g.P(helper.Indent(indent+2), "const key = elem;")
			x.emitPushAll(lm, idx, ordered, indent+2, "key", itemVar)
			x.g.P(helper.Indent(indent+1), "}")
		} else {
			x.g.P(helper.Indent(indent+1), "const key = ", access, ";")
			x.emitPushAll(lm, idx, ordered, indent+1, "key", itemVar)
		}
	} else {
		x.genMultiCol(lm, idx, ordered, 0, indent+1, itemVar, nil)
	}
	x.g.P(helper.Indent(indent), "}")
}

// genMultiCol recursively emits nested loops for list columns and finally
// builds the key-column tuple, which is stored into the leaf TupleKeyMap
// container(s) via getOrSet (the container serializes it and records the tuple
// internally).
func (x *indexGen) genMultiCol(lm *index.LevelMessage, idx *index.LevelIndex, ordered bool, cursor, indent int, itemVar string, parts []string) {
	if cursor >= len(idx.ColFields) {
		x.g.P(helper.Indent(indent), "const keyParts = [", strings.Join(parts, ", "), "];")
		x.emitPushAll(lm, idx, ordered, indent, "keyParts", itemVar)
		return
	}
	field := idx.ColFields[cursor]
	access, isList := x.fieldAccess(itemVar, field)
	if isList {
		loopVar := fmt.Sprintf("indexItem%d", cursor)
		x.g.P(helper.Indent(indent), "for (const ", loopVar, " of ", access, ") {")
		x.genMultiCol(lm, idx, ordered, cursor+1, indent+1, itemVar, append(append([]string{}, parts...), loopVar))
		x.g.P(helper.Indent(indent), "}")
	} else {
		x.genMultiCol(lm, idx, ordered, cursor+1, indent, itemVar, append(append([]string{}, parts...), access))
	}
}

// emitPushAll appends itemVar into the global container and every leveled
// container, bucketed under keyExpr. Multi-column indexes store a value list
// per key tuple in a leaf TupleKeyMap (get-or-create the list via getOrSet,
// then push); single-column indexes push the raw key into a native Map's value
// list. Leveled containers are first scoped to the upper map key(s): a single
// upper key (i==1) indexes a native Map by its raw value, while composite upper
// keys (i>=2) get-or-create the inner container in a TupleKeyMap via getOrSet.
func (x *indexGen) emitPushAll(lm *index.LevelMessage, idx *index.LevelIndex, ordered bool, indent int, keyExpr, itemVar string) {
	multi := len(idx.ColFields) > 1
	if multi {
		x.g.P(helper.Indent(indent), "this.", x.containerField(idx, ordered, 0), ".getOrSet(", keyExpr, ", () => []).push(", itemVar, ");")
	} else {
		x.emitPushInto(indent, "this."+x.containerField(idx, ordered, 0), keyExpr, itemVar)
	}
	for i := 1; i < lm.LeveledContainerDepth(); i++ {
		container := x.containerField(idx, ordered, i)
		x.g.P(helper.Indent(indent), "{")
		if i == 1 {
			// Single upper key: native Map keyed by the raw k1 value.
			if multi {
				x.g.P(helper.Indent(indent+1), "let m = this.", container, ".get(k1);")
				x.g.P(helper.Indent(indent+1), "if (!m) { m = ", x.newLeafContainer(idx, ordered), "; this.", container, ".set(k1, m); }")
				x.g.P(helper.Indent(indent+1), "m.getOrSet(", keyExpr, ", () => []).push(", itemVar, ");")
			} else {
				x.g.P(helper.Indent(indent+1), "let map = this.", container, ".get(k1);")
				x.g.P(helper.Indent(indent+1), "if (!map) { map = new Map(); this.", container, ".set(k1, map); }")
				x.emitPushInto(indent+1, "map", keyExpr, itemVar)
			}
		} else {
			// Composite upper keys: TupleKeyMap keyed by the [k1..ki] tuple.
			var ks []string
			for j := 1; j <= i; j++ {
				ks = append(ks, fmt.Sprintf("k%d", j))
			}
			tuple := "[" + strings.Join(ks, ", ") + "]"
			if multi {
				x.g.P(helper.Indent(indent+1), "const m = this.", container, ".getOrSet(", tuple, ", () => ", x.newLeafContainer(idx, ordered), ");")
				x.g.P(helper.Indent(indent+1), "m.getOrSet(", keyExpr, ", () => []).push(", itemVar, ");")
			} else {
				x.g.P(helper.Indent(indent+1), "const map = this.", container, ".getOrSet(", tuple, ", () => new Map());")
				x.emitPushInto(indent+1, "map", keyExpr, itemVar)
			}
		}
		x.g.P(helper.Indent(indent), "}")
	}
}

// emitPushInto pushes itemVar into the value list bucketed under keyExpr in the
// given map expression.
func (x *indexGen) emitPushInto(indent int, mapExpr, keyExpr, itemVar string) {
	x.g.P(helper.Indent(indent), "{")
	x.g.P(helper.Indent(indent+1), "const list = ", mapExpr, ".get(", keyExpr, ");")
	x.g.P(helper.Indent(indent+1), "if (list) { list.push(", itemVar, "); } else { ", mapExpr, ".set(", keyExpr, ", [", itemVar, "]); }")
	x.g.P(helper.Indent(indent), "}")
}

// fieldAccess builds the TypeScript access expression for a level field rooted
// at itemVar, returning the expression and whether the (terminal) field is a
// list (which the caller iterates).
func (x *indexGen) fieldAccess(itemVar string, field *index.LevelField) (string, bool) {
	expr := itemVar
	needEmpty := len(field.LeveledFDList) > 1
	isTimestamp := false
	for i, fd := range field.LeveledFDList {
		sep := "."
		if i != 0 {
			sep = "?."
		}
		expr += sep + helper.FieldLocalName(fd)
		if i == len(field.LeveledFDList)-1 && fd.Message() != nil {
			switch fd.Message().FullName() {
			case "google.protobuf.Timestamp", "google.protobuf.Duration":
				expr += "?.seconds ?? 0n"
				isTimestamp = true
				needEmpty = false
			}
		}
	}
	if field.FD.IsList() {
		return expr + " ?? []", true
	}
	if needEmpty && !isTimestamp {
		expr += " ?? " + helper.TSEmptyValue(field.FD)
	}
	return expr, false
}

// genValueListSorter emits a comparator for an index's sorted columns and sorts
// every value list (global and leveled) with it.
func (x *indexGen) genValueListSorter(lm *index.LevelMessage, idx *index.LevelIndex, ordered bool) {
	if len(idx.SortedColFields) == 0 {
		return
	}
	label := "Index"
	if ordered {
		label = "OrderedIndex"
	}
	v := x.valueType(idx)
	cmp := "cmp" + idx.Name()
	var aParts, bParts []string
	for _, field := range idx.SortedColFields {
		a, _ := x.fieldAccess("a", field)
		b, _ := x.fieldAccess("b", field)
		aParts = append(aParts, a)
		bParts = append(bParts, b)
	}
	x.g.P(helper.Indent(2), "// ", label, "(sort): ", idx.Index)
	x.g.P(helper.Indent(2), "const ", cmp, " = (a: ", v, ", b: ", v, "): number => compareTuples([", strings.Join(aParts, ", "), "], [", strings.Join(bParts, ", "), "]);")
	x.g.P(helper.Indent(2), "for (const list of this.", x.containerField(idx, ordered, 0), ".values()) {")
	x.g.P(helper.Indent(3), "list.sort(", cmp, ");")
	x.g.P(helper.Indent(2), "}")
	for i := 1; i < lm.LeveledContainerDepth(); i++ {
		x.g.P(helper.Indent(2), "for (const m of this.", x.containerField(idx, ordered, i), ".values()) {")
		x.g.P(helper.Indent(3), "for (const list of m.values()) {")
		x.g.P(helper.Indent(4), "list.sort(", cmp, ");")
		x.g.P(helper.Indent(3), "}")
		x.g.P(helper.Indent(2), "}")
	}
}

// genOrderedRebuild rebuilds an ordered index's global and leveled maps so they
// are sorted by key. Multi-column indexes sort the TupleKeyMap in place by its
// key tuples (sortKeys); single-column indexes re-sort the native Map by key.
func (x *indexGen) genOrderedRebuild(lm *index.LevelMessage, idx *index.LevelIndex) {
	multi := len(idx.ColFields) > 1
	c0 := x.containerField(idx, true, 0)
	if multi {
		x.g.P(helper.Indent(2), "this.", c0, ".sortKeys();")
		for i := 1; i < lm.LeveledContainerDepth(); i++ {
			ci := x.containerField(idx, true, i)
			x.g.P(helper.Indent(2), "for (const m of this.", ci, ".values()) {")
			x.g.P(helper.Indent(3), "m.sortKeys();")
			x.g.P(helper.Indent(2), "}")
		}
		return
	}
	x.g.P(helper.Indent(2), "this.", c0, " = sortMapByKey(this.", c0, ", compareValues);")
	for i := 1; i < lm.LeveledContainerDepth(); i++ {
		ci := x.containerField(idx, true, i)
		x.g.P(helper.Indent(2), "for (const m of this.", ci, ".values()) {")
		x.g.P(helper.Indent(3), "sortMapByKey(m, compareValues);")
		x.g.P(helper.Indent(2), "}")
	}
}

// ----------------------------------------------------------------------------
// Getters / finders
// ----------------------------------------------------------------------------

// GenGetters emits the ordered-map getter and index/ordered-index finders.
func (x *indexGen) GenGetters() {
	if x.needOrderedMap() {
		x.genOrderedMapGetters(x.message.Desc, 1, nil)
	}
	if x.needIndex() {
		x.eachIndex(false, func(lm *index.LevelMessage, idx *index.LevelIndex) {
			x.genFinders(lm, idx, false)
		})
	}
	if x.needOrderedIndex() {
		x.eachIndex(true, func(lm *index.LevelMessage, idx *index.LevelIndex) {
			x.genFinders(lm, idx, true)
		})
	}
}

// orderedMapGetterName returns the getter name for the depth-level ordered map:
// getOrderedMap for the 1st level, getOrderedMap1/2/... for deeper levels (the
// suffix is the number of upper keys needed to reach it), mirroring the Go /
// C++ / C# GetOrderedMap / GetOrderedMap1 naming.
func orderedMapGetterName(depth int) string {
	if depth == 1 {
		return "getOrderedMap"
	}
	return fmt.Sprintf("getOrderedMap%d", depth-1)
}

// genOrderedMapGetters recursively emits the ordered-map getters along the
// first-map-field chain. The 1st-level getter returns the whole ordered map;
// each deeper getter takes the accumulated upper keys and descends one level by
// looking the next key up and following the node's .first sub-map, returning
// undefined if any key is absent (matching the TS finder convention).
func (x *indexGen) genOrderedMapGetters(md protoreflect.MessageDescriptor, depth int, keys helper.MapKeySlice) {
	fd := firstMapField(md)
	if fd == nil {
		return
	}
	mapType := x.orderedMapTypeOf(fd)
	getter := orderedMapGetterName(depth)
	ordinal := loadutil.Ordinal(depth)
	x.g.P()
	if depth == 1 {
		x.g.P(helper.Indent(1), "/** ", getter, " returns the 1st-level ordered map (sorted by key). */")
		x.g.P(helper.Indent(1), getter, "(): ", mapType, " {")
		x.g.P(helper.Indent(2), "return this.", orderedMapField, ";")
		x.g.P(helper.Indent(1), "}")
	} else {
		params := keys.GenGetParams()
		lastKey := keys[len(keys)-1].Name
		x.g.P(helper.Indent(1), "/** ", getter, " returns the ", ordinal, "-level ordered map scoped to the given upper key(s), or undefined. */")
		x.g.P(helper.Indent(1), getter, "(", params, "): ", mapType, " | undefined {")
		if depth == 2 {
			x.g.P(helper.Indent(2), "return this.", orderedMapField, ".get(", lastKey, ")?.first;")
		} else {
			prevArgs := keys[:len(keys)-1].GenGetArguments()
			x.g.P(helper.Indent(2), "return this.", orderedMapGetterName(depth-1), "(", prevArgs, ")?.get(", lastKey, ")?.first;")
		}
		x.g.P(helper.Indent(1), "}")
	}
	nextKeys := keys.AddMapKey(helper.ParseMapKey(fd.MapKey(), helper.ParseMapFieldNameAsFuncParam(fd)))
	if fd.MapValue().Kind() == protoreflect.MessageKind {
		x.genOrderedMapGetters(fd.MapValue().Message(), depth+1, nextKeys)
	}
}

func (x *indexGen) genFinders(lm *index.LevelMessage, idx *index.LevelIndex, ordered bool) {
	kind := "index"
	if ordered {
		kind = "ordered index"
	}
	name := idx.Name()
	v := x.valueType(idx)
	at := x.indexType(idx, ordered)
	keys := x.params(idx)
	params := keys.GenGetParams()
	args := keys.GenGetArguments()

	// keyExpr is the key passed to a container's get(): a raw value / native-Map
	// key for single-column, or a key-column tuple for the multi-column
	// TupleKeyMap (which serializes it internally).
	keyExpr := args
	multi := len(idx.ColFields) > 1
	if multi {
		keyExpr = "[" + args + "]"
	}

	container0 := x.containerField(idx, ordered, 0)

	// The map finders return the global / leveled container directly: a native
	// Map for single-column indexes, or the owning TupleKeyMap (which exposes
	// readable tuple keys) for multi-column indexes.
	x.g.P()
	x.g.P(helper.Indent(1), "/** find", name, "Map returns the ", kind, " map: key(", idx.Index, ") -> values. */")
	x.g.P(helper.Indent(1), "find", name, "Map(): ", at, " {")
	x.g.P(helper.Indent(2), "return this.", container0, ";")
	x.g.P(helper.Indent(1), "}")

	x.g.P()
	x.g.P(helper.Indent(1), "/** find", name, " returns all values for the given key(s), or undefined. */")
	x.g.P(helper.Indent(1), "find", name, "(", params, "): ", v, "[] | undefined {")
	x.g.P(helper.Indent(2), "return this.", container0, ".get(", keyExpr, ");")
	x.g.P(helper.Indent(1), "}")

	x.g.P()
	x.g.P(helper.Indent(1), "/** findFirst", name, " returns the first value for the given key(s), or undefined. */")
	x.g.P(helper.Indent(1), "findFirst", name, "(", params, "): ", v, " | undefined {")
	x.g.P(helper.Indent(2), "return this.find", name, "(", args, ")?.[0];")
	x.g.P(helper.Indent(1), "}")

	for i := 1; i < lm.LeveledContainerDepth(); i++ {
		container := x.containerField(idx, ordered, i)
		upperKeys := x.keys[:i]
		upperParams := upperKeys.GenGetParams()
		upperArgs := upperKeys.GenGetArguments()
		// A single upper key indexes a native Map by its raw value; composite
		// upper keys index a TupleKeyMap by the upper key tuple.
		upperKeyExpr := upperArgs
		if i > 1 {
			upperKeyExpr = "[" + upperArgs + "]"
		}
		ordinal := loadutil.Ordinal(i)

		x.g.P()
		x.g.P(helper.Indent(1), "/** find", name, "Map", i, " returns the ", kind, " map scoped to the upper ", ordinal, "-level map key(s). */")
		x.g.P(helper.Indent(1), "find", name, "Map", i, "(", upperParams, "): ", at, " | undefined {")
		x.g.P(helper.Indent(2), "return this.", container, ".get(", upperKeyExpr, ");")
		x.g.P(helper.Indent(1), "}")

		x.g.P()
		x.g.P(helper.Indent(1), "/** find", name, i, " returns all values for the given key(s) within the upper ", ordinal, "-level map. */")
		x.g.P(helper.Indent(1), "find", name, i, "(", upperParams, ", ", params, "): ", v, "[] | undefined {")
		x.g.P(helper.Indent(2), "return this.find", name, "Map", i, "(", upperArgs, ")?.get(", keyExpr, ");")
		x.g.P(helper.Indent(1), "}")

		x.g.P()
		x.g.P(helper.Indent(1), "/** findFirst", name, i, " returns the first value for the given key(s) within the upper ", ordinal, "-level map. */")
		x.g.P(helper.Indent(1), "findFirst", name, i, "(", upperParams, ", ", params, "): ", v, " | undefined {")
		x.g.P(helper.Indent(2), "return this.find", name, i, "(", upperArgs, ", ", args, ")?.[0];")
		x.g.P(helper.Indent(1), "}")
	}
}

// collectIndexPackages registers the owning packages of every enum/message
// referenced by an index key or value, across all levels, so their barrel
// imports are emitted.
func collectIndexPackages(descriptor *index.IndexDescriptor, pkgs *pbPackages) {
	addFields := func(idx *index.LevelIndex) {
		pkgs.add(string(idx.MD.ParentFile().Package()))
		fields := append(append([]*index.LevelField{}, idx.ColFields...), idx.SortedColFields...)
		for _, field := range fields {
			// Only enum key fields reference a generated named type. Message
			// key fields are exclusively Timestamp/Duration, which are encoded
			// as bigint (via .seconds) and never reference the well-known type
			// module (protobuf-es serves those from @bufbuild/protobuf/wkt, not
			// a generated google/protobuf/*_pb module), so they add no package.
			if field.FD.Kind() == protoreflect.EnumKind {
				pkgs.add(string(field.FD.Enum().ParentFile().Package()))
			}
		}
	}
	for lm := descriptor.LevelMessage; lm != nil; lm = lm.NextLevel {
		for _, idx := range lm.Indexes {
			addFields(idx)
		}
		for _, idx := range lm.OrderedIndexes {
			addFields(idx)
		}
	}
}
