// +build windows

package platform

import (
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/platform/sys"

	"syscall"
)

func SetLowPriority() {
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
