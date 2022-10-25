package index

import (
	"strings"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Single-column index:
//	- ID
//	- ID@Item
//
// Multi-column index (composite index):
//  - (ID, Name)
//  - (ID, Name)@Item

const colAndNameSep = "@"
const multiColSep = ","
const multiColGroupCutset = "()"

type Index struct {
	Cols []string // column names in CamelCase (single-column or multi-column)
	Name string   // index name in CamelCase
}

func (index *Index) String() string {
	syntax := ""
	for i, col := range index.Cols {
		syntax += col
		if i != len(index.Cols)-1 {
			syntax += ","
		}
	}

	if len(index.Cols) > 1 {
		syntax = "(" + syntax + ")"
	}
	if index.Name != "" {
		syntax += "@" + index.Name
	}
	return syntax
}

// parse worksheet option index
func parseWSOptionIndex(md protoreflect.MessageDescriptor) []*Index {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	return parseIndexFrom(wsOpts.Index)
}

func parseColsFrom(multiColGroup string) []string {
	trimmedStr := strings.Trim(multiColGroup, multiColGroupCutset)
	cols := strings.Split(trimmedStr, multiColSep)
	for i, col := range cols {
		cols[i] = strings.TrimSpace(col)
	}
	return cols
}

func parseIndex(indexStr string) *Index {
	var cols []string
	var name string
	splits := strings.SplitN(indexStr, colAndNameSep, 2)
	switch len(splits) {
	case 1:
		cols = parseColsFrom(splits[0])
	case 2:
		cols = parseColsFrom(splits[0])
		name = splits[1]
	default:
		return nil
	}

	return &Index{
		Cols: cols,
		Name: name,
	}
}

func parseIndexFrom(indexList []string) []*Index {
	indexes := []*Index{}
	for _, indexStr := range indexList {
		index := parseIndex(indexStr)
		if index != nil {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func NeedGenIndex(md protoreflect.MessageDescriptor) bool {
	indexes := parseWSOptionIndex(md)
	return len(indexes) != 0
}
