package main

import (
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

	"os"
	"path/filepath"
	"time"
	"io/ioutil"
	"strings"
)

func main() {
	ourpid := os.Getpid()

	log.Infof("This is Spyre version %s, pid=%d", spyre.Version, ourpid)

	basename := stripExeSuffix(os.Args[0])
	if zr, err := appendedzip.OpenFile(os.Args[0]); err == nil {
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
		os.Exit(1)
	}

	if !config.HighPriority {
		log.Notice("Setting low CPU, I/O priority...")
		platform.SetLowPriority()
	} else {
		log.Info("Running at regular CPU, I/O priority")
	}

	if err := report.Init(); err != nil {
		log.Errorf("Failed to initialize report target: %v", err)
		os.Exit(1)
	}

	if err := scanner.InitModules(); err != nil {
		log.Errorf("Initialize: %v", err)
		os.Exit(1)
	}

	report.AddStringf("This is Spyre version %s, running on host %s, pid=%d",
		spyre.Version, spyre.Hostname, ourpid)
	defer report.Close()

	ts := time.Now().Format("2006-01-02 15:04:05.000 -0700 MST")
	log.Infof("Scan started at %s", ts)
	report.AddStringf("Scan started at %s", ts)

	if err := scanner.ScanSystem(); err != nil {
		log.Errorf("Error scanning system:: %v", err)
	}
        f, err := os.Open(config.IgnorePath)
	if err != nil {
		return log.Errorf("open: %s: %v", config.IgnorePath, err)
	}
	tmpdata, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return log.Errorf("read: %s: %v", config.IgnorePath, err)
	}
	IgnorePathValue := strings.Split(string(tmpdata), "\n")
	fs := afero.NewOsFs()
	for _, path := range config.Paths {
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
			if sliceContains(IgnorePathValue, path) {
				return nil
			}
			const specialMode = os.ModeSymlink | os.ModeDevice | os.ModeNamedPipe | os.ModeSocket | os.ModeCharDevice
			if info.Mode()&specialMode != 0 {
				return nil
			}
			if int64(config.MaxFileSize) > 0 && info.Size() > int64(config.MaxFileSize) {
				return nil
                        }
			f, err := fs.Open(path)
			if err != nil {
				log.Errorf("Could not open %s", path)
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
			if sliceContains(config.ProcIgnoreList, exe) {
				log.Debugf("Skipping process (found on ignore list) %s[%d].", exe, pid)
				continue
			}
			log.Debugf("Scanning process %s[%d]...", exe, pid)
			if err := scanner.ScanProc(proc); err != nil {
				log.Errorf("Error scanning %s[%d]: %v", exe, pid, err)
			}
		}
	}

	ts = time.Now().Format("2006-01-02 15:04:05.000 -0700 MST")
	log.Infof("Scan finished at %s", ts)
	report.AddStringf("Scan finished at %s", ts)
}

func sliceContains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
