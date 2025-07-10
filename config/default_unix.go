//go:build linux || darwin || freebsd || netbsd || openbsd || solaris || aix
// +build linux darwin freebsd netbsd openbsd solaris aix

package config

func defaultPaths() []string { return []string{"/"} }
