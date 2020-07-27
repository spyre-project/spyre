// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package platform

import (
	"github.com/spyre-project/spyre/log"

	"runtime"
)

func SetLowPriority() {
	log.Error("priority setting is not supported on " + runtime.GOOS)
	return
}
