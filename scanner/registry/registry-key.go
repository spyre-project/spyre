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
	// description -> objectname
	IOCs map[string]obj `yaml:"iocs"`
}

type obj struct {
	Key  string `yaml:"key"`
	Name string `yaml:"name"`
}

func (s *systemScanner) FriendlyName() string { return "Registry-Key" }
func (s *systemScanner) ShortName() string    { return "registry" }

func (s *systemScanner) Init(c *config.ScannerConfig) error {
	if err := c.Config.Decode(s); err != nil {
		return err
	}
	log.Debugf("%s: Initialized %d rules", s.ShortName(), len(s.IOCs))
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
	for description, ioc := range s.IOCs {
		if keyExists(ioc.Key, ioc.Name) {
			var name string
			typ := "key"
			if ioc.Name != "" {
				name = " " + ioc.Name
				typ = "value"
			}
			report.AddStringf("registry: Found key %s [%s]%s -- IOC for %s", typ, ioc.Key, name, description)
		}
	}
	return nil
}
