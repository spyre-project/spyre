package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// ReadIOCs reads IOCs from iocFile into iocs. iocs is typically a
// struct with a single member, most likely string-based map or slice,
// which is tagged with the name of a subkey.
//
// Example:
//
// type iocs struct {
// 	EventObjects []keyIOC `json:"registry-keys"`
// }
//
// type keyIOC struct { ... }
func ReadIOCs(iocFile string, iocs interface{}) error {
	f, err := Fs.Open(iocFile)
	if err != nil {
		return fmt.Errorf("open: %s: %v", iocFile, err)
	}
	jsondata, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("read: %s: %v", iocFile, err)
	}
	if err := json.Unmarshal(jsondata, &iocs); err != nil {
		return fmt.Errorf("parse: %s: %v", iocFile, err)
	}
	return nil
}
