package main

import (
	"flag"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const version = "0.3.1"

var namespace *string
var messagerSuffix *string

func main() {
	var flags flag.FlagSet
	namespace = flags.String("namespace", "tableau", "tableau namespace")
	messagerSuffix = flags.String("suffix", "Mgr", "tableau messager name suffix")
	flag.Parse()

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
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
		generateRegistry(gen)
		generateHub(gen)
		return nil
	})
}
