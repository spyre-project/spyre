// +build !windows

package platform

func GetPaths(path string) (paths []string) { return []string{path} }
