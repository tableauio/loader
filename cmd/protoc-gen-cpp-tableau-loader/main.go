package main

import (
	"flag"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const version = "0.4.4"
const pcExt = "pc" // protoconf file extension
const pbExt = "pb" // protobuf file extension

var namespace *string
var messagerSuffix *string

// genRegistryOnly 只生成 registry 代码
// registry 代码依赖输入所有要生成的 proto 文件，而 loader 代码可以单个文件独立生成
// 把 registry 单独拆出来，方便做依赖管理，避免每次改动只能全量生成
var genRegistryOnly *bool

func main() {
	var flags flag.FlagSet
	namespace = flags.String("namespace", "tableau", "tableau namespace")
	messagerSuffix = flags.String("suffix", "Mgr", "tableau messager name suffix")
	genRegistryOnly = flags.Bool("gen-registry-only", false, "gen registry code only")
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

			// 如果只生成 register 代码，则跳过 loader 代码生成的步骤
			if *genRegistryOnly {
				appendMessager(gen, f)
			} else {
				generateMessager(gen, f)
			}
		}

		if *genRegistryOnly {
			generateRegistry(gen)
			return nil
		} else {
			generateRegistry(gen)
			generateEmbed(gen)
		}
		return nil
	})
}
