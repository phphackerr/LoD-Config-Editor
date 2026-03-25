package map_downloader

import (
	"strings"
	"testing"
)

func TestParseMapInfos(t *testing.T) {
	t.Parallel()

	html := `
<table>
  <tr>
    <td>1</td>
    <td>
      <a href="/maps/12345/">LoD v6.89b</a>
      <a href="/maps/12345/download/?token=abc">Download</a>
      (145.54 MB)
    </td>
    <td>AoS</td>
    <td>19 Mar 2026</td>
  </tr>
  <tr>
    <td>2</td>
    <td>
      <a href="/maps/12345/">LoD v6.89b</a>
      <a href="/maps/12345/download/?token=abc">Duplicate Download</a>
      (145.54 MB)
    </td>
    <td>AoS</td>
    <td>19 Mar 2026</td>
  </tr>
</table>`

	md := NewMapDownloader(nil)
	infos, err := md.parseMapInfos(html)
	if err != nil {
		t.Fatalf("parseMapInfos returned error: %v", err)
	}

	if len(infos) != 1 {
		t.Fatalf("expected 1 unique result, got %d (%v)", len(infos), infos)
	}

	got := infos[0]
	if got.Name != "LoD v6.89b" {
		t.Fatalf("unexpected name: %q", got.Name)
	}
	if got.Version != "v6.89b" {
		t.Fatalf("unexpected version: %q", got.Version)
	}
	if !strings.HasPrefix(got.DownloadLink, epicwarBaseURL+"/maps/12345/download/") {
		t.Fatalf("unexpected normalized download link: %q", got.DownloadLink)
	}
	if got.Date != "19 Mar 2026" {
		t.Fatalf("unexpected date: %q", got.Date)
	}
	if got.Size <= 145*1024*1024 {
		t.Fatalf("unexpected size in bytes: %d", got.Size)
	}
}

func TestParseMapInfosModernMarkup(t *testing.T) {
	t.Parallel()

	html := `
<div class="text-xs grid gap-1 grid-cols-[128px,1fr] md:grid-cols-[128px,3fr,1fr,1fr,1fr,1fr]">
  <div class="p-1">
    <a href="/maps/347687/"><img src="/img/thumb.png" alt=""></a>
    <a href="/maps/347687/">DotA v6.93e LoD</a>
    <div class="mt-2">
      <a href="https://www.epicwar.com/maps/347687/download/?token=abc&time=1">lod_fixed.w3x</a>
      (145.54 MB)
    </div>
  </div>
  <div>Category: Tower Defense</div>
  <div>Date: 7 Nov 2025</div>
</div>
<div class="text-xs grid gap-1 grid-cols-[128px,1fr] md:grid-cols-[128px,3fr,1fr,1fr,1fr,1fr]">
  <div class="p-1">
    <a href="/maps/347687/">DotA v6.93e LoD</a>
    <div class="mt-2">
      <a href="https://www.epicwar.com/maps/347687/download/?token=def&time=2">lod_fixed.w3x</a>
      (145.54 MB)
    </div>
  </div>
  <div>Date: 7 Nov 2025</div>
</div>`

	md := NewMapDownloader(nil)
	infos, err := md.parseMapInfos(html)
	if err != nil {
		t.Fatalf("parseMapInfos returned error: %v", err)
	}

	if len(infos) != 1 {
		t.Fatalf("expected 1 unique result by map id, got %d (%v)", len(infos), infos)
	}

	got := infos[0]
	if got.Name != "DotA v6.93e LoD" {
		t.Fatalf("unexpected name: %q", got.Name)
	}
	if got.Version != "v6.93e" {
		t.Fatalf("unexpected version: %q", got.Version)
	}
	if !strings.Contains(got.DownloadLink, "/maps/347687/download/") {
		t.Fatalf("unexpected download link: %q", got.DownloadLink)
	}
	if got.Date != "7 Nov 2025" {
		t.Fatalf("unexpected date: %q", got.Date)
	}
	if got.Size <= 145*1024*1024 {
		t.Fatalf("unexpected size in bytes: %d", got.Size)
	}
}

func TestNormalizeDownloadURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "relative", in: "/maps/123/download/?token=x"},
		{name: "absolute epicwar", in: "https://www.epicwar.com/maps/123/download/?token=x"},
		{name: "invalid host", in: "https://example.com/maps/123/download/", wantErr: true},
		{name: "invalid scheme", in: "file:///tmp/a", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeDownloadURL(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got url=%q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasPrefix(got, "http") {
				t.Fatalf("expected normalized absolute http(s) url, got %q", got)
			}
		})
	}
}
