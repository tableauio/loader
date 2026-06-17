package main

import (
	"flag"
	"fmt"

	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.1.0"

// pbImportPath is the relative path from the loader output dir to the
// protobuf-es generated base output dir. It is used to build the imports of
// the base message modules (e.g. "<pbImportPath>/<prefix>_pb.js") and the
// tableau extension descriptors. Configurable via the "pb_path" plugin option;
// defaults to ".." (loader output nested one level under the base output).
var pbImportPath = ".."

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-ts-tableau-loader %v\n", version)
		return
	}

	var flags flag.FlagSet
	flags.StringVar(&pbImportPath, "pb_path", "..", "relative path from the loader output dir to the protobuf-es base output dir")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL | pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
		gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
		gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024
		reg := newBarrelRegistry(gen)
		for _, f := range gen.Files {
			if !options.NeedGenFile(f) {
				continue
			}
			generateMessager(gen, f, reg)
		}
		generateHub(gen)
		generateEmbed(gen)
		generateBarrels(gen, reg)
		return nil
	})
}
