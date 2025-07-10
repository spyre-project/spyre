//go:build linux
// +build linux

package platform

import (
	"os"
)

func GetProgramFilename() string {
	if filename, err := os.Readlink("/proc/self/exe"); err != nil {
		return os.Args[0]
	} else {
		return filename
	}
}
