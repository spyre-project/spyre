// +build linux darwin freebsd netbsd openbsd solaris aix

package config

var defaultPaths = []string{"/"}

func getdrive() []string {
	return defaultPaths
}
