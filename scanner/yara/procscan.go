package yara

import (
	yr "github.com/hillu/go-yara/v4"
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

func (s *procScanner) ScanProc(pid int) error {
	var matches yr.MatchRules
	proc, err := ps.FindProcess(pid)
	if err != nil {
		return err
	}
	for _, v := range []struct {
		name  string
		value interface{}
	}{
		{"pid", pid},
		{"executable", proc.Executable()},
	} {
		if err = s.rules.DefineVariable(v.name, v.value); err != nil {
			return err
		}
	}
	err = s.rules.ScanProc(pid, yr.ScanFlagsProcessMemory, 1*time.Minute, &matches)
	for _, m := range matches {
		report.AddProcInfo(proc, "yara", "YARA rule match", "rule", m.Rule)
	}
	if err != nil {
		return err
	}
	return nil
}
