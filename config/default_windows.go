package config

import (
	"github.com/spyre-project/spyre/platform/sys"
)

var defaultPaths []string
var defaultEvtxPaths = []string{os.Getenv("SYSTEMROOT") + "\\system32\\winevt\\Logs\\"}

func init() {
	drives, _ := sys.GetLogicalDriveStrings()
	for _, d := range drives {
		if t, _ := sys.GetDriveType(d); t == sys.DRIVE_FIXED {
			defaultPaths = append(defaultPaths, d)
		}
	}
}
