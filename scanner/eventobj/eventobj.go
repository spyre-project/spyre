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
	iocs []eventIOC
}

type eventIOC struct {
	Event       string `json:"event"`
	Description string `json:description`
}

type iocFile struct {
	EventObjects []eventIOC `json:"event-objects"`
}

func (s *systemScanner) Name() string { return "Event-Object" }

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
		for _, ioc := range current.EventObjects {
			s.iocs = append(s.iocs, ioc)
		}
	}
	return nil
}

func (s *systemScanner) Scan() error {
	for _, ioc := range s.iocs {
		u16, err := windows.UTF16PtrFromString(ioc.Event)
		if err != nil {
			log.Noticef("invalid event path: %s", err)
			continue
		}
		h, err := windows.OpenEvent(0x00100000, false, u16)
		if err != nil {
			continue
		}
		windows.CloseHandle(h)
		report.AddStringf("Found event %s: Indicator for %s", ioc.Event, ioc.Description)
	}
	return nil
}
