package main

import (
	"flag"

	"github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader/firstpass"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.2.5"

var pkg *string
var protoconfPkg *string

func main() {
	var flags flag.FlagSet
	pkg = flags.String("pkg", "tableau", "tableau package name")
	protoconfPkg = flags.String("protoconf-pkg", "protoconf", "protoconf package name")
	flag.Parse()

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		firstpass.Init(gen)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			opts := f.Desc.Options().(*descriptorpb.FileOptions)
			workbook := proto.GetExtension(opts, tableaupb.E_Workbook).(*tableaupb.WorkbookOptions)
			if workbook == nil {
				continue
			}
			generateMessager(gen, f)
		}
		generateHub(gen)
		generateError(gen)
		generateCode(gen)
		return nil
	})
}
