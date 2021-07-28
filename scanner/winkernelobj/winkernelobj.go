// +build windows

package winkernelobj

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"github.com/hillu/go-ntdll"

	"errors"
	"strings"
	"unsafe"
)

func init() { scanner.RegisterSystemScanner(&systemScanner{}) }

type systemScanner struct {
	// description -> objectname
	IOCs map[string]Obj `yaml:"iocs"`
}

type Obj struct {
	Type   string `yaml:"type"`
	String string `yaml:"string"`
}

func (s *systemScanner) FriendlyName() string { return "Windows-Kernel-Object" }
func (s *systemScanner) ShortName() string    { return "winkernelobj" }

func (s *systemScanner) Init(c *config.ScannerConfig) error {
	if err := c.Config.Decode(s); err != nil {
		return err
	}
	log.Debugf("%s: Initialized %d rules", s.ShortName(), len(s.IOCs))
	return nil
}

type walkFunc func(string, string) error

var skipDir = errors.New("skip this directory")

func walk(entry string, fn walkFunc) error {
	var h ntdll.Handle
	if st := ntdll.NtOpenDirectoryObject(&h, ntdll.STANDARD_RIGHTS_READ|ntdll.DIRECTORY_QUERY,
		ntdll.NewObjectAttributes(entry, 0, 0, nil),
	); st != 0 {
		return st.Error()
	}
	defer ntdll.NtClose(h)
	var context uint32
	for {
		var buf [32768]byte
		var length uint32
		switch st := ntdll.NtQueryDirectoryObject(
			h,
			&buf[0],
			uint32(len(buf)),
			true,
			context == 0,
			&context,
			&length,
		); st {
		case ntdll.STATUS_SUCCESS:
		case ntdll.STATUS_NO_MORE_ENTRIES:
			return nil
		default:
			return st.Error()
		}
		odi := (*ntdll.ObjectDirectoryInformationT)(unsafe.Pointer(&buf[0]))
		var path string
		if entry == `\` {
			path = `\` + odi.Name.String()
		} else {
			path = entry + `\` + odi.Name.String()
		}
		switch typ := odi.TypeName.String(); typ {
		case "Directory":
			if err := walk(path, fn); err != nil {
				return err
			}
		default:
			switch err := fn(path, typ); err {
			case skipDir:
				return nil
			case nil:
				continue
			default:
				return err
			}
		}
	}
}

func (s *systemScanner) Scan() error {
	walk(`\`, func(path, typ string) error {
		for description, ioc := range s.IOCs {
			if typ == ioc.Type && strings.Contains(path, ioc.String) {
				report.AddStringf("%s: Fonud %s:%s - indicator for %s", s.ShortName(), typ, path, description)
			}
		}
		return nil
	})
	return nil
}
