package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-ts-tableau-loader/helper"
	"github.com/tableauio/loader/internal/loadutil"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// isWorksheet reports whether a message is a tableau worksheet.
func isWorksheet(message *protogen.Message) bool {
	opts := message.Desc.Options().(*descriptorpb.MessageOptions)
	worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	return worksheet != nil
}

// generateMessager generates a loader file corresponding to the protobuf file.
// Each wrapped class extends the Messager base class.
func generateMessager(gen *protogen.Plugin, file *protogen.File, reg *barrelRegistry) {
	filename := file.GeneratedFilenamePrefix + ".pc.ts"
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, file, g, version)

	var messagers []*protogen.Message
	for _, message := range file.Messages {
		if isWorksheet(message) {
			messagers = append(messagers, message)
		}
	}

	// Runtime imports.
	g.P(`import { create } from "@bufbuild/protobuf";`)
	g.P(`import { Messager } from "./messager.pc.js";`)
	g.P(`import { Format } from "./util.pc.js";`)
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
		genMessage(gen, g, message, pkgs)
	}
}

// genMessage generates a single messager class definition.
func genMessage(gen *protogen.Plugin, g *protogen.GeneratedFile, message *protogen.Message, pkgs *pbPackages) {
	md := message.Desc
	name := helper.MessagerName(md)
	alias := pkgs.aliasOf(md)
	schema := alias + "." + helper.LocalSchemaName(md) // runtime schema value
	dataType := alias + "." + helper.LocalTypeName(md) // message type

	g.P("/**")
	g.P(" * ", name, " is a wrapper around protobuf message ", md.FullName(), ".")
	g.P(" */")
	g.P("export class ", name, " extends Messager {")
	g.P(helper.Indent(1), "private data_: ", dataType, " = create(", schema, ");")
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
	g.P(helper.Indent(3), "this.data_ = loadMessagerInDir(", schema, ", dir, fmt, options);")
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
	g.P(helper.Indent(2), "return this.data_;")
	g.P(helper.Indent(1), "}")
	g.P()

	// message()
	g.P(helper.Indent(1), "/** message returns the ", name, "'s inner message data. */")
	g.P(helper.Indent(1), "override message(): ", dataType, " {")
	g.P(helper.Indent(2), "return this.data_;")
	g.P(helper.Indent(1), "}")

	// syntactic sugar for accessing map items
	genMapGetters(g, md, 1, nil, pkgs)

	g.P("}")
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
			access = "this.data_." + localName + "[" + last.IndexExpr() + "]"
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

// pbPackages tracks the protobuf packages a single loader file imports and
// emits one namespace import per package, pointing at that package's barrel
// module under the barrel/ subdir: `import * as <alias> from
// "<rel>barrel/<alias>.pc.js"`. The alias is the proto package name (dots ->
// underscores), so generated loaders qualify types exactly like the loaders of
// other languages (e.g. protoconf.ItemConf, base.Hero), keeping cross-language
// naming consistent.
type pbPackages struct {
	rel   string          // relative path prefix from this loader to the barrel dir
	order []string        // referenced package names, in first-seen order
	seen  map[string]bool // deduplication set
}

// newPBPackages creates a registry for a loader whose source-relative output
// prefix is loaderPrefix (e.g. "hero_conf" or "sub/foo"). The number of path
// segments determines how many "../" are needed to reach the loader output
// root; barrel modules live in the barrel/ subdir under that root.
func newPBPackages(loaderPrefix string) *pbPackages {
	rel := "./"
	if depth := strings.Count(loaderPrefix, "/"); depth > 0 {
		rel = strings.Repeat("../", depth)
	}
	return &pbPackages{rel: rel, seen: map[string]bool{}}
}

// add registers a proto package (idempotent), preserving first-seen order.
func (p *pbPackages) add(pkg string) {
	if p.seen[pkg] {
		return
	}
	p.seen[pkg] = true
	p.order = append(p.order, pkg)
}

// aliasOf returns the namespace alias for a descriptor's owning package.
func (p *pbPackages) aliasOf(d protoreflect.Descriptor) string {
	return helper.PackageAlias(string(d.ParentFile().Package()))
}

// emit writes one namespace import statement per package, in registration order.
func (p *pbPackages) emit(g *protogen.GeneratedFile) {
	for _, pkg := range p.order {
		alias := helper.PackageAlias(pkg)
		g.P(`import * as `, alias, ` from "`, p.rel, `barrel/`, alias, `.pc.js";`)
	}
}

// moduleOf returns the protobuf-es base module specifier for a descriptor's
// parent file (e.g. "../protoconf/base/base_pb.js"), relative to the loader
// output root.
func moduleOf(d protoreflect.Descriptor) string {
	return pbImportPath + "/" + strings.TrimSuffix(d.ParentFile().Path(), ".proto") + "_pb.js"
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

// barrelRegistry collects, per proto package, all generated files (used to fill
// a package's barrel) and tracks which packages are actually referenced by some
// loader (so only those packages get a barrel emitted).
type barrelRegistry struct {
	filesByPkg map[string][]*protogen.File // package -> all files of that package
	referenced []string                    // referenced packages, in first-seen order
	seen       map[string]bool             // deduplication set for referenced
}

// newBarrelRegistry groups every file known to the generator by proto package.
func newBarrelRegistry(gen *protogen.Plugin) *barrelRegistry {
	r := &barrelRegistry{filesByPkg: map[string][]*protogen.File{}, seen: map[string]bool{}}
	for _, f := range gen.Files {
		pkg := string(f.Desc.Package())
		r.filesByPkg[pkg] = append(r.filesByPkg[pkg], f)
	}
	return r
}

// markReferenced records that a package is imported by some loader.
func (r *barrelRegistry) markReferenced(pkg string) {
	if r.seen[pkg] {
		return
	}
	r.seen[pkg] = true
	r.referenced = append(r.referenced, pkg)
}

// generateBarrels emits one barrel module per referenced proto package into the
// barrel/ subdir of the loader output root. Each barrel re-exports every
// protobuf-es generated module of that package via `export *`, so a loader can
// import the whole package under a single namespace alias. This is
// collision-free because protobuf guarantees fully-qualified names are unique
// within a package, so the wildcard re-exports never clash. Because the barrel
// sits one level under the loader output root, an extra "../" is prepended to
// each module path (which moduleOf computes relative to that root).
func generateBarrels(gen *protogen.Plugin, reg *barrelRegistry) {
	for _, pkg := range reg.referenced {
		alias := helper.PackageAlias(pkg)
		g := gen.NewGeneratedFile("barrel/"+alias+".pc.ts", "")
		helper.GenerateFileHeader(gen, nil, g, version)
		g.P(`// Barrel for proto package "`, pkg, `": re-exports every protobuf-es`)
		g.P("// generated module of this package, so loaders import the whole package")
		g.P("// under one namespace alias (import * as ", alias, `), matching the`)
		g.P("// package-qualified naming used by loaders of other languages.")
		g.P()
		files := append([]*protogen.File(nil), reg.filesByPkg[pkg]...)
		sort.Slice(files, func(i, j int) bool {
			return files[i].Desc.Path() < files[j].Desc.Path()
		})
		for _, f := range files {
			g.P(`export * from "../`, moduleOf(f.Desc), `";`)
		}
	}
}
