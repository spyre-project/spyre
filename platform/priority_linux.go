//go:build linux
// +build linux

package platform

import (
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/platform/sys"

	"syscall"
)

func SetLowPriority() {
	pid := syscall.Getpid()
	if err := syscall.Setpriority(syscall.PRIO_PROCESS, pid, 10); err != nil {
		log.Errorf("Failed to set CPU priority: %v", err)
	}
	if err := sys.IoPrioSet(sys.IOPRIO_WHO_PROCESS, pid, sys.IoprioPrioValue(sys.IOPRIO_CLASS_IDLE, 0)); err != nil {
		log.Errorf("Failed to set I/O priority: %v", err)
	}
	return
}
