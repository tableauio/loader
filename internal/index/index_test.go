package index

import (
	"reflect"
	"testing"
)

func Test_parseIndex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *Index
	}{
		{
			name:  "Single column with single sorted column and name",
			input: "Column1<SortedCol1>@IndexName",
			want: &Index{
				Cols: []string{"Column1"},
				SortedCols: []string{"SortedCol1"},
				Name: "IndexName",
			},
		},
		{
			name:  "Multi-column with multi sorted column and name",
			input: "( Column1 , Column2 )< SortedCol1 , SortedCol2 >@IndexName",
			want: &Index{
				Cols: []string{"Column1", "Column2"},
				SortedCols: []string{"SortedCol1", "SortedCol2"},
				Name: "IndexName",
			},
		},
		{
			name:  "Single column with name only",
			input: "Column3@IndexName",
			want: &Index{
				Cols: []string{"Column3"},
				Name: "IndexName",
			},
		},
		{
			name:  "Multi-column without sorted columns or name",
			input: "(Column4, Column5)",
			want: &Index{
				Cols: []string{"Column4", "Column5"},
			},
		},
		{
			name:  "Single column with single sorted column only",
			input: "Column6<SortedCol>",
			want: &Index{
				Cols: []string{"Column6"},
				SortedCols: []string{"SortedCol"},
			},
		},
		{
			name:  "zinotest",
			input: "ActivityID<Goal,ID>",
			want: &Index{
				Cols: []string{"ActivityID"},
				SortedCols: []string{"Goal", "ID"},
			},
		},
		{
			name:  "Multi-column with spaces around commas",
			input: "(Column7,  Column8,  Column9)<SortedCol7,  SortedCol8,  SortedCol9>@IndexName",
			want: &Index{
				Cols: []string{"Column7", "Column8", "Column9"},
				SortedCols: []string{"SortedCol7", "SortedCol8", "SortedCol9"},
				Name: "IndexName",
			},
		},
		{
			name:  "Invalid format (multi-column without parentheses)",
			input: "Column10, Column11<SortedCol10, SortedCol11>@IndexName",
			want:  nil,
		},
		{
			name:  "Invalid format (single column with parentheses)",
			input: "(Column12)<SortedCol12>@IndexName",
			want:  nil,
		},
		{
			name:  "Invalid format (empty columns)",
			input: "<SortedCol13>@IndexName",
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseIndex(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
