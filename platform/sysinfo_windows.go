// +build windows

package platform

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func GetSystemInformation() SystemInformation {
	osi := windows.RtlGetVersion()
	return SystemInformation{
		{"version", fmt.Sprintf("%d.%d", osi.MajorVersion, osi.MinorVersion)},
		{"build", fmt.Sprintf("%d", osi.BuildNumber)},
		{"platform_id", fmt.Sprintf("%d", osi.PlatformId)},
		// CsdVersion       [128]uint16
		{"servicepack", fmt.Sprintf("%d.%d", osi.ServicePackMajor, osi.ServicePackMinor)},
		{"suite", fmt.Sprintf("%04x", osi.SuiteMask)},
		{"product_type", fmt.Sprintf("%02x", osi.ProductType)},
	}
}
