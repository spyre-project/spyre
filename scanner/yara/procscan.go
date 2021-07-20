package yara

import (
	yr "github.com/hillu/go-yara/v4"
	"github.com/mitchellh/go-ps"

	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"time"
)

var procRules = config.StringSlice([]string{"yara-proc-rules"})

func init() {
	scanner.RegisterProcScanner(&procScanner{})
}

type procScanner struct {
	RuleFiles      []string `yaml:"rule-files"`
	FailOnWarnings bool     `yaml:"fail-on-warnings"`
	rules          *yr.Rules
}

func (s *procScanner) FriendlyName() string { return "YARA-proc" }
func (s *procScanner) ShortName() string    { return "yara" }

func (s *procScanner) Init(c *config.ScannerConfig) error {
	s.RuleFiles = []string{"procscan.yar"}
	s.FailOnWarnings = true
	var err error
	if err = c.Config.Decode(s); err != nil {
		return err
	}
	s.rules, err = compile(procscan, s.RuleFiles, s.FailOnWarnings)
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
		report.AddProcInfo(proc, "yara", "YARA rule match", "rule", m.Rule)
	}
	return err
}
