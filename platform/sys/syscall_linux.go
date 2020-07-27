// +build linux

package sys

import (
	"syscall"
)

const (
	IOPRIO_WHO_PROCESS = 1
	IOPRIO_WHO_PGRP    = 2
	IOPRIO_WHO_USER    = 3
)

const (
	IOPRIO_CLASS_RT   = 1
	IOPRIO_CLASS_BE   = 2
	IOPRIO_CLASS_IDLE = 3
)

const ioprioClassShift = 13

// Given a scheduling class and priority (data), this macro combines
// the two values to produce an ioprio value, which is returned as the
// result of the macro.
func IoprioPrioValue(class, data int) int {
	return class<<ioprioClassShift | data
}

func IoPrioSet(which, who, ioprio int) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOPRIO_SET,
		uintptr(which), uintptr(who), uintptr(ioprio),
	)
	if e1 != 0 {
		err = syscall.Errno(e1)
	}
	return
}
