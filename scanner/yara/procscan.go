package yara

import (
	"strings"

	yr "github.com/lprat/go-yara/v4"
	"github.com/mitchellh/go-ps"
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"time"
)

func init() { scanner.RegisterProcScanner(&procScanner{}) }

type procScanner struct{ rules *yr.Rules }

func (s *procScanner) Name() string { return "YARA-proc" }

func (s *procScanner) Init() error {
	var err error
	s.rules, err = compile(procscan, config.YaraProcRules)
	return err
}

func (s *procScanner) ScanProc(proc ps.Process) error {
	var matches yr.MatchRules
	pid, exe := proc.Pid(), proc.Executable()
	for _, v := range []struct {
		name  string
		value interface{}
	}{
		{"pid", pid},
		{"executable", exe},
	} {
		if err := s.rules.DefineVariable(v.name, v.value); err != nil {
			return err
		}
	}
	err := s.rules.ScanProc(pid, yr.ScanFlagsProcessMemory, 1*time.Minute, &matches)
	for _, m := range matches {
		var matchx []string
		for _, ms := range m.Strings {
			if stringInSlice(ms.Name+"-->"+string(ms.Data), matchx) {
				matchx = append(matchx, ms.Name+"-->"+string(ms.Data))
			}
		}
		matched := strings.Join(matchx[:], " | ")
		message := m.Rule+" (yara) matched on process: "+p.Executable()
		report.AddProcInfo(proc, "yara_on_pid", message, "rule", m.Rule, "string_match", string(matched))
	}
	return err
}
