//go:build windows
// +build windows

package findwindow

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/platform/sys"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"fmt"
	"syscall"
)

func init() { scanner.RegisterSystemScanner(&systemScanner{}) }

type systemScanner struct {
	// description -> objectname
	IOCs map[string]obj `yaml:"iocs"`
}

type obj struct {
	Class string `yaml:"class"`
	Name  string `yaml:"name"`
}

func (s *systemScanner) FriendlyName() string { return "Find-Window" }
func (s *systemScanner) ShortName() string    { return "findwindow" }

func (s *systemScanner) Init(c *config.ScannerConfig) error {
	if err := c.Config.Decode(s); err != nil {
		return err
	}
	log.Debugf("%s: Read %d IOCs", s.ShortName(), len(s.IOCs))
	return nil
}

func (s *systemScanner) Scan() error {
	for description, ioc := range s.IOCs {
		var name, class *uint8
		if ioc.Name != "" {
			buf := []byte(ioc.Name)
			buf = append(buf, 0)
			name = &(buf[0])
		}
		if ioc.Class != "" {
			buf := []byte(ioc.Class)
			buf = append(buf, 0)
			class = &(buf[0])
		}
		if h, _ := sys.FindWindow(class, name); h != 0 {
			report.AddSystemInfo(s.ShortName(), fmt.Sprintf("Found window for %s", description))
			syscall.CloseHandle(h)
		}
	}
	return nil
}
