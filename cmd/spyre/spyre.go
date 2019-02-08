package main

import (
	"github.com/spf13/afero"

	"github.com/dcso/spyre"
	"github.com/dcso/spyre/appendedzip"
	"github.com/dcso/spyre/config"
	"github.com/dcso/spyre/log"
	"github.com/dcso/spyre/platform"
	"github.com/dcso/spyre/report"
	"github.com/dcso/spyre/scanner"
	"github.com/dcso/spyre/zipfs"

	// Pull in scan modules
	_ "github.com/dcso/spyre/module_config"

	"os"
	"path/filepath"
	"time"
)

func main() {
	log.Infof("This is Spyre version %s", version)

	if zr, err := appendedzip.OpenFile(os.Args[0]); err == nil {
		log.Notice("using embedded zip for configuration")
		config.Fs = zipfs.New(zr)
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
		setLowPriority()
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

	report.AddStringf("This is Spyre version %s, running on host %s", version, spyre.Hostname)
	defer report.Close()

	log.Infof("Scan started at %s", time.Now())
	report.AddStringf("Scan started at %s", time.Now())

	if err := scanner.ScanSystem(); err != nil {
		log.Errorf("Error scanning system:: %v", err)
	}

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
			const specialMode = os.ModeSymlink | os.ModeDevice | os.ModeNamedPipe | os.ModeSocket | os.ModeCharDevice
			if info.Mode()&specialMode != 0 {
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
	log.Infof("Scan finished at %s", time.Now())
	report.AddStringf("Scan finished at %s", time.Now())
}
