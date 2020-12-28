package yara

import (
	yr "github.com/hillu/go-yara/v4"

	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"
	"github.com/spyre-project/spyre/log"

	"time"
)

func init() { scanner.RegisterEvtxScanner(&evtxScanner{}) }

type evtxScanner struct{ rules *yr.Rules }

func (s *evtxScanner) Name() string { return "YARA-evtx" }

func (s *evtxScanner) Init() error {
	var err error
	s.rules, err = compile(evtxscan, config.YaraEvtxRules)
	return err
}

func (s *evtxScanner) ScanEvtx(evt string) error {
	var (
		matches yr.MatchRules
		err     error
	)
	log.Noticef("detect yara: %s", evt)
  err = s.rules.ScanMem([]byte(evt), 0, 1*time.Minute, &matches)
	for _, m := range matches {
		report.AddEvtxInfo(evt, "yara", "YARA rule match",
			"rule", m.Rule)
	}
	return err
}
