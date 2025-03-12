package main

import (
	"errors"
	"math"
	"sort"
	"strconv"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func getAllOrderedFilesAndMessagers(gen *protogen.Plugin) (protofiles []string, fileMessagers map[string][]string) {
	fileMessagers = map[string][]string{}
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		opts := f.Desc.Options().(*descriptorpb.FileOptions)
		workbook := proto.GetExtension(opts, tableaupb.E_Workbook).(*tableaupb.WorkbookOptions)
		if workbook == nil {
			continue
		}
		protofiles = append(protofiles, f.GeneratedFilenamePrefix)
		var messagers []string
		for _, message := range f.Messages {
			opts, ok := message.Desc.Options().(*descriptorpb.MessageOptions)
			if !ok {
				gen.Error(errors.New("get message options failed"))
			}
			worksheet, ok := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
			if !ok {
				gen.Error(errors.New("get worksheet extension failed"))
			}
			if worksheet != nil {
				messagerName := string(message.Desc.Name())
				messagers = append(messagers, messagerName)
			}
		}
		// sort messagers in one file to keep in order
		sort.Strings(messagers)
		fileMessagers[f.GeneratedFilenamePrefix] = messagers
	}
	// sort all files to keep in order
	sort.Strings(protofiles)
	return
}

// generateRegistry generates related registry files.
func generateRegistry(gen *protogen.Plugin) {
	if *registryShards <= 1 {
		hppFilename := "registry." + pcExt + ".h"
		g1 := gen.NewGeneratedFile(hppFilename, "")
		helper.GenerateCommonHeader(gen, g1, version)
		g1.P()
		g1.P(registryHppTop)
		g1.P(registryHppBottom)

		protofiles, fileMessagers := getAllOrderedFilesAndMessagers(gen)
		cppFilename := "registry." + pcExt + ".cc"
		g2 := gen.NewGeneratedFile(cppFilename, "")
		helper.GenerateCommonHeader(gen, g2, version)
		g2.P()
		generateRegistryCppFileContent(gen, g2, protofiles, fileMessagers)
	} else {
		// sharding
		generateShardedRegistry(gen)
	}
}

// generateRegistryCppFileContent generates the registry logic.
func generateRegistryCppFileContent(gen *protogen.Plugin, g *protogen.GeneratedFile, protofiles []string, fileMessagers map[string][]string) {
	g.P(`#include "`, "registry.", pcExt, `.h"`)
	g.P()
	for _, proto := range protofiles {
		g.P(`#include "`, proto, ".", pcExt, `.h"`)
	}
	g.P()

	g.P("namespace ", *namespace, " {")
	g.P("Registrar Registry::registrar = Registrar();")
	g.P("void Registry::Init() {")
	for _, messagers := range fileMessagers {
		for _, messager := range messagers {
			g.P("  Register<", messager, ">();")
		}
	}
	g.P("}")
	g.P("}  // namespace ", *namespace)
}

// generateShardedRegistry generates related registry files.
func generateShardedRegistry(gen *protogen.Plugin) {
	protofiles, fileMessagers := getAllOrderedFilesAndMessagers(gen)
	hppFilename := "registry." + pcExt + ".h"
	g1 := gen.NewGeneratedFile(hppFilename, "")
	helper.GenerateCommonHeader(gen, g1, version)
	g1.P()
	g1.P(registryHppTop)

	cppFilename := "registry." + pcExt + ".cc"
	g2 := gen.NewGeneratedFile(cppFilename, "")
	helper.GenerateCommonHeader(gen, g2, version)
	g2.P()
	g2.P(`#include "`, "registry.", pcExt, `.h"`)
	g2.P()

	g2.P("namespace ", *namespace, " {")
	g2.P("Registrar Registry::registrar = Registrar();")
	g2.P("void Registry::Init() {")

	shardSize := int(math.Ceil(float64(len(protofiles)) / float64((*registryShards))))
	for i := 0; i < *registryShards; i++ {
		cursor := (i + 1) * shardSize
		if cursor > len(protofiles) {
			cursor = len(protofiles)
		}
		if i*shardSize >= cursor {
			break
		}
		g1.P("  static void InitShard", i, "();")
		g2.P("  InitShard", i, "();")
		cppFilename := "registry_shard" + strconv.Itoa(i) + "." + pcExt + ".cc"
		g := gen.NewGeneratedFile(cppFilename, "")
		helper.GenerateCommonHeader(gen, g, version)
		g.P()
		shardedProtofiles := protofiles[i*shardSize : cursor]
		generateShardedRegistryCppFileContent(gen, g, i, shardedProtofiles, fileMessagers)
	}

	g1.P(registryHppBottom)
	g2.P("}")
	g2.P("}  // namespace ", *namespace)
}

// generateShardedRegistryCppFileContent generates one registry shard logic.
func generateShardedRegistryCppFileContent(gen *protogen.Plugin, g *protogen.GeneratedFile, shardIndex int, protofiles []string, fileMessagers map[string][]string) {
	g.P(`#include "`, "registry.", pcExt, `.h"`)
	g.P()
	for _, proto := range protofiles {
		g.P(`#include "`, proto, ".", pcExt, `.h"`)
	}
	g.P()

	g.P("namespace ", *namespace, " {")
	g.P("void Registry::InitShard", shardIndex, "() {")
	for _, proto := range protofiles {
		messagers := fileMessagers[proto]
		for _, messager := range messagers {
			g.P("  Register<", messager, ">();")
		}
	}

	g.P("}")
	g.P("}  // namespace ", *namespace)
}

const registryHppTop = `#pragma once
#include "hub.pc.h"
namespace tableau {
using MessagerGenerator = std::function<std::shared_ptr<Messager>()>;
// messager name -> messager generator
using Registrar = std::unordered_map<std::string, MessagerGenerator>;
class Registry {
 public:
  static void Init();`

const registryHppBottom = `
  template <typename T>
  static void Register();

  static Registrar registrar;
};

template <typename T>
void Registry::Register() {
  registrar[T::Name()] = []() { return std::make_shared<T>(); };
}

}  // namespace tableau`
