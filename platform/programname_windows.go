// +build windows

package platform

import (
	"os"

	"golang.org/x/sys/windows"
)

func GetProgramFilename() string {
	var filename [2048]uint16
	if l, err := windows.GetModuleFileName(0, &filename[0], uint32(len(filename))); err != nil {
		return os.Args[0]
	} else {
		return windows.UTF16ToString(filename[:int(l)])
	}
}
