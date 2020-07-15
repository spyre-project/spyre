// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package main

import (
	"github.com/spyre-project/spyre/log"
	"runtime"
)

func setLowPriority() {
	log.Error("priority setting is not supported on " + runtime.GOOS)
	return
}
