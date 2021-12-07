package scanner

import (
	"github.com/spyre-project/spyre"
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/report"

	"github.com/mitchellh/go-ps"
	"github.com/spf13/afero"

	"errors"
	"io"
)

type Scanner interface {
	// used in logs
	FriendlyName() string
	// used as config section
	ShortName() string
	Init(*config.ScannerConfig) error
}

// SystemScanner scans are run right after Spyre initialization. They
// are desgined to check for "simple" queryable facts such as Mutexes
// etc.
type SystemScanner interface {
	Scanner
	Scan() error
}

// FileScanner scans are run after SystemScanner scans. The ScanFile
// method is run for every file.
type FileScanner interface {
	Scanner
	ScanFile(afero.File) error
}

// ProcScanner scans are run after SystemScanner scans. The ScanProc
// ismethod is run for every process that can be accessed, except for
// Spyre itself.
type ProcScanner interface {
	Scanner
	ScanProc(ps.Process) error
}

var (
	systemScanners []SystemScanner
	fileScanners   []FileScanner
	procScanners   []ProcScanner
)

// RegisterSystemScanner is called by a system scanner module's init()
// function to register the module so that it is called via the
// ScanSystem function
func RegisterSystemScanner(s SystemScanner) { systemScanners = append(systemScanners, s) }

// RegisterFileScanner is called by a file scanner module's init()
// function to register the module so that it is called via the
// ScanFile function
func RegisterFileScanner(s FileScanner) { fileScanners = append(fileScanners, s) }

// RegisterProcScanner is called by a proc scanner module's init()
// function to register the module so that it is called via the
// ScanProc function
func RegisterProcScanner(s ProcScanner) { procScanners = append(procScanners, s) }

func InitModules() error {
	var ss []SystemScanner
	for _, s := range systemScanners {
		sn, fn := s.ShortName(), s.FriendlyName()
		conf := config.Global.SystemScanners[sn]
		if conf.Disabled {
			log.Debugf("Skipping system scan module %s.", fn)
			continue
		}
		log.Debugf("Initializing system scan module %s ...", fn)
		if err := s.Init(&conf); err != nil {
			log.Infof("Error initializing %s module: %v", fn, err)
			continue
		}
		ss = append(ss, s)
	}
	systemScanners = ss
	var fs []FileScanner
	for _, s := range fileScanners {
		sn, fn := s.ShortName(), s.FriendlyName()
		conf := config.Global.FileScanners[sn]
		if conf.Disabled {
			log.Debugf("Skipping file scan module %s.", fn)
			continue
		}
		log.Debugf("Initializing file scan module %s ...", fn)
		if err := s.Init(&conf); err != nil {
			log.Infof("Error initializing %s module: %v", fn, err)
			continue
		}
		fs = append(fs, s)
	}
	fileScanners = fs
	var ps []ProcScanner
	for _, s := range procScanners {
		sn, fn := s.ShortName(), s.FriendlyName()
		conf := config.Global.ProcScanners[sn]
		if conf.Disabled {
			log.Debugf("Skipping process scan module %s.", fn)
			continue
		}
		log.Debugf("Initializing process scan module %s ...", fn)
		if err := s.Init(&conf); err != nil {
			log.Infof("Error initializing %s module: %v", fn, err)
			continue
		}
		ps = append(ps, s)
	}
	procScanners = ps
	if len(systemScanners)+len(fileScanners)+len(procScanners) == 0 {
		return errors.New("No scan modules were initialized")
	}
	return nil
}

func ScanSystem() (err error) {
	for _, s := range systemScanners {
		if e := s.Scan(); err == nil && e != nil {
			err = e
		}
	}
	return
}

func ScanFile(path string) (err error) {
	f, err := spyre.FS.Open(path)
	if err != nil {
		log.Debugf("Error: %v", err)
		report.AddStringf("Error: %v", err)
		report.Stats.File.NoAccess++
		return err
	}
	defer f.Close()
	for _, s := range fileScanners {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			log.Errorf("Could not seek to start of file %s: %v", path, err)
			report.Stats.File.NoAccess++
			return err
		}
		if e := s.ScanFile(f); err == nil && e != nil {
			err = e
		}
	}
	return
}

func ScanProc(proc ps.Process) (err error) {
	for _, s := range procScanners {
		if e := s.ScanProc(proc); err == nil && e != nil {
			err = e
		}
	}
	return
}
