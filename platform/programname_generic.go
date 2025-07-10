//go:build !linux && !windows
// +build !linux,!windows

package platform

import (
	"os"
)

func GetProgramFilename() string {
	return os.Args[0]
}
