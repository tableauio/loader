package helper

import (
	"errors"
	"sort"

	"github.com/tableauio/loader/internal/options"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type ProtoFiles []*ProtoFile

type ProtoFile struct {
	Name      string
	Messagers []string
}

func ParseProtoFiles(gen *protogen.Plugin) ProtoFiles {
	var protofiles []*ProtoFile
	for _, f := range gen.Files {
		if !options.NeedGenFile(f) {
			continue
		}
		opts := f.Desc.Options().(*descriptorpb.FileOptions)
		workbook := proto.GetExtension(opts, tableaupb.E_Workbook).(*tableaupb.WorkbookOptions)
		if workbook == nil {
			continue
		}
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
		protofiles = append(protofiles, &ProtoFile{
			Name:      f.GeneratedFilenamePrefix,
			Messagers: messagers,
		})
	}
	// sort all files to keep in order
	sort.Slice(protofiles, func(i, j int) bool {
		return protofiles[i].Name < protofiles[j].Name
	})
	return protofiles
}
