package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"regexp"

	"github.com/spyre-project/spyre/platform/sys"
	"golang.org/x/sys/windows/registry"
)

var defaultPaths []string
var defaultEvtxPaths = []string{os.Getenv("SYSTEMROOT") + "\\system32\\winevt\\Logs\\"}

func init() {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\ProfileList", registry.QUERY_VALUE)
	val, err := getRegistryValueAsString(k, "ProfilesDirectory")
	if err == nil {
		m1 := regexp.MustCompile(`%([^\%]+)%`)
		val = m1.ReplaceAllString(val, "$${$1}")
		val = os.ExpandEnv(val)
		if stringInSlice(val, defaultPaths) {
			defaultPaths = append(defaultPaths, val)
		}
	}
	val = os.ExpandEnv("${windir}")
	if stringInSlice(val, defaultPaths) {
		defaultPaths = append(defaultPaths, val)
	}
	val = os.ExpandEnv("${SystemRoot}")
	if stringInSlice(val, defaultPaths) {
		defaultPaths = append(defaultPaths, val)
	}
	val = os.ExpandEnv("${ProgramFiles}")
	if stringInSlice(val, defaultPaths) {
		defaultPaths = append(defaultPaths, val)
	}
	val = os.ExpandEnv("${ProgramFiles(x86)}")
	if stringInSlice(val, defaultPaths) {
		defaultPaths = append(defaultPaths, val)
	}
	val = os.ExpandEnv("${ProgramData}")
	if stringInSlice(val, defaultPaths) {
		defaultPaths = append(defaultPaths, val)
	}
	val = os.ExpandEnv("${ALLUSERSPROFILE}")
	if stringInSlice(val, defaultPaths) {
		defaultPaths = append(defaultPaths, val)
	}
}

func getdrive() []string {
	defaultPaths = nil
	drives, _ := sys.GetLogicalDriveStrings()
	for _, d := range drives {
		if t, _ := sys.GetDriveType(d); t == sys.DRIVE_FIXED {
			defaultPaths = append(defaultPaths, d)
		}
	}
	return defaultPaths
}
func stringInSlice(a string, list []string) bool {
	if a == "" {
		return false
	}
	for _, b := range list {
		if b == a {
			return false
		}
	}
	return true
}

func getRegistryValueAsString(key registry.Key, subKey string) (string, error) {
	valString, _, err := key.GetStringValue(subKey)
	if err == nil {
		return valString, nil
	}
	valStrings, _, err := key.GetStringsValue(subKey)
	if err == nil {
		return strings.Join(valStrings, "\n"), nil
	}
	valBinary, _, err := key.GetBinaryValue(subKey)
	if err == nil {
		return string(valBinary), nil
	}
	valInteger, _, err := key.GetIntegerValue(subKey)
	if err == nil {
		return strconv.FormatUint(valInteger, 10), nil
	}
	return "", errors.New("Can't get type for sub key " + subKey)
}
