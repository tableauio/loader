package index

import (
	"reflect"
	"testing"
)

func Test_parseColsFrom(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *Index
	}{
		{
			name:  "Single column with single key and name",
			input: "Column1<Key1>@IndexName",
			want: &Index{
				Cols: []string{"Column1"},
				Keys: []string{"Key1"},
				Name: "IndexName",
			},
		},
		{
			name:  "Multi-column with multi-key and name",
			input: "( Column1 , Column2 )< Key1 , Key2 >@IndexName",
			want: &Index{
				Cols: []string{"Column1", "Column2"},
				Keys: []string{"Key1", "Key2"},
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
			name:  "Multi-column without keys or name",
			input: "(Column4, Column5)",
			want: &Index{
				Cols: []string{"Column4", "Column5"},
			},
		},
		{
			name:  "Single column with single key only",
			input: "Column6<Key6>",
			want: &Index{
				Cols: []string{"Column6"},
				Keys: []string{"Key6"},
			},
		},
		{
			name:  "Multi-column with spaces around commas",
			input: "(Column7,  Column8,  Column9)<Key7,  Key8,  Key9>@IndexName",
			want: &Index{
				Cols: []string{"Column7", "Column8", "Column9"},
				Keys: []string{"Key7", "Key8", "Key9"},
				Name: "IndexName",
			},
		},
		{
			name:  "Multi-column without parentheses",
			input: "Column10, Column11<Key10, Key11>@IndexName",
			want: &Index{
				Cols: []string{"Column10, Column11"},
				Keys: []string{"Key10", "Key11"},
				Name: "IndexName",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseIndex(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseColsFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
