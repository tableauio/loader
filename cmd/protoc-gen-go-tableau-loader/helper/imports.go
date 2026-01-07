package helper

import "google.golang.org/protobuf/compiler/protogen"

const (
	FormatPackage  = protogen.GoImportPath("github.com/tableauio/tableau/format")
	LoadPackage    = protogen.GoImportPath("github.com/tableauio/tableau/load")
	StorePackage   = protogen.GoImportPath("github.com/tableauio/tableau/store")
	TreeMapPackage = protogen.GoImportPath("github.com/tableauio/loader/pkg/treemap")
	PairPackage    = protogen.GoImportPath("github.com/tableauio/loader/pkg/pair")
	TimePackage    = protogen.GoImportPath("time")
	SortPackage    = protogen.GoImportPath("sort")
	FmtPackage     = protogen.GoImportPath("fmt")
	ProtoPackage   = protogen.GoImportPath("google.golang.org/protobuf/proto")
)
