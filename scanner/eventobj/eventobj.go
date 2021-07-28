// +build windows

package eventobj

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"golang.org/x/sys/windows"
)

func init() { scanner.RegisterSystemScanner(&systemScanner{}) }

type systemScanner struct {
	// description -> objectname
	IOCs map[string]string `yaml:"iocs"`
}

func (s *systemScanner) FriendlyName() string { return "Event-Object" }
func (s *systemScanner) ShortName() string    { return "eventobj" }

func (s *systemScanner) Init(c *config.ScannerConfig) error {
	return c.Config.Decode(s)
}

func (s *systemScanner) Scan() error {
	for description, objname := range s.IOCs {
		u16, err := windows.UTF16PtrFromString(objname)
		if err != nil {
			log.Noticef("invalid event path: %s", err)
			continue
		}
		h, err := windows.OpenEvent(0x00100000, false, u16)
		if err != nil {
			continue
		}
		windows.CloseHandle(h)
		report.AddStringf("%s: Found event %s: Indicator for %s", s.ConfigSection(), objname, description)
	}
	return nil
}
