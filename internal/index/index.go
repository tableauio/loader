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
	//  - ID<SortedCol>@Item
	//  - ID<SortedCol1, SortedCol2>@Item
	//
	// Multi-column index (composite index):
	//  - (ID, Name)
	//  - (ID, Name)@Item
	//  - (ID, Name)<SortedCol>@Item
	//  - (ID, Name)<SortedCol1, SortedCol2>@Item
	indexRegexp = regexp.MustCompile(`^(?P<Cols>\([^)]+\)|[^<@]+)?(<(?P<SortedCols>[^>]+)>)?(@(?P<Name>.+))?$`)
}

// matchIndex parses the index syntax and returns the columns, sorted columns and name.
func matchIndex(text string) (cols, sortedCols, name string) {
	match := indexRegexp.FindStringSubmatch(text)
	if match == nil {
		return "", "", ""
	}
	for i, expName := range indexRegexp.SubexpNames() {
		value := strings.TrimSpace(match[i])
		switch expName {
		case "Cols":
			cols = value
		case "SortedCols":
			sortedCols = value
		case "Name":
			name = value
		default:
			continue
		}
	}
	return cols, sortedCols, name
}

type Index struct {
	Cols       []string // column names in CamelCase (single-column or multi-column)
	Name       string   // index name in CamelCase
	SortedCols []string // sorted column names in CamelCase (single-column or multi-column)
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
	if len(index.SortedCols) != 0 {
		syntax += "<" + strings.Join(index.SortedCols, ",") + ">"
	}
	if index.Name != "" {
		syntax += "@" + index.Name
	}
	return syntax
}

// parse worksheet option index
func ParseWSOptionIndex(md protoreflect.MessageDescriptor) ([]*Index, []*Index) {
	opts := md.Options().(*descriptorpb.MessageOptions)
	wsOpts := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
	return parseIndexFrom(wsOpts.Index), parseIndexFrom(wsOpts.OrderedIndex)
}

func parseIndex(indexStr string) *Index {
	cols, sortedCols, name := matchIndex(indexStr)
	index := &Index{}
	// Extract columns
	if cols != "" {
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
	// Extract sortedCols
	if sortedCols != "" {
		index.SortedCols = strings.Split(sortedCols, ",")
		for i, col := range index.SortedCols {
			index.SortedCols[i] = strings.TrimSpace(col)
		}
	}
	// Extract name
	index.Name = name
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
