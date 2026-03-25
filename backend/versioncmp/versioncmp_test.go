package versioncmp

import "testing"

func TestCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "equal with v prefix", a: "v1.2.3", b: "1.2.3", want: 0},
		{name: "numeric compare", a: "1.2.10", b: "1.2.9", want: 1},
		{name: "stable newer than suffixed", a: "1.2.3", b: "1.2.3b", want: 1},
		{name: "suffix lexical compare", a: "1.2.3a", b: "1.2.3b", want: -1},
		{name: "implicit zero segments", a: "1.2", b: "1.2.0", want: 0},
		{name: "empty smaller than non-empty", a: "", b: "0.0.1", want: -1},
		{name: "uppercase V prefix", a: "V2.0", b: "v1.9.9", want: 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Compare(tt.a, tt.b)
			if got != tt.want {
				t.Fatalf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
