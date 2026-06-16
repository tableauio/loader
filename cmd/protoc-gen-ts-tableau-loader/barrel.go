package main

import (
	"sort"
	"strings"

	"github.com/tableauio/loader/cmd/protoc-gen-ts-tableau-loader/helper"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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
