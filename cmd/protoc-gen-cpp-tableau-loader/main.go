package main

import (
	"flag"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const version = "0.4.6"
const pcExt = "pc" // protoconf file extension
const pbExt = "pb" // protobuf file extension

var namespace *string
var messagerSuffix *string

// genMode 可以控制只生成 registry 代码，或者只生成 loader 代码
// registry 代码依赖输入所有要生成的 proto 文件，而 loader 代码可以单个文件独立生成
// 把 registry 单独拆出来，方便做依赖管理，避免每次改动只能全量生成
var genMode *string

const (
	// normalMode 普通模式，需要一次性输入所有配置协议，重新生成所有代码
	normalMode = "normal"
	// registryOnly 只生成 registry 代码的模式，任意一个协议有变化，registry 代码都会有变化
	registryOnly = "registry_only"
	// loaderOnly 只生成 loader 代码的模式，跳过公共代码生成
	loaderOnly = "loader_only"
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
			case loaderOnly:
				fallthrough
			case normalMode:
				generateMessager(gen, f)
			case registryOnly:
				recordFileAndMessagers(gen, f)
			}
		}

		switch *genMode {
		case loaderOnly:
			// skip common code generation
		case normalMode:
			generateRegistry(gen)
			generateEmbed(gen)
		case registryOnly:
			generateRegistry(gen)
		}
		return nil
	})
}
