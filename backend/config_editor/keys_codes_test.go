package config_editor

import "testing"

func TestKeyCodesLookupWheel(t *testing.T) {
	t.Parallel()

	if got := Lookup("0x0A"); got != "wheelUp" {
		t.Fatalf("Lookup(0x0A) = %q, want %q", got, "wheelUp")
	}
	if got := Lookup("0x0B"); got != "wheelDn" {
		t.Fatalf("Lookup(0x0B) = %q, want %q", got, "wheelDn")
	}
}

func TestKeyCodesReverseLookupCaseInsensitive(t *testing.T) {
	t.Parallel()

	if got := ReverseLookup("wheelUp"); got != "0x0A" {
		t.Fatalf("ReverseLookup(wheelUp) = %q, want %q", got, "0x0A")
	}
	if got := ReverseLookup("wheelup"); got != "0x0A" {
		t.Fatalf("ReverseLookup(wheelup) = %q, want %q", got, "0x0A")
	}
	if got := ReverseLookup("wheelDn"); got != "0x0B" {
		t.Fatalf("ReverseLookup(wheelDn) = %q, want %q", got, "0x0B")
	}
	if got := ReverseLookup("wheeldn"); got != "0x0B" {
		t.Fatalf("ReverseLookup(wheeldn) = %q, want %q", got, "0x0B")
	}
}
