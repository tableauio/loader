package main

import (
	"flag"
	"fmt"

	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.10.1"

var pkg *string

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-tableau-loader %v\n", version)
		return
	}

	var flags flag.FlagSet
	pkg = flags.String("pkg", "tableau", "tableau package name")

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
		generateError(gen)
		return nil
	})
}
