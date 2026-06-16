package main

import (
	"fmt"

	"github.com/tableauio/loader/cmd/protoc-gen-ts-tableau-loader/helper"
	"github.com/tableauio/loader/internal/index"
	"github.com/tableauio/loader/internal/loadutil"
	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// generateMessager generates a loader file corresponding to the protobuf file.
// Each wrapped class extends the Messager base class.
func generateMessager(gen *protogen.Plugin, file *protogen.File, reg *barrelRegistry) {
	filename := file.GeneratedFilenamePrefix + ".pc.ts"
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, file, g, version)

	var messagers []*protogen.Message
	for _, message := range file.Messages {
		if options.IsWorksheet(message.Desc) {
			messagers = append(messagers, message)
		}
	}

	// Parse index descriptors up-front: they drive both the extra package
	// imports (enum/message index-key types) and whether the index runtime
	// helpers need importing.
	descriptors := make(map[*protogen.Message]*index.IndexDescriptor, len(messagers))
	needIndexRuntime := false
	needOrderedMapValue := false
	for _, message := range messagers {
		desc := index.ParseIndexDescriptor(message.Desc)
		descriptors[message] = desc
		ig := newIndexGen(g, desc, message, nil)
		if ig.NeedGenerate() {
			needIndexRuntime = true
		}
		if ig.needNestedOrderedMap() {
			needOrderedMapValue = true
		}
	}

	// Runtime imports. The index container runtime (TupleKeyMap & comparators)
	// lives alongside Format in util.pc.ts, so it is pulled from the same module
	// and only when this file uses an index / ordered-index / ordered-map
	// container.
	g.P(`import { create } from "@bufbuild/protobuf";`)
	g.P(`import { Messager } from "./messager.pc.js";`)
	if needIndexRuntime {
		if needOrderedMapValue {
			g.P(`import { Format, compareValues, compareTuples, sortMapByKey, TupleKeyMap, type OrderedMapValue } from "./util.pc.js";`)
		} else {
			g.P(`import { Format, compareValues, compareTuples, sortMapByKey, TupleKeyMap } from "./util.pc.js";`)
		}
	} else {
		g.P(`import { Format } from "./util.pc.js";`)
	}
	g.P(`import { loadMessagerInDir, type MessagerOptions } from "./load.pc.js";`)
	// Namespace imports for the protobuf-es generated types, one per proto
	// package: the worksheet's own package plus any other package owning a
	// map-value message/enum referenced by a getter. Each package is imported
	// from its barrel module under an alias equal to the proto package name
	// (e.g. import * as protoconf / import * as base), so the loader qualifies
	// types exactly like the loaders of other languages (protoconf.ItemConf,
	// base.Hero) — consistent cross-language naming, and the alias also sidesteps
	// the wrapper-class vs. message-type name collision.
	pkgs := newPBPackages(file.GeneratedFilenamePrefix)
	pkgs.add(string(file.Desc.Package()))
	for _, message := range messagers {
		collectValuePackages(message.Desc, pkgs)
		collectIndexPackages(descriptors[message], pkgs)
	}
	pkgs.emit(g)
	for _, pkg := range pkgs.order {
		reg.markReferenced(pkg)
	}
	g.P()

	for i, message := range messagers {
		if i > 0 {
			g.P()
		}
		genMessage(g, message, descriptors[message], pkgs)
	}
}

