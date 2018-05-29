package config

import (
	"github.com/dcso/spyre/sys"
)

var defaultPaths []string

func init() {
	drives, _ := sys.GetLogicalDriveStrings()
	for _, d := range drives {
		if t, _ := sys.GetDriveType(d); t == sys.DRIVE_FIXED {
			defaultPaths = append(defaultPaths, d)
		}
	}
}
