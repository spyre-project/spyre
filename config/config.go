package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/pflag"

	"github.com/spyre-project/spyre"
	"github.com/spyre-project/spyre/log"

	"os"
	"strings"
)

var (
	Paths              simpleStringSlice
	EvtxPaths          simpleStringSlice
	MaxFileSize        = fileSize(32 * 1024 * 1024)
	ReportTargets      = simpleStringSlice([]string{"spyre.log"})
	Hostname           string
	HighPriority       bool
	YaraFailOnWarnings bool
	YaraFileRules      simpleStringSlice = []string{"filescan.yar"}
	YaraProcRules      simpleStringSlice = []string{"procscan.yar"}
	YaraEvtxRules      simpleStringSlice = []string{"evtxscan.yar"}
	ProcIgnoreList     simpleStringSlice
	IocFiles           simpleStringSlice
	IgnorePath         string  = "ignorepath.txt"
)

// Fs is the "filesystem" in which configuration and rules are found.
// This can be provided through a ZIP file appended to the binary.
var Fs afero.Fs

func Init() error {
	Paths = simpleStringSlice(defaultPaths)
	EvtxPaths = simpleStringSlice(defaultEvtxPaths)
	pflag.VarP(&Paths, "path", "p", "paths to be scanned (default: / on Unix, all fixed drives on Windows)")
	pflag.VarP(&EvtxPaths, "evtxpath", "e", "paths of evtx (Windows only)")
	pflag.Var(&YaraFileRules, "yara-file-rules",
		"yara files to be used for file scan (default: filescan.yar)")
	pflag.Var(&YaraProcRules, "yara-proc-rules",
		"yara files to be used for proc scan (default: procscan.yar)")
	pflag.Var(&YaraEvtxRules, "yara-evtx-rules",
		"yara files to be used for evtx scan (default: evtxscan.yar)")
	pflag.Var(&IocFiles, "ioc-files",
		"IOC files to be used for descriptive IOCs (default: ioc.json)")
	pflag.Var(&MaxFileSize, "max-file-size",
		"maximum size of individual files to be scanned, turn off by setting to 0 or negative value")
	pflag.StringVar(&spyre.Hostname, "set-hostname", spyre.DefaultHostname, "hostname")
	pflag.VarP(&log.GlobalLevel, "loglevel", "l", "loglevel")
	pflag.VarP(&ReportTargets, "report", "r", "report target(s)")
	pflag.BoolVar(&HighPriority, "high-priority", false,
		"run at high priority instead of giving up CPU and I/O resources to other processes")
	pflag.BoolVar(&YaraFailOnWarnings, "yara-fail-on-warnings", false,
		"fail if yara emits a warning on at least one rule")
	pflag.Var(&ProcIgnoreList, "proc-ignore", "Names of processes to be ignored from scanning")
        pflag.StringVar(&IgnorePath, "path-ignore","ignorepath.txt" ,"file contains path to ignore")
	pflag.Var(&YaraFileRules, "yara-rule-files", "")
	pflag.CommandLine.MarkHidden("yara-rule-files")
	var args []string
	if len(os.Args) > 1 {
		log.Debug("Using user-provided command line parameters.")
		args = os.Args[1:]
	} else if buf, err := afero.ReadFile(Fs, "params.txt"); err != nil {
		log.Debug("Using default parameters.")
	} else {
		log.Debug("Using parametes form params.txt.")
		for _, line := range strings.Split(string(buf), "\n") {
			line = strings.TrimSpace(line)
			if len(line) == 0 || line[0] == '#' {
				continue
			}
			if tokens := strings.Fields(line); len(tokens) > 1 && !strings.Contains(tokens[0], "=") {
				args = append(args, tokens[0])
				args = append(args, strings.Join(tokens[1:], " "))
			} else {
				args = append(args, line)
			}
		}
	}
	pflag.CommandLine.Parse(args)

	pflag.VisitAll(func(f *pflag.Flag) {
		log.Debugf("config: --%s %s%s", f.Name, f.Value, map[bool]string{false: " (unchanged)"}[f.Changed])
	})

	log.Init()
	return nil
}
