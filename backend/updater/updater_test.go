package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSelectExecutableAsset(t *testing.T) {
	t.Parallel()

	assets := []ReleaseAsset{
		{Name: "LCE_Setup.exe", BrowserDownloadURL: "https://github.com/x"},
		{Name: "LCE_portable.exe", BrowserDownloadURL: "https://github.com/y"},
	}

	got, err := selectExecutableAsset(assets)
	if err != nil {
		t.Fatalf("selectExecutableAsset error: %v", err)
	}
	if got.Name != "LCE_portable.exe" {
		t.Fatalf("unexpected selected asset: %s", got.Name)
	}
}

func TestParseChecksumFile(t *testing.T) {
	t.Parallel()

	content := `
# checksums
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  LCE_portable.exe
LCE_Setup.exe: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
`
	result := parseChecksumFile(content)
	if result["LCE_portable.exe"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("missing or invalid hash for portable asset")
	}
	if result["LCE_Setup.exe"] != "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatalf("missing or invalid hash for setup asset")
	}
}

func TestValidateRemoteURL(t *testing.T) {
	t.Parallel()

	if err := validateRemoteURL("https://api.github.com/repos/a/b/releases/latest"); err != nil {
		t.Fatalf("expected allowed URL, got error: %v", err)
	}
	if err := validateRemoteURL("http://api.github.com/repos/a/b/releases/latest"); err == nil {
		t.Fatalf("expected http URL to be rejected")
	}
	if err := validateRemoteURL("https://example.com/file.exe"); err == nil {
		t.Fatalf("expected foreign host to be rejected")
	}
}

func TestVerifyDownloadedExecutable(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "lce_new.exe")
	content := make([]byte, 700*1024)
	content[0] = 'M'
	content[1] = 'Z'
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test executable: %v", err)
	}

	sum, err := verifyDownloadedExecutable(path, 512*1024)
	if err != nil {
		t.Fatalf("verifyDownloadedExecutable returned error: %v", err)
	}
	if len(sum) != 64 {
		t.Fatalf("unexpected sha256 length: %d", len(sum))
	}
}
