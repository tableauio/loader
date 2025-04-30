package main

import (
	"flag"

	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.1.0"

func main() {
	var flags flag.FlagSet
	flag.Parse()

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !options.NeedGenFile(f) {
				continue
			}
			generateMessager(gen, f)
		}
		generateHub(gen)
		return nil
	})
}
