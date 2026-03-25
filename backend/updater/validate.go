package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func verifyDownloadedExecutable(path string, minSize int64) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() < minSize {
		return "", fmt.Errorf("executable is unexpectedly small: %d bytes", info.Size())
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	header := make([]byte, 2)
	if _, err := io.ReadFull(file, header); err != nil {
		return "", err
	}
	if header[0] != 'M' || header[1] != 'Z' {
		return "", fmt.Errorf("file is not a valid PE executable (missing MZ header)")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
