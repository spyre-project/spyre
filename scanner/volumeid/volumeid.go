// +build windows

package volumeid

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/platform/sys"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"golang.org/x/sys/windows"
)

func init() { scanner.RegisterSystemScanner(&systemScanner{}) }

type systemScanner struct{}

func (s *systemScanner) FriendlyName() string { return "Diag-VolumeID" }
func (s *systemScanner) ShortName() string    { return "volumeid" }

func (s *systemScanner) Init(*config.ScannerConfig) error { return nil }

func (s *systemScanner) Scan() error {
	drives, _ := sys.GetLogicalDriveStrings()
	for _, d := range drives {
		if t, _ := sys.GetDriveType(d); t == sys.DRIVE_FIXED {
			var volNameU16 [1024]uint16
			var volSerial uint32
			if err := sys.GetVolumeInformation(
				d+`\`,
				&volNameU16[0], uint32(len(volNameU16)),
				&volSerial,
				nil, nil, nil, 0,
			); err != nil {
				log.Errorf("Could not determine volume information for %s", d)
				continue
			}
			volName := windows.UTF16ToString(volNameU16[:])
			report.AddStringf("%s: %s %08x / \"%s\"", s.ShortName(), d, volSerial, volName)
		}
	}

	return nil
}
