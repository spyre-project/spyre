// +build windows

package registry

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"golang.org/x/sys/windows"

	"strings"
)

func init() { scanner.RegisterSystemScanner(&systemScanner{}) }

type systemScanner struct {
	iocs []eventIOC
}

type eventIOC struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type iocFile struct {
	Keys []eventIOC `json:"registry-keys"`
}

func (s *systemScanner) Name() string { return "Registry-Key" }

func (s *systemScanner) Init() error {
	iocFiles := config.IocFiles
	if len(iocFiles) == 0 {
		iocFiles = []string{"ioc.json"}
	}
	for _, file := range iocFiles {
		var current iocFile
		if err := config.ReadIOCs(file, &current); err != nil {
			log.Error(err.Error())
		}
		for _, ioc := range current.Keys {
			s.iocs = append(s.iocs, ioc)
		}
	}
	return nil
}

func keyExists(key string, name string) bool {
	var baseHandle windows.Handle = 0xbad
	for prefix, handle := range map[string]windows.Handle{
		"HKEY_CLASSES_ROOT":     windows.HKEY_CLASSES_ROOT,
		"HKEY_CURRENT_USER":     windows.HKEY_CURRENT_USER,
		"HKCU":                  windows.HKEY_CURRENT_USER,
		"HKEY_LOCAL_MACHINE":    windows.HKEY_LOCAL_MACHINE,
		"HKLM":                  windows.HKEY_LOCAL_MACHINE,
		"HKEY_USERS":            windows.HKEY_USERS,
		"HKU":                   windows.HKEY_USERS,
		"HKEY_PERFORMANCE_DATA": windows.HKEY_PERFORMANCE_DATA,
		"HKEY_CURRENT_CONFIG":   windows.HKEY_CURRENT_CONFIG,
		"HKEY_DYN_DATA":         windows.HKEY_DYN_DATA,
	} {
		if strings.HasPrefix(key, prefix+`\`) {
			baseHandle = handle
			key = key[len(prefix)+1:]
			break
		}
	}
	log.Debugf("Looking for %s %s ...", key, name)
	if baseHandle == 0xbad {
		log.Debugf("Unknown registry key prefix: %s", key)
		return false
	}
	var u16 *uint16
	var err error
	if u16, err = windows.UTF16PtrFromString(key); err != nil {
		log.Debug("failed to convert key to utf16")
		return false
	}
	var h windows.Handle
	if err := windows.RegOpenKeyEx(baseHandle, u16, 0, windows.KEY_READ, &h); err != nil {
		return false
	}
	if name == "" {
		return true
	}
	defer windows.RegCloseKey(h)
	if u16, err = windows.UTF16PtrFromString(name); err != nil {
		log.Debug("failed to convert value name to utf16")
		return false
	}
	if err := windows.RegQueryValueEx(h, u16, nil, nil, nil, nil); err != nil {
		return false
	}
	return true
}

func (s *systemScanner) Scan() error {
	for _, ioc := range s.iocs {
		if keyExists(ioc.Key, ioc.Name) {
			var name string
			typ := "key"
			if ioc.Name != "" {
				name = " " + ioc.Name
				typ = "value"
			}
			report.AddStringf("Found registry %s [%s]%s -- IOC for %s", typ, ioc.Key, name, ioc.Description)
		}
	}
	return nil
}
