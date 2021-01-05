// +build windows

package registry

import (
	"errors"
	"os"
	"strconv"

	"io/ioutil"

	"github.com/forensicanalysis/fslib"
	"github.com/forensicanalysis/fslib/filesystem/systemfs"
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"
	"golang.org/x/sys/windows/registry"
	"www.velocidex.com/golang/regparser"

	"regexp"
	"strings"
)

func init() { scanner.RegisterSystemScanner(&systemScanner{}) }

type systemScanner struct {
	iocs []eventIOC
}

type eventIOC struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	//type:
	// 0 == key exist
	// 1 == name exist
	// 2 == name contains exist
	// 3 == key value Contains
	// 4 == key value regex match
	// 5 == key value Contains (without name)
	// 6 == key value regex match (without name)
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

func keyCheck(key string, name string, valuex string, typex int) bool {
	var baseHandle registry.Key = 0xbad
	var hkcu bool = false
	for prefix, handle := range map[string]registry.Key{
		"HKEY_CLASSES_ROOT":     registry.CLASSES_ROOT,
		"HKEY_CURRENT_USER":     registry.CURRENT_USER,
		"HKCU":                  registry.CURRENT_USER,
		"HKEY_LOCAL_MACHINE":    registry.LOCAL_MACHINE,
		"HKLM":                  registry.LOCAL_MACHINE,
		"HKEY_USERS":            registry.USERS,
		"HKU":                   registry.USERS,
		"HKEY_PERFORMANCE_DATA": registry.PERFORMANCE_DATA,
		"HKEY_CURRENT_CONFIG":   registry.CURRENT_CONFIG,
	} {
		if strings.HasPrefix(key, prefix+`\`) {
			if strings.Contains(prefix, "HKEY_CURRENT_USER") || strings.Contains(prefix, "HKCU") {
				hkcu = true
			}
			baseHandle = handle
			key = key[len(prefix)+1:]
			break
		}
	}
	// if CURRENT_USER
	if hkcu {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\ProfileList", registry.QUERY_VALUE)
		val, err := getRegistryValueAsString(k, "ProfilesDirectory")
		if err != nil {
			log.Noticef("Error to open ProfilesDirectory : %s", err)
			return false
		}
		m1 := regexp.MustCompile(`%([^\%]+)%`)
		val = m1.ReplaceAllString(val, "$${$1}")
		val = os.ExpandEnv(val)
		files, err := ioutil.ReadDir(val)
		if err != nil {
			log.Noticef("Error open user profils directory : %s", err)
		}
		for _, f := range files {
			log.Noticef("Debug Open registre hive: %s", val+"\\"+f.Name()+"\\NTUSER.dat")
			if _, err := os.Stat(val + "\\" + f.Name() + "\\NTUSER.dat"); err == nil {
				log.Noticef("Open registre hive: %s", val+"\\"+f.Name()+"\\NTUSER.dat")
				//fr, err := os.OpenFile(val+"\\"+f.Name()+"\\NTUSER.dat", os.O_RDONLY, 0600)
				//fr, err := os.OpenFile(val+"\\"+f.Name()+"\\NTUSER.dat", syscall.O_RDONLY|syscall.FILE_SHARE_READ, 0444)
				systemFS, _ := fslib.FS.(*systemfs.FS)
				fr, _, err := systemFS.NTFSOpen(val + "\\" + f.Name() + "\\NTUSER.dat")
				//fr := bytes.NewReader(frx)
				if err != nil {
					log.Noticef("Error open base NTUSER: %s -- %s", val+"\\"+f.Name()+"\\NTUSER.dat", err)
					continue
				}
				uregistry, err := regparser.NewRegistry(fr)
				if err != nil {
					log.Noticef("Error load base NTUSER: %s -- %s", val+"\\"+f.Name()+"\\NTUSER.dat", err)
					continue
				}
				xkeys := uregistry.OpenKey(key)
				if xkeys == nil {
					log.Noticef("Can't open registry key: %s in %s", key, val+"\\"+f.Name()+"\\NTUSER.dat")
					continue
				}
				for _, vals := range xkeys.Values() {
					log.Noticef("Registre val %s : %#v\n", vals.ValueName(), vals.ValueData())
				}
			}
		}
		return false
	}
	log.Debugf("Looking for %s %s ...", key, name)
	if baseHandle == 0xbad {
		log.Debugf("Unknown registry key prefix: %s", key)
		return false
	}
	//
	var err error
	k, err := registry.OpenKey(baseHandle, key, registry.QUERY_VALUE)
	if err != nil {
		log.Debugf("Can't open registry key : %s", key)
		return false
	}
	defer k.Close()
	if typex == 0 {
		//key name exist
		return true
	}
	switch typex {
	case
		2,
		5,
		6:
		params, err := k.ReadValueNames(0)
		if err != nil {
			log.Debugf("Can't ReadSubKeyNames : %s %#v", key, err)
			return false
		}
		for _, param := range params {
			if typex == 2 {
				res := strings.Contains(param, name)
				if res {
					return true
				}
			}
			if typex == 5 {
				val, err := getRegistryValueAsString(k, param)
				if err != nil {
					log.Debugf("Error : %s", err)
					continue
				}
				res := strings.Contains(val, valuex)
				if res {
					return true
				}
			}
			if typex == 6 {
				val, err := getRegistryValueAsString(k, param)
				if err != nil {
					log.Debugf("Error : %s", err)
					continue
				}
				matched, err := regexp.MatchString(valuex, val)
				if err != nil {
					log.Debugf("Error regexp : %s", err)
					return false
				}
				if matched {
					return true
				}
			}
		}
		return false
	}
	val, err := getRegistryValueAsString(k, name)
	if err != nil {
		log.Debugf("Error : %s", err)
		return false
	}
	if typex == 1 {
		//key name exist
		return true
	}
	if typex == 3 {
		//value Contains
		res := strings.Contains(val, valuex)
		if res {
			return true
		}
		return false
	}
	if typex == 4 {
		matched, err := regexp.MatchString(valuex, val)
		if err != nil {
			log.Debugf("Error regexp : %s", err)
			return false
		}
		if matched {
			return true
		}
		return false
	}
	// settings[param] = val
	// test val according by type
	return false
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

func (s *systemScanner) Scan() error {
	for _, ioc := range s.iocs {
		if keyCheck(ioc.Key, ioc.Name, ioc.Value, ioc.Type) {
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
