package main

import (
	"math"
	"strconv"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
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
	hppFilename := "hub." + pcExt + ".h"
	g1 := gen.NewGeneratedFile(hppFilename, "")
	helper.GenerateCommonHeader(gen, g1, version)
	g1.P()
	if err := tpl.Lookup("hub.pc.h.tpl").Execute(g1, params); err != nil {
		panic(err)
	}
	// generate hub cpp
	cppFilename := "hub." + pcExt + ".cc"
	g2 := gen.NewGeneratedFile(cppFilename, "")
	helper.GenerateCommonHeader(gen, g2, version)
	g2.P()
	if err := tpl.Lookup("hub.pc.cc.tpl").Execute(g2, params); err != nil {
		panic(err)
	}
	// generate shards
	for i := 0; i < realShardNum; i++ {
		shardSize := int(math.Ceil(float64(len(protofiles)) / float64((realShardNum))))
		begin := i * shardSize
		end := (i + 1) * shardSize
		if end > len(protofiles) {
			end = len(protofiles)
		}
		type Param struct {
			Shard      int
			Protofiles helper.ProtoFiles
		}
		params := &Param{Shard: i, Protofiles: protofiles[begin:end]}
		cppFilename := "hub_shard" + strconv.Itoa(i) + "." + pcExt + ".cc"
		g := gen.NewGeneratedFile(cppFilename, "")
		helper.GenerateCommonHeader(gen, g, version)
		g.P()
		if err := tpl.Lookup("hub_shard.pc.cc.tpl").Execute(g, params); err != nil {
			panic(err)
		}
	}
}
