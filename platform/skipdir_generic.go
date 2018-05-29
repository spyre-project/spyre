// +build !linux

package platform

import (
	"github.com/spf13/afero"
)

func SkipDir(fs afero.Fs, path string) bool {
	return false
}
