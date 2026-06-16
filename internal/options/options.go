package options

import (
	"slices"
	"strings"

	"github.com/tableauio/tableau/proto/tableaupb"
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
	LangCS  Language = "cs"
	LangTS  Language = "ts"
)

// GetWorksheetOptions returns the worksheet options of the message descriptor.
// It returns nil if the message is not a tableau worksheet.
func GetWorksheetOptions(md protoreflect.MessageDescriptor) *tableaupb.WorksheetOptions {
	opts := md.Options().(*descriptorpb.MessageOptions)
	return proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
}

// IsWorksheet reports whether the message is a tableau worksheet, i.e. it has
// the tableaupb.E_Worksheet extension set.
func IsWorksheet(md protoreflect.MessageDescriptor) bool {
	return GetWorksheetOptions(md) != nil
}

func NeedGenOrderedMap(md protoreflect.MessageDescriptor, lang Language) bool {
	wsOpts := GetWorksheetOptions(md)
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
	wsOpts := GetWorksheetOptions(md)
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
	wsOpts := GetWorksheetOptions(md)
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
		if IsWorksheet(message.Desc) {
			return true
		}
	}
	return false
}
