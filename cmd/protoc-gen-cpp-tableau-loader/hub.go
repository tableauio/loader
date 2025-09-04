package main

import (
	"strconv"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/extensions"
	"google.golang.org/protobuf/compiler/protogen"
)

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"toSnake": strcase.ToSnake,
}).ParseFS(efs, "embed/templates/*"))

// generateHub generates related hub files.
func generateHub(gen *protogen.Plugin) {
	protofiles := helper.ParseProtoFiles(gen)
	// detect real shard num
	realShardNum := *shards
	if realShardNum > len(protofiles) {
		realShardNum = len(protofiles)
	}
	if realShardNum <= 1 {
		realShardNum = 0
	}
	shards := make([]int, realShardNum)
	for i := range shards {
		shards[i] = i
	}
	type Param struct {
		Shards     []int
		Protofiles helper.ProtoFiles
	}
	params := &Param{Shards: shards, Protofiles: protofiles}
	// generate hub hpp
	hppFilename := "hub." + extensions.PC + ".h"
	g1 := gen.NewGeneratedFile(hppFilename, "")
	helper.GenerateCommonHeader(gen, g1, version)
	g1.P()
	if err := tpl.Lookup(hppFilename+".tpl").Execute(g1, params); err != nil {
		panic(err)
	}
	// generate hub cpp
	cppFilename := "hub." + extensions.PC + ".cc"
	g2 := gen.NewGeneratedFile(cppFilename, "")
	helper.GenerateCommonHeader(gen, g2, version)
	g2.P()
	if err := tpl.Lookup(cppFilename+".tpl").Execute(g2, params); err != nil {
		panic(err)
	}
	// generate shards
	for i, shard := range splitShards(protofiles, realShardNum) {
		type Param struct {
			Shard      int
			Protofiles helper.ProtoFiles
		}
		params := &Param{Shard: i, Protofiles: shard}
		cppTplname := "hub_shard" + "." + extensions.PC + ".cc"
		cppFilename := "hub_shard" + strconv.Itoa(i) + "." + extensions.PC + ".cc"
		g := gen.NewGeneratedFile(cppFilename, "")
		helper.GenerateCommonHeader(gen, g, version)
		g.P()
		if err := tpl.Lookup(cppTplname+".tpl").Execute(g, params); err != nil {
			panic(err)
		}
	}
}

func splitShards(protofiles helper.ProtoFiles, shardNum int) []helper.ProtoFiles {
	if shardNum <= 1 {
		// no need to split
		return nil
	} else {
		cursor := 0
		shards := []helper.ProtoFiles{}
		for i := 0; i < shardNum; i++ {
			shardSize := len(protofiles) / shardNum
			if i < len(protofiles)%shardNum {
				shardSize++
			}
			begin := cursor
			end := cursor + shardSize
			shards = append(shards, protofiles[begin:end])
			cursor = end
		}
		return shards
	}
}
