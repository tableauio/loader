package loadutil

import (
	"fmt"
	"testing"
)

func TestOrdinal(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0th"},
		{1, "1st"},
		{2, "2nd"},
		{3, "3rd"},
		{4, "4th"},
		{10, "10th"},
		{11, "11th"},
		{12, "12th"},
		{13, "13th"},
		{21, "21st"},
		{22, "22nd"},
		{23, "23rd"},
		{101, "101st"},
		{111, "111th"},
		{-1, "-1st"},
		{-12, "-12th"},
		{1001, "1001st"},
		{1112, "1112th"}, // 12 -> th
		{1123, "1123rd"},
		{214, "214th"}, // 14 -> th
	}

	for _, tc := range tests {
		tc := tc // capture
		t.Run(fmt.Sprintf("%d", tc.in), func(t *testing.T) {
			t.Parallel()
			got := Ordinal(tc.in)
			if got != tc.want {
				t.Fatalf("ordinal(%d) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}
