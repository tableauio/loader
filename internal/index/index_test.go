package index

import (
	"reflect"
	"testing"
)

func Test_parseColsFrom(t *testing.T) {
	type args struct {
		multiColGroup string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{
			name: "single-column",
			args: args{
				multiColGroup: "ID",
			},
			want: []string{"ID"},
		},
		{
			name: "multi-column without with parentheses",
			args: args{
				multiColGroup: "ID, Type",
			},
			want: []string{"ID", "Type"},
		},
		{
			name: "multi-column with parentheses",
			args: args{
				multiColGroup: "(ID,Type)",
			},
			want: []string{"ID", "Type"},
		},
		{
			name: "multi-column with space",
			args: args{
				multiColGroup: "(ID, Type)",
			},
			want: []string{"ID", "Type"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseColsFrom(tt.args.multiColGroup); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseColsFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
