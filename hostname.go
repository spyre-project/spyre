package spyre

import (
	"os"
)

var (
	DefaultHostname string
	Hostname        string
)

func init() {
	var err error
	if DefaultHostname, err = os.Hostname(); err != nil {
		DefaultHostname = "<unknown-hostname>"
	}
}
