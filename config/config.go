package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/spyre-project/spyre"
	"github.com/spyre-project/spyre/log"

	"os"
)

var Global GlobalConfig
var FlagSet *pflag.FlagSet

type GlobalConfig struct {
	MaxFileSize     FileSize                 `yaml:"max-file-size"`
	Paths           StringSlice              `yaml:"paths"`
	ProcIgnoreNames StringSlice              `yaml:"proc-ignore-names"`
	ReportTargets   StringSlice              `yaml:"report"`
	Hostname        string                   `yaml:"hostname"`
	HighPriority    bool                     `yaml:"high-priority"`
	UI              UIConfig                 `yaml:"ui"`
	SystemScanners  map[string]ScannerConfig `yaml:"system"`
	FileScanners    map[string]ScannerConfig `yaml:"file"`
	ProcScanners    map[string]ScannerConfig `yaml:"proc"`
}

type UIConfig struct {
	PromptOnExit bool `yaml:"prompt-on-exit"`
}

type ScannerConfig struct {
	Disabled bool      `yaml:"disabled"`
	Config   yaml.Node `yaml:"config"`
}

func init() {
	Global.Paths = defaultPaths()
	Global.MaxFileSize = 32 * 1024 * 1024
	Global.ReportTargets = []string{"spyre_${hostname}_${time}.log"}

	FlagSet = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	// global config
	FlagSet.VarP(&Global.Paths, "path", "p", "paths to be scanned (default: / on Unix, all fixed drives on Windows)")
	FlagSet.Var(&Global.ProcIgnoreNames, "proc-ignore", "Names of processes to be ignored from scanning")
	FlagSet.Var(&Global.MaxFileSize, "max-file-size",
		"maximum size of individual files to be scanned, turn off by setting to 0 or negative value")
	FlagSet.BoolVar(&Global.HighPriority, "high-priority", false,
		"run at high priority instead of giving up CPU and I/O resources to other processes")
	FlagSet.StringVar(&spyre.Hostname, "set-hostname", spyre.DefaultHostname, "hostname")
	FlagSet.VarP(&Global.ReportTargets, "report", "r", "report target(s)")

	// not yet sorted
	FlagSet.VarP(&log.GlobalLevel, "loglevel", "l", "loglevel")
}

// Fs is the "filesystem" in which configuration and rules are found.
// This can be provided through a ZIP file appended to the binary.
var Fs afero.Fs

func Init() error {
	if len(os.Args) > 1 {
		log.Debug("Using user-provided command line parameters.")
		FlagSet.Parse(os.Args[1:])
	} else {
		log.Debug("Using default parameters.")
	}

	FlagSet.VisitAll(func(f *pflag.Flag) {
		log.Debugf("config: --%s %s%s", f.Name, f.Value, map[bool]string{false: " (unchanged)"}[f.Changed])
	})

	f, err := Fs.Open("spyre.yaml")
	if err != nil {
		log.Debugf("cannot open spyre.yaml: %v", err)
		return nil
	}

	if err := yaml.NewDecoder(f).Decode(&Global); err != nil {
		log.Errorf("cannot parse spyre.yaml: %v", err)
		return err
	}

	log.Debugf("Global config: %+v", Global)

	return nil
}
