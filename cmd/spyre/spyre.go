package main

import (
	"github.com/daviddengcn/go-colortext"
	"github.com/hillu/go-archive-zip-crypto"
	"github.com/mitchellh/go-ps"
	"github.com/spf13/afero"

	"github.com/spyre-project/spyre"
	"github.com/spyre-project/spyre/appendedzip"
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/platform"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"
	"github.com/spyre-project/spyre/zipfs"

	// Pull in scan modules
	_ "github.com/spyre-project/spyre/module_config"

	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func promptOnExit() {
	if !config.Global.UI.PromptOnExit {
		return
	}
	fmt.Print("Press ENTER to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func die() {
	fmt.Println()
	ct.Foreground(ct.Red, true)
	fmt.Println("Scan failed to complete.")
	ct.ResetColor()
	promptOnExit()
	os.Exit(1)
}

func main() {
	ourpid := os.Getpid()

	log.Infof("This is Spyre version %s, pid=%d", spyre.Version, ourpid)

	basename := stripExeSuffix(os.Args[0])
	if zr, err := appendedzip.OpenFile(platform.GetProgramFilename()); err == nil {
		log.Notice("using embedded zip for configuration")
		config.Fs = zipfs.New(zr, "infected")
	} else if zrc, err := zip.OpenReader(basename + ".zip"); err == nil {
		log.Noticef("using file %s.zip for configuration", basename)
		config.Fs = zipfs.New(&zrc.Reader, "infected")
	} else {
		abs, _ := filepath.Abs(
			filepath.Join(filepath.Dir(os.Args[0])),
		)
		log.Noticef("using directory %s for configuration", abs)
		config.Fs = afero.NewBasePathFs(afero.NewOsFs(), abs)
	}

	if err := config.Init(); err != nil {
		log.Errorf("Failed to parse configuration: %s", err)
		die()
	}
	displayLogo()

	log.Init()
	if m := config.Global.RulesetMarker; m != "" {
		log.Infof("Ruleset marker: %s", m)
	} else {
		log.Infof("Ruleset marker not specified")
	}

	if !config.Global.HighPriority {
		log.Notice("Setting low CPU, I/O priority...")
		platform.SetLowPriority()
	} else {
		log.Info("Running at regular CPU, I/O priority")
	}

	if err := report.Init(); err != nil {
		log.Errorf("Failed to initialize report target: %v", err)
		die()
	}

	if err := scanner.InitModules(); err != nil {
		log.Errorf("Initialize: %v", err)
		die()
	}

	report.AddStringf("This is Spyre version %s, running on host %s, pid=%d",
		spyre.Version, spyre.Hostname, ourpid)

	ts := time.Now().Format("2006-01-02 15:04:05.000 -0700 MST")
	log.Infof("Scan started at %s", ts)
	report.AddStringf("Scan started at %s", ts)

	if err := scanner.ScanSystem(); err != nil {
		log.Errorf("Error scanning system:: %v", err)
	}

	fs := afero.NewOsFs()
	for _, path := range config.Global.Paths {
		afero.Walk(fs, path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if platform.SkipDir(fs, path) {
					log.Noticef("Skipping %s", path)
					return filepath.SkipDir
				}
				return nil
			}
			const specialMode = os.ModeSymlink | os.ModeDevice | os.ModeNamedPipe | os.ModeSocket | os.ModeCharDevice
			if info.Mode()&specialMode != 0 {
				return nil
			}
			f, err := fs.Open(path)
			if err != nil {
				log.Debugf("Could not open %s", path)
				report.Stats.FileNoAccess++
				return nil
			}
			defer f.Close()
			log.Debugf("Scanning %s...", path)
			if err = scanner.ScanFile(f); err != nil {
				log.Errorf("Error scanning file: %s: %v", path, err)
			}
			return nil
		})
	}

	procs, err := ps.Processes()
	if err != nil {
		log.Errorf("Error while enumerating processes: %v", err)
	} else {
		for _, proc := range procs {
			pid := proc.Pid()
			exe := proc.Executable()
			if pid == ourpid {
				log.Debugf("Skipping process %s[%d].", exe, pid)
				continue
			}
			if sliceContains(config.Global.ProcIgnoreNames, exe) {
				log.Debugf("Skipping process (found on ignore list) %s[%d].", exe, pid)
				continue
			}
			log.Debugf("Scanning process %s[%d]...", exe, pid)
			if err := scanner.ScanProc(proc); err != nil {
				log.Errorf("Error scanning %s[%d]: %v", exe, pid, err)
			}
		}
	}

	report.Close()

	fmt.Println()
	if report.Stats.FileEntries > 0 || report.Stats.ProcEntries > 0 {
		ct.Foreground(ct.Yellow, true)
	} else {
		ct.Foreground(ct.Green, true)
	}
	fmt.Printf("Scan completed with %d file findings and %d process findings\n",
		report.Stats.FileEntries, report.Stats.ProcEntries,
	)
	ct.ResetColor()
	fmt.Printf("%d files could not be accessed.\n", report.Stats.FileNoAccess)
	promptOnExit()
}

func sliceContains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
