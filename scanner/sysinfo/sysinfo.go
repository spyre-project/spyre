// +build windows

package sysinfo

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/platform/sys"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"golang.org/x/sys/windows"

	"fmt"
	"strings"
	"unsafe"
)

func init() { scanner.RegisterSystemScanner(&systemScanner{}) }

type systemScanner struct{}

func (s *systemScanner) FriendlyName() string { return "Diag-Sysinfo" }
func (s *systemScanner) ShortName() string    { return "sysinfo" }

func (s *systemScanner) Init(*config.ScannerConfig) error { return nil }

func getAdaptersInfo() *windows.IpAdapterInfo {
	var l uint32
	windows.GetAdaptersInfo(nil, &l)
	if l == 0 {
		return nil
	}
	buf := make([]byte, int(l))
	ai := (*windows.IpAdapterInfo)(unsafe.Pointer(&buf[0]))
	if err := windows.GetAdaptersInfo(ai, &l); err != nil {
		return nil
	}
	return ai
}

func (s *systemScanner) Scan() error {
	for ai := getAdaptersInfo(); ai != nil; ai = ai.Next {
		var mac string
		for _, c := range ai.Address[:int(ai.AddressLength)] {
			if len(mac) > 0 {
				mac += ":"
			}
			mac += fmt.Sprintf("%02x", c)
		}
		var ipaddr string
		for ca := &ai.IpAddressList; ca != nil; ca = ca.Next {
			if len(ipaddr) > 0 {
				ipaddr += ";"
			}
			ipaddr += fmt.Sprintf("%s/%s",
				strings.Trim(string(ca.IpAddress.String[:]), " \t\n\000"),
				strings.Trim(string(ca.IpMask.String[:]), " \t\n\000"),
			)
		}
		report.AddStringf("%s: network interface: '%s'(%s): mac=%s, ipv4=%s",
			s.ShortName(),
			strings.Trim(string(ai.Description[:]), " \t\n\000"),
			strings.Trim(string(ai.AdapterName[:]), " \t\n\000"),
			mac, ipaddr,
		)
	}

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
