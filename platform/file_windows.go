//go:build windows
// +build windows

package platform

import (
	"os"
	"strings"
	"unsafe"

	"github.com/hillu/go-ntdll"
)

// Returns main and alternate data stream paths (if any)
func GetPaths(path string) (paths []string) {
	paths = []string{path}
	var info [32768]byte
	var iostatus ntdll.IoStatusBlock

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	nts := ntdll.NtQueryInformationFile(
		ntdll.Handle(f.Fd()),
		&iostatus,
		&info[0],
		uint32(len(info)),
		ntdll.FileStreamInformation)
	if nts != ntdll.STATUS_SUCCESS {
		return
	}

	i := 0
	for {
		fi := (*ntdll.FileStreamInformationT)(unsafe.Pointer(&info[i]))
		name := ntdll.NewUnicodeStringFromBuffer(&fi.StreamName[0], int(fi.StreamNameLength)).String()
		if name != "" && name != "::$DATA" {
			if strings.HasSuffix(name, ":$DATA") {
				name = name[:len(name)-6]
			}
			paths = append(paths, path+name)
		}
		if i = int(fi.NextEntryOffset); i == 0 {
			break
		}
	}

	return
}
