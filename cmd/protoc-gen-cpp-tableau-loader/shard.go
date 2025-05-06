package main

import (
	"math"
	"strconv"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"google.golang.org/protobuf/compiler/protogen"
)

// generateShardedHub generates related hub files.
func generateShardedHub(gen *protogen.Plugin) {
	protofiles, fileMessagers := getAllOrderedFilesAndMessagers(gen)

	// detect real shard num
	realShardNum := *shards
	if realShardNum > len(protofiles) {
		realShardNum = len(protofiles)
	}
	shardSize := int(math.Ceil(float64(len(protofiles)) / float64((realShardNum))))

	hppFilename := "hub." + pcExt + ".h"
	g1 := gen.NewGeneratedFile(hppFilename, "")
	helper.GenerateCommonHeader(gen, g1, version)
	g1.P()
	g1.P(hubHpp)
	generateHubHppTplSpec(gen, g1, protofiles, fileMessagers)
	g1.P(msgContainerHpp)
	generateShardedHubHppMsgContainerShards(gen, g1, realShardNum)
	generateHubHppMsgContainerMembers(gen, g1, protofiles, fileMessagers)
	g1.P(registryHpp)
	generateShardedHubHppRegistryShards(gen, g1, realShardNum)
	g1.P(bottomHpp)

	cppFilename := "hub." + pcExt + ".cc"
	g2 := gen.NewGeneratedFile(cppFilename, "")
	helper.GenerateCommonHeader(gen, g2, version)
	g2.P()
	g2.P(hubCppHeader)
	g2.P(hubCpp)
	g2.P(msgContainerCpp)
	generateShardedHubCppMsgContainerShards(gen, g2, realShardNum)
	g2.P(registryCpp)
	generateShardedHubCppRegistryShards(gen, g2, realShardNum)
	g2.P(bottomCpp)

	for i := 0; i < realShardNum; i++ {
		cursor := (i + 1) * shardSize
		if cursor > len(protofiles) {
			cursor = len(protofiles)
		}
		cppFilename := "hub_shard" + strconv.Itoa(i) + "." + pcExt + ".cc"
		g := gen.NewGeneratedFile(cppFilename, "")
		helper.GenerateCommonHeader(gen, g, version)
		g.P()
		shardedProtofiles := protofiles[i*shardSize : cursor]
		generateShardedHubCppFileContent(gen, g, i, shardedProtofiles, fileMessagers)
	}
}

func generateShardedHubHppMsgContainerShards(gen *protogen.Plugin, g *protogen.GeneratedFile, shardNum int) {
	for i := 0; i < shardNum; i++ {
		g.P("  void InitShard", i, "();")
	}
	g.P()
	g.P(" private:")
}

func generateShardedHubHppRegistryShards(gen *protogen.Plugin, g *protogen.GeneratedFile, shardNum int) {
	for i := 0; i < shardNum; i++ {
		g.P("  static void InitShard", i, "();")
	}
	g.P()
	g.P(" private:")
}

func generateShardedHubCppMsgContainerShards(gen *protogen.Plugin, g *protogen.GeneratedFile, shardNum int) {
	for i := 0; i < shardNum; i++ {
		g.P("  InitShard", i, "();")
	}
}

func generateShardedHubCppRegistryShards(gen *protogen.Plugin, g *protogen.GeneratedFile, shardNum int) {
	for i := 0; i < shardNum; i++ {
		g.P("  InitShard", i, "();")
	}
}

func generateShardedHubCppFileContent(gen *protogen.Plugin, g *protogen.GeneratedFile, shardIndex int, protofiles []string, fileMessagers map[string][]string) {
	g.P(`#include "`, "hub.", pcExt, `.h"`)
	g.P()
	for _, proto := range protofiles {
		g.P(`#include "`, proto, ".", pcExt, `.h"`)
	}
	g.P()
	g.P("namespace ", *namespace, " {")

	for _, proto := range protofiles {
		for _, messager := range fileMessagers[proto] {
			g.P("template <>")
			g.P("const std::shared_ptr<" + messager + "> Hub::Get<" + messager + ">() const {")
			g.P("  return GetMessagerContainer()->", strcase.ToSnake(messager), "_;")
			g.P("}")
			g.P()
		}
	}

	g.P("void MessagerContainer::InitShard", shardIndex, "() {")
	for _, proto := range protofiles {
		for _, messager := range fileMessagers[proto] {
			g.P("  ", strcase.ToSnake(messager), "_ = std::dynamic_pointer_cast<", messager, `>((*msger_map_)["`, messager, `"]);`)
		}
	}
	g.P("}")
	g.P()

	g.P("void Registry::InitShard", shardIndex, "() {")
	for _, proto := range protofiles {
		for _, messager := range fileMessagers[proto] {
			g.P("  Register<", messager, ">();")
		}
	}
	g.P("}")

	g.P("}  // namespace ", *namespace)
}
