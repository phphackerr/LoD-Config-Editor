package map_downloader

import "testing"

func TestMapFileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "lod_latest", want: "lod_latest.w3x"},
		{in: "lod_latest.w3m", want: "lod_latest.w3m"},
		{in: "../evil", want: "_evil.w3x"},
		{in: "CON.w3x", want: "_CON.w3x"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got := mapFileName(tt.in)
			if got != tt.want {
				t.Fatalf("mapFileName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
