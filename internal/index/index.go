package index

import (
	"regexp"
	"strings"

	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var indexRegexp *regexp.Regexp

func init() {
	// Single-column index:
	//  - ID
	//  - ID@Item
	//  - ID<Key>@Item
	//  - ID<Key, Key2>@Item
	//
	// Multi-column index (composite index):
	//  - (ID, Name)
	//  - (ID, Name)@Item
	//  - (ID, Name)<Key>@Item
	//  - (ID, Name)<Key1, Key2>@Item
	indexRegexp = regexp.MustCompile(`^(?P<cols>\([^)]+\)|[^<@]+)?(<(?P<keys>[^>]+)>)?(@(?P<name>.+))?$`)
}

type Index struct {
	Cols []string // column names in CamelCase (single-column or multi-column)
	Name string   // index name in CamelCase
	Keys []string // key names in CamelCase (single-column or multi-column)
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
	if len(index.Keys) != 0 {
		syntax += "<" + strings.Join(index.Keys, ",") + ">"
	}
	if index.Name != "" {
		syntax += "@" + index.Name
	}
	return syntax
}

// parse worksheet option index
func ParseWSOptionIndex(md protoreflect.MessageDescriptor) []*Index {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	return parseIndexFrom(wsOpts.Index)
}

func parseIndex(indexStr string) *Index {
	index := &Index{}
	matches := indexRegexp.FindStringSubmatch(indexStr)
	// Extract columns
	if cols := matches[indexRegexp.SubexpIndex("cols")]; cols != "" {
		if strings.HasPrefix(cols, "(") && strings.HasSuffix(cols, ")") {
			// Multi-column index
			cols = cols[1 : len(cols)-1]
			splitCols := strings.Split(cols, ",")
			if len(splitCols) <= 1 {
				return nil
			}
			for _, col := range splitCols {
				col = strings.TrimSpace(col)
				if col != "" {
					index.Cols = append(index.Cols, col)
				}
			}
		} else {
			// Single-column index
			if len(strings.Split(cols, ",")) > 1 {
				return nil
			}
			col := strings.TrimSpace(cols)
			if col != "" {
				index.Cols = append(index.Cols, col)
			}
		}
	}
	if len(index.Cols) == 0 {
		return nil
	}
	// Extract keys
	if keys := matches[indexRegexp.SubexpIndex("keys")]; keys != "" {
		index.Keys = strings.Split(keys, ",")
		for i, key := range index.Keys {
			index.Keys[i] = strings.TrimSpace(key)
		}
	}
	// Extract name
	if name := matches[indexRegexp.SubexpIndex("name")]; name != "" {
		index.Name = name
	}
	return index
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
