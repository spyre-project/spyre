// +build linux darwin dragonfly freebsd netbsd openbsd solaris

package platform

import (
	"C"
	"syscall"
	"unsafe"
)

func GetSystemInformation() SystemInformation {
	var utsname syscall.Utsname
	if err := syscall.Uname(&utsname); err != nil {
		return nil
	}
	return SystemInformation{
		{"sysname", C.GoString((*C.char)(unsafe.Pointer(&utsname.Sysname[0])))},
		{"nodename", C.GoString((*C.char)(unsafe.Pointer(&utsname.Nodename[0])))},
		{"release", C.GoString((*C.char)(unsafe.Pointer(&utsname.Release[0])))},
		{"version", C.GoString((*C.char)(unsafe.Pointer(&utsname.Version[0])))},
		{"machine", C.GoString((*C.char)(unsafe.Pointer(&utsname.Machine[0])))},
		{"domainname", C.GoString((*C.char)(unsafe.Pointer(&utsname.Domainname[0])))},
	}

	/*
	   $ lsb_release --all 2>/dev/null
	   Distributor ID:	Debian
	   Description:	Debian GNU/Linux 11 (bullseye)
	   Release:	11
	   Codename:	bullseye
	*/

}
