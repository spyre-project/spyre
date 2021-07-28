package sys

import (
	"unicode/utf16"
)

//sys	getLogicalDriveStrings(bufferLength uint32, lpBuffer *uint16) (requiredLength uint32, err error) = GetLogicalDriveStringsW
//sys	GetDriveType(RootPathName string) (driveType uint32, err error) = GetDriveTypeW
//sys	FindWindow(className *uint8, windowName *uint8) (handle syscall.Handle, err error) = user32.FindWindowA
//sys	GetPriorityClass(process syscall.Handle) (priorityClass uint32, err error) = GetPriorityClass
//sys	SetPriorityClass(process syscall.Handle, priorityClass uint32) (err error) = SetPriorityClass

const (
	DRIVE_UNKNOWN     = 0
	DRIVE_NO_ROOT_DIR = 1
	DRIVE_REMOVABLE   = 2
	DRIVE_FIXED       = 3
	DRIVE_REMOTE      = 4
	DRIVE_CDROM       = 5
	DRIVE_RAMDISK     = 6
)

func GetLogicalDriveStrings() ([]string, error) {
	n, _ := getLogicalDriveStrings(0, nil)
	buf := make([]uint16, n+1)
	n, err := getLogicalDriveStrings(n, &buf[0])
	if err != nil {
		return nil, err
	}
	var rv []string
	var b int
	for i := 0; i < int(n); i++ {
		if buf[i] == 0 {
			if s := string(utf16.Decode(buf[b:i])); len(s) > 0 {
				rv = append(rv, s)
			}
			b = i + 1
			continue
		}
	}
	return rv, nil
}

const (
	ABOVE_NORMAL_PRIORITY_CLASS   = 0x00008000
	BELOW_NORMAL_PRIORITY_CLASS   = 0x00004000
	HIGH_PRIORITY_CLASS           = 0x00000080
	IDLE_PRIORITY_CLASS           = 0x00000040
	NORMAL_PRIORITY_CLASS         = 0x00000020
	PROCESS_MODE_BACKGROUND_BEGIN = 0x00100000
	PROCESS_MODE_BACKGROUND_END   = 0x00200000
	REALTIME_PRIORITY_CLASS       = 0x00000100
)
