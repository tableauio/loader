package main

import (
	"flag"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.4.0"

var pkg *string

func main() {
	var flags flag.FlagSet
	pkg = flags.String("pkg", "tableau", "tableau package name")
	flag.Parse()

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			opts := f.Desc.Options().(*descriptorpb.FileOptions)
			workbook := proto.GetExtension(opts, tableaupb.E_Workbook).(*tableaupb.WorkbookOptions)
			if workbook == nil {
				continue
			}

			containsWorksheet := false
			for _, message := range f.Messages {
				opts := message.Desc.Options().(*descriptorpb.MessageOptions)
				worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
				if worksheet != nil {
					containsWorksheet = true
					break
				}
			}
			if !containsWorksheet {
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
