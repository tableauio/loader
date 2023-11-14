package main

import (
	"errors"
	"sort"

	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func getAllOrderedFilesAndMessagers(gen *protogen.Plugin) (protofiles, messagers []string) {
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
	}
	// sort to keep in order
	sort.Strings(protofiles)
	sort.Strings(messagers)
	return
}

// generateRegistry generates related registry files.
func generateRegistry(gen *protogen.Plugin) {
	hppFilename := "registry." + pcExt + ".h"
	g1 := gen.NewGeneratedFile(hppFilename, "")
	helper.GenerateCommonHeader(gen, g1, version)
	g1.P()
	g1.P(registryHpp)

	cppFilename := "registry." + pcExt + ".cc"
	g2 := gen.NewGeneratedFile(cppFilename, "")
	helper.GenerateCommonHeader(gen, g2, version)
	g2.P()
	generateRegistryCppFileContent(gen, g2)
}

// generateRegistryCppFileContent generates type definitions.
func generateRegistryCppFileContent(gen *protogen.Plugin, g *protogen.GeneratedFile) {
	protofiles, messagers := getAllOrderedFilesAndMessagers(gen)

	g.P(`#include "`, "registry.", pcExt, `.h"`)
	g.P()
	for _, proto := range protofiles {
		g.P(`#include "`, proto, ".", pcExt, `.h"`)
	}
	g.P()

	g.P("namespace ", *namespace, " {")
	g.P("Registrar Registry::registrar = Registrar();")
	g.P("void Registry::Init() {")
	for _, messager := range messagers {
		g.P("  Register<", messager, ">();")
	}
	g.P("}")
	g.P("}  // namespace ", *namespace)
}

const registryHpp = `#pragma once
#include "hub.pc.h"
namespace tableau {
using MessagerGenerator = std::function<std::shared_ptr<Messager>()>;
// messager name -> messager generator
using Registrar = std::unordered_map<std::string, MessagerGenerator>;
class Registry {
 public:
  static void Init();
  template <typename T>
  static void Register();

  static Registrar registrar;
};

template <typename T>
void Registry::Register() {
  registrar[T::Name()] = []() { return std::make_shared<T>(); };
}

}  // namespace tableau`
