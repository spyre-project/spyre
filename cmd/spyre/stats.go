package main

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"

	"fmt"
)

type Spinner int

func (s *Spinner) String() string {
	const spin = `\|/-`
	*s = Spinner((int(*s) + 1) % len(spin))
	return string(spin[*s])
}

var spinner = new(Spinner)

func printStats() {
	fmt.Printf("%s File: %d/%s; skip=%d, match=%d; Proc: %d, skip=%d, match=%d        \r",
		spinner,
		report.Stats.File.ScanCount,
		config.FileSize(report.Stats.File.ScanBytes),
		report.Stats.File.SkipCount,
		report.Stats.File.Matches,
		report.Stats.Process.ScanCount,
		report.Stats.Process.SkipCount,
		report.Stats.Process.Matches,
	)

}
