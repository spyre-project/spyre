// +build darwin dragonfly freebsd netbsd openbsd solaris

package platform

import (
	"github.com/spyre-project/spyre/log"

	"syscall"
)

func SetLowPriority() {
	if err := syscall.Setpriority(syscall.PRIO_PROCESS, syscall.Getpid(), 10); err != nil {
		log.Errorf("Failed to set priority: %v", err)
	}
	return
}
