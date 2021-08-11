package report

import (
	"github.com/spf13/afero"

	"crypto/sha256"
	"io"
)

func hashFile(f afero.File) ([]byte, error) {
	hash := sha256.New()
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if _, err := io.Copy(hash, f); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
