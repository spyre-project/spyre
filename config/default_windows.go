package config

import (
	"github.com/spyre-project/spyre/platform/sys"
)

func defaultPaths() (paths []string) {
	drives, _ := sys.GetLogicalDriveStrings()
	for _, d := range drives {
		if t, _ := sys.GetDriveType(d); t == sys.DRIVE_FIXED {
			paths = append(paths, d)
		}
	}
	return
}