// genMessage generates a single messager class definition.
func genMessage(g *protogen.GeneratedFile, message *protogen.Message, descriptor *index.IndexDescriptor, pkgs *pbPackages) {
	md := message.Desc
	name := string(md.Name())
	alias := pkgs.aliasOf(md)
	schema := alias + "." + helper.LocalSchemaName(md) // runtime schema value
	dataType := alias + "." + helper.LocalTypeName(md) // message type
	idxGen := newIndexGen(g, descriptor, message, pkgs)

	g.P("/**")
	g.P(" * ", name, " is a wrapper around protobuf message ", md.FullName(), ".")
	g.P(" */")
	g.P("export class ", name, " extends Messager {")
	g.P(helper.Indent(1), "#data: ", dataType, " = create(", schema, ");")
	idxGen.GenDecls()
	g.P()

	// name()
	g.P(helper.Indent(1), "/** name returns the ", name, "'s message name. */")
	g.P(helper.Indent(1), "name(): string {")
	g.P(helper.Indent(2), "return ", schema, ".name;")
	g.P(helper.Indent(1), "}")
	g.P()

	// load()
	g.P(helper.Indent(1), "/** load loads ", name, "'s content in the given dir, based on format and messager options. Throws on failure. */")
	g.P(helper.Indent(1), "load(dir: string, fmt: Format, options?: MessagerOptions): void {")
	g.P(helper.Indent(2), "const start = Date.now();")
	g.P(helper.Indent(2), "try {")
	g.P(helper.Indent(3), "this.#data = loadMessagerInDir(", schema, ", dir, fmt, options);")
	g.P(helper.Indent(2), "} catch (e) {")
	g.P(helper.Indent(3), "throw new Error(`failed to load ", name, "`, { cause: e });")
	g.P(helper.Indent(2), "}")
	g.P(helper.Indent(2), "this.loadStats.durationMs = Date.now() - start;")
	g.P(helper.Indent(2), "this.processAfterLoad();")
	g.P(helper.Indent(1), "}")
	g.P()

	// data()
	g.P(helper.Indent(1), "/** data returns the ", name, "'s inner message data. */")
	g.P(helper.Indent(1), "data(): ", dataType, " {")
	g.P(helper.Indent(2), "return this.#data;")
	g.P(helper.Indent(1), "}")
	g.P()

	// message()
	g.P(helper.Indent(1), "/** message returns the ", name, "'s inner message data. */")
	g.P(helper.Indent(1), "override message(): ", dataType, " {")
	g.P(helper.Indent(2), "return this.#data;")
	g.P(helper.Indent(1), "}")

	// processAfterLoad() override: build index / ordered index / ordered map.
	if idxGen.NeedGenerate() {
		g.P()
		g.P(helper.Indent(1), "/** processAfterLoad builds the index, ordered index and ordered map containers. */")
		g.P(helper.Indent(1), "override processAfterLoad(): void {")
		idxGen.GenProcessAfterLoadBody()
		g.P(helper.Indent(1), "}")
	}

	// syntactic sugar for accessing map items
	genMapGetters(g, md, 1, nil, pkgs)

	// index / ordered index finders and the ordered map getter
	idxGen.GenGetters()

	g.P("}")

	// Type aliases for the index / ordered index / ordered map containers,
	// declared in a namespace merged with the class above.
	idxGen.GenTypeAliases()
}

// genMapGetters generates nested map getters (get1/get2/...) for a message.
func genMapGetters(g *protogen.GeneratedFile, md protoreflect.MessageDescriptor, depth int, keys helper.MapKeySlice, pkgs *pbPackages) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if !fd.IsMap() {
			continue
		}
		localName := helper.FieldLocalName(fd)
		paramName := helper.ParseMapFieldNameAsFuncParam(fd)
		keys = keys.AddMapKey(helper.ParseMapKey(fd.MapKey(), paramName))
		last := keys[len(keys)-1]

		// The value type comes straight from the map value: message/enum values
		// use their own package-qualified named type (protoconf.Foo); scalar
		// values use the plain TypeScript type.
		returnType := mapValueType(fd, pkgs) + " | undefined"
		getter := fmt.Sprintf("get%d", depth)

		var access string
		if depth == 1 {
			access = "this.#data." + localName + "[" + last.IndexExpr() + "]"
		} else {
			prevArgs := keys[:len(keys)-1].GenGetArguments()
			access = fmt.Sprintf("this.get%d(%s)?.%s[%s]", depth-1, prevArgs, localName, last.IndexExpr())
		}

		g.P()
		g.P(helper.Indent(1), "/** ", getter, " finds value in the ", loadutil.Ordinal(depth), "-level map; returns undefined if not found. */")
		g.P(helper.Indent(1), getter, "(", keys.GenGetParams(), "): ", returnType, " {")
		g.P(helper.Indent(2), "return ", access, ";")
		g.P(helper.Indent(1), "}")

		if fd.MapValue().Kind() == protoreflect.MessageKind {
			genMapGetters(g, fd.MapValue().Message(), depth+1, keys, pkgs)
		}
		break
	}
}

// mapValueType returns the TypeScript type for a map field's value: a
// package-qualified named type (protoconf.Foo) for message/enum values, or the
// plain scalar type otherwise.
func mapValueType(fd protoreflect.FieldDescriptor, pkgs *pbPackages) string {
	v := fd.MapValue()
	switch v.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return pkgs.aliasOf(v.Message()) + "." + helper.LocalTypeName(v.Message())
	case protoreflect.EnumKind:
		return pkgs.aliasOf(v.Enum()) + "." + helper.LocalTypeName(v.Enum())
	default:
		s, _ := helper.ScalarTSType(v.Kind())
		return s
	}
}

// collectValuePackages walks the same first-map-field chain as genMapGetters and
// registers the owning package of every message/enum map value referenced by a
// getter (including nested messages and those from other proto files), so each
// gets a namespace barrel import.
func collectValuePackages(md protoreflect.MessageDescriptor, pkgs *pbPackages) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if !fd.IsMap() {
			continue
		}
		switch v := fd.MapValue(); v.Kind() {
		case protoreflect.MessageKind, protoreflect.GroupKind:
			pkgs.add(string(v.Message().ParentFile().Package()))
			collectValuePackages(v.Message(), pkgs)
		case protoreflect.EnumKind:
			pkgs.add(string(v.Enum().ParentFile().Package()))
		}
		break
	}
}
