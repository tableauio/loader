package options

import (
	"strings"

	"github.com/tableauio/tableau/proto/tableaupb"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	optionOrderedMap   = "OrderedMap"
	optionIndex        = "Index"
	optionOrderedIndex = "OrderedIndex"
)

type Language = string

const (
	LangCPP Language = "cpp"
	LangGO  Language = "go"
)

func NeedGenOrderedMap(md protoreflect.MessageDescriptor, lang Language) bool {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	if !wsOpts.GetOrderedMap() {
		// Not an ordered map.
		return false
	}
	if languages, ok := wsOpts.GetLangOptions()[optionOrderedMap]; ok {
		if !slices.Contains(strings.Split(languages, " "), lang) {
			// Do not generate ordered map for curr language
			return false
		}
	}
	return true
}

func NeedGenIndex(md protoreflect.MessageDescriptor, lang Language) bool {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	if len(wsOpts.GetIndex()) == 0 {
		// No index.
		return false
	}
	if languages, ok := wsOpts.GetLangOptions()[optionIndex]; ok {
		if !slices.Contains(strings.Split(languages, " "), lang) {
			// Do not generate index for curr language
			return false
		}
	}
	return true
}

func NeedGenOrderedIndex(md protoreflect.MessageDescriptor, lang Language) bool {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	if len(wsOpts.GetOrderedIndex()) == 0 {
		// No index.
		return false
	}
	if languages, ok := wsOpts.GetLangOptions()[optionOrderedIndex]; ok {
		if !slices.Contains(strings.Split(languages, " "), lang) {
			// Do not generate index for curr language
			return false
		}
	}
	return true
}

func NeedGenFile(f *protogen.File) bool {
	if !f.Generate {
		return false
	}

	opts := f.Desc.Options().(*descriptorpb.FileOptions)
	workbook := proto.GetExtension(opts, tableaupb.E_Workbook).(*tableaupb.WorkbookOptions)
	if workbook == nil {
		return false
	}

	for _, message := range f.Messages {
		opts := message.Desc.Options().(*descriptorpb.MessageOptions)
		worksheet := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
		if worksheet != nil {
			return true
		}
	}
	return false
}
