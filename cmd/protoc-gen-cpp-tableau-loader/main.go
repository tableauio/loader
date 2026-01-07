package main

import (
	"flag"
	"fmt"

	"github.com/tableauio/loader/internal/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.11.0"

// specify protobuf namespace
var namespace *string

// specify messager suffix
var messagerSuffix *string

// mode control generating rules for better dependency management.
var mode *string

// count of generated hub cpp files, aimed to boost compiling speed.
var shards *int

const (
	ModeDefault  = "default"  // generate all at once.
	ModeHub      = "hub"      // only generate "hub.pc.h/cc" files.
	ModeMessager = "messager" // only generate "*.pc.h/cc" for each .proto files.
)

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-cpp-tableau-loader %v\n", version)
		return
	}

	var flags flag.FlagSet
	namespace = flags.String("namespace", "tableau", "tableau namespace")
	messagerSuffix = flags.String("suffix", "Mgr", "tableau messager name suffix")
	mode = flags.String("mode", "default", `available mode: default, hub, and messager. 
  - default: generate all at once.
  - hub: only generate "hub.pc.h/cc" files.
  - messager: only generate "*.pc.h/cc" for each .proto files.
`)
	shards = flags.Int("shards", 1, "count of generated hub cpp files for distributed compiling speed-up")

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
			case ModeHub:
				// pass
			case ModeDefault:
				generateMessager(gen, f)
			}
		}

		switch *mode {
		case ModeDefault:
			generateHub(gen)
			generateEmbed(gen)
		case ModeMessager:
			// pass
		case ModeHub:
			generateHub(gen)
		}
		return nil
	})
}
