package main

import (
	"flag"

	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.7.0"
const pcExt = "pc" // protoconf file extension
const pbExt = "pb" // protobuf file extension

// specify protobuf namespace
var namespace *string

// specify messager suffix
var messagerSuffix *string

// mode control generating rules for better dependency management.
var mode *string

// count of generated registry files, aimed to boost compiling speed.
var registryShards *int

const (
	ModeDefault  = "default"  // generate all at once.
	ModeRegistry = "registry" // only generate "registry.pc.h/cc" files.
	ModeMessager = "messager" // only generate "*.pc.h/cc" for each .proto files.
)

func main() {
	var flags flag.FlagSet
	namespace = flags.String("namespace", "tableau", "tableau namespace")
	messagerSuffix = flags.String("suffix", "Mgr", "tableau messager name suffix")
	mode = flags.String("mode", "default", `available mode: default, registry, and messager. 
  - default: generate all at once.
  - registry: only generate "registry.pc.h/cc" files.
  - messager: only generate "*.pc.h/cc" for each .proto files.
`)
	registryShards = flags.Int("registry-shards", 1, "count of generated registry files")
	flag.Parse()

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !options.NeedGenFile(f) {
				continue
			}

			switch *mode {
			case ModeMessager:
				generateMessager(gen, f)
			case ModeRegistry:
				// pass
			case ModeDefault:
				generateMessager(gen, f)
			}
		}

		switch *mode {
		case ModeDefault:
			generateRegistry(gen)
			generateEmbed(gen)
		case ModeMessager:
			// pass
		case ModeRegistry:
			generateRegistry(gen)
		}
		return nil
	})
}
