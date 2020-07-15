// +build darwin dragonfly freebsd netbsd openbsd solaris

package main

import (
	"github.com/spyre-project/spyre/log"

	"syscall"
)

func setLowPriority() {
	if err := syscall.Setpriority(syscall.PRIO_PROCESS, syscall.Getpid(), 10); err != nil {
		log.Errorf("Failed to set priority: %v", err)
	}
	return
}
