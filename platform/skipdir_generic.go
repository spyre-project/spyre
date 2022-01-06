// +build !linux,!darwin

package platform

func SkipDir(path string) bool { return false }
