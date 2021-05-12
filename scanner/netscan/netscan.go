// +build linux amd64

package netscan

import (
	"fmt"
	"strings"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"
)

func init() { scanner.RegisterSystemScanner(&systemScanner{}) }

type systemScanner struct {
	iocs []eventIOC
}

type eventIOC struct {
	Dip         []string `json:"dip"`
	Sip         []string `json:"sip"`
	Sport       []int    `json:"sport"`
	Dport       []int    `json:"dport"`
	Pname       []string `json:"pname"`
	NPname      []string `json:"notpname"`
	State       []string `json:"state"`
	Proto       string   `json:"proto"`
	Description string   `json:"description"`
}

type iocFile struct {
	Keys []eventIOC `json:"netstat"`
}

func (s *systemScanner) Name() string { return "Netstat" }

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

func intInSlice(a int, list []int) bool {
	if len(list) == 0 {
		return true
	}
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func stringInSlice(a string, list []string) bool {
	if len(list) == 0 {
		return true
	}
	for _, b := range list {
		if strings.EqualFold(b, a) {
			return true
		}
	}
	return false
}

func nstringInSlice(a string, list []string) bool {
	if len(list) == 0 {
		return false
	}
	for _, b := range list {
		if strings.EqualFold(b, a) {
			return true
		}
	}
	return false
}

func (s *systemScanner) Scan() error {
	tsocks, err := netstat.TCPSocks(netstat.NoopFilter)
	if err != nil {
		log.Debugf("Error to get TCP socks : %s", err)
	}
	usocks, err := netstat.UDPSocks(netstat.NoopFilter)
	if err != nil {
		log.Debugf("Error to get UDP socks : %s", err)
	}
	for _, ioc := range s.iocs {
		//netCheck(ioc.Dip, ioc.Sip, ioc.Sport, ioc.Dport, ioc.Pname, ioc.State)
		for _, e := range tsocks {
			//fmt.Printf("%v\n", e)
			pid := "unknown"
			proc_name := "unknown"
			port_src := fmt.Sprintf("%d", e.LocalAddr.Port)
			port_dst := fmt.Sprintf("%d", e.RemoteAddr.Port)
			uid := fmt.Sprintf("%d", e.UID)

			if !(strings.EqualFold(ioc.Proto, "tcp") || ioc.Proto == "*" || ioc.Proto == "") {
				continue
			}
			if e.Process != nil {
				proc_name = fmt.Sprintf("%s", e.Process.Name)
				pid = fmt.Sprintf("%d", e.Process.Pid)
				if !(stringInSlice(e.Process.Name, ioc.Pname)) {
				  continue
			  }
				if nstringInSlice(e.Process.Name, ioc.NPname) {
					continue
				}
			}
			dip := e.RemoteAddr.IP.String()
			if !(stringInSlice(dip, ioc.Dip)) {
				continue
			}
			sip := e.LocalAddr.IP.String()
			if !(stringInSlice(sip, ioc.Sip)) {
				continue
			}
			if !(intInSlice(int(e.LocalAddr.Port), ioc.Sport)) {
				continue
			}
			if !(intInSlice(int(e.RemoteAddr.Port), ioc.Dport)) {
				continue
			}
			state := fmt.Sprintf("%s", e.State)
			if !(stringInSlice(state, ioc.State)) {
				continue
			}
			message := fmt.Sprintf("Found netstat rule: %s on TCP %v",ioc.Description, e)
			report.AddNetstatInfo("ioc_on_netstat", message,
				"rule", ioc.Description, "State", state, "ip_src", sip, "ip_dst", dip, "uid", uid, "PID", pid, "Process", proc_name, "port_dst", port_dst, "port_src", port_src, "proto", "TCP")
		}
		for _, e := range usocks {
			pid := "unknown"
			proc_name := "unknown"
			port_src := fmt.Sprintf("%d", e.LocalAddr.Port)
			port_dst := fmt.Sprintf("%d", e.RemoteAddr.Port)
			uid := fmt.Sprintf("%d", e.UID)
			//fmt.Printf("%v\n", e)
			if !(strings.EqualFold(ioc.Proto, "udp") || ioc.Proto == "*" || ioc.Proto == "") {
				continue
			}
			if e.Process != nil {
				proc_name = fmt.Sprintf("%s", e.Process.Name)
				pid = fmt.Sprintf("%d", e.Process.Pid)
				if !(stringInSlice(e.Process.Name, ioc.Pname)) {
				  continue
			  }
				if nstringInSlice(e.Process.Name, ioc.NPname) {
					continue
				}
			}
			dip := e.RemoteAddr.IP.String()
			if !(stringInSlice(dip, ioc.Dip)) {
				continue
			}
			sip := e.LocalAddr.IP.String()
			if !(stringInSlice(sip, ioc.Sip)) {
				continue
			}
			if !(intInSlice(int(e.LocalAddr.Port), ioc.Sport)) {
				continue
			}
			if !(intInSlice(int(e.RemoteAddr.Port), ioc.Dport)) {
				continue
			}
			state := fmt.Sprintf("%s", e.State)
			if !(stringInSlice(state, ioc.State)) {
				continue
			}
			message := fmt.Sprintf("Found netstat rule: %s on UDP %v",ioc.Description, e)
			report.AddNetstatInfo("ioc_on_netstat", message,
				"rule", ioc.Description, "State", state, "ip_src", sip, "ip_dst", dip, "uid", uid, "PID", pid, "Process", proc_name, "port_dst", port_dst, "port_src", port_src, "proto", "UDP")
		}
	}
	return nil
}
