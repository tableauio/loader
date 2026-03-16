package xproto

import (
	"errors"
	"sort"

	"github.com/tableauio/loader/internal/options"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ProtoFiles is a list of parsed proto files with their messager names.
type ProtoFiles []*ProtoFile

// ProtoFile represents a parsed proto file with its name prefix and messager names.
type ProtoFile struct {
	Name      string
	Messagers []string
}

// ParseProtoFiles parses all proto files from the protogen plugin,
// and returns sorted ProtoFiles (sorted by file name, and messagers
// within each file are also sorted).
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

// FlatMessagers returns a flat sorted list of all messager names
// across all proto files. The order is determined by file order first,
// then messager order within each file.
func (pfs ProtoFiles) FlatMessagers() []string {
	var messagers []string
	for _, pf := range pfs {
		messagers = append(messagers, pf.Messagers...)
	}
	return messagers
}

func (pfs ProtoFiles) SplitShards(shardNum int) []ProtoFiles {
	if shardNum <= 1 {
		// no need to split
		return nil
	} else {
		cursor := 0
		shards := []ProtoFiles{}
		for i := 0; i < shardNum; i++ {
			shardSize := len(pfs) / shardNum
			if i < len(pfs)%shardNum {
				shardSize++
			}
			begin := cursor
			end := cursor + shardSize
			shards = append(shards, pfs[begin:end])
			cursor = end
		}
		return shards
	}
}
