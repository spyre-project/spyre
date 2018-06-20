// +build windows

package main

import (
	"github.com/dcso/spyre/log"
	"github.com/dcso/spyre/sys"

	"syscall"
)

func setLowPriority() {
	procHandle, err := syscall.GetCurrentProcess()
	if err != nil {
		log.Errorf("Failed to get handle to process: %v", err)
		return
	}
	if err = sys.SetPriorityClass(procHandle, sys.IDLE_PRIORITY_CLASS|sys.PROCESS_MODE_BACKGROUND_BEGIN); err != nil {
		log.Errorf("Failed to set priority class: %v", err)
	}
	return
}
