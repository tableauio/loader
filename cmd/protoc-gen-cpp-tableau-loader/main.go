package main

import (
	"flag"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.4.7"
const pcExt = "pc" // protoconf file extension
const pbExt = "pb" // protobuf file extension

var namespace *string
var messagerSuffix *string

// genMode can control only the `registry` code, or only generate the `messager` code
// To avoid each change need a fully regenerated, for better dependency management
var genMode *string

// Available gen modes
const (
	ModeDefault  = "default"  // generate all at once.
	ModeRegistry = "registry" // only generate "registry.pc.h/cc" files.
	ModeMessager = "messager" // only generate "*.pc.h/cc" files.
)

func main() {
	var flags flag.FlagSet
	namespace = flags.String("namespace", "tableau", "tableau namespace")
	messagerSuffix = flags.String("suffix", "Mgr", "tableau messager name suffix")
	genMode = flags.String("gen-mode", "normal", "gen registry code only")
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

			switch *genMode {
			case ModeMessager:
				fallthrough
			case ModeDefault:
				generateMessager(gen, f)
			case ModeRegistry:
				recordFileAndMessagers(gen, f)
			}
		}

		switch *genMode {
		case ModeMessager:
			// skip common code generation
		case ModeDefault:
			generateRegistry(gen)
			generateEmbed(gen)
		case ModeRegistry:
			generateRegistry(gen)
		}
		return nil
	})
}
