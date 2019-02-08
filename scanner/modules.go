package scanner

import (
	"github.com/dcso/spyre/log"

	"github.com/spf13/afero"

	"errors"
)

// SystemScanner scans are run right after Spyre initialization. They
// are desgined to check for "simple" queryable facts such as Mutexes
// etc.
type SystemScanner interface {
	Name() string
	Init() error
	Scan() error
}

// FileScanner scans are run after SystemScanner scans. The Scan
// method is run for every file.
type FileScanner interface {
	Name() string
	Init() error
	ScanFile(afero.File) error
}

var (
	systemScanners []SystemScanner
	fileScanners   []FileScanner
)

// RegisterSystemScanner is called by a system scanner module's init()
// function to register the module so that it is called via the
// ScanSystem function
func RegisterSystemScanner(s SystemScanner) { systemScanners = append(systemScanners, s) }

// RegisterFileScanner is called by a file scanner module's init()
// function to register the module so that it is called via the
// ScanFile function
func RegisterFileScanner(s FileScanner) { fileScanners = append(fileScanners, s) }

func InitModules() error {
	var ss []SystemScanner
	for _, s := range systemScanners {
		log.Debugf("Initializing module %s ...", s.Name())
		if err := s.Init(); err != nil {
			log.Infof("Error initializing %s module: %v", s.Name(), err)
			continue
		}
		ss = append(ss, s)
	}
	systemScanners = ss
	var fs []FileScanner
	for _, s := range fileScanners {
		log.Debugf("Initializing module %s ...", s.Name())
		if err := s.Init(); err != nil {
			log.Infof("Error initializing %s module: %v", s.Name(), err)
			continue
		}
		fs = append(fs, s)
	}
	fileScanners = fs
	if len(systemScanners) == 0 && len(fileScanners) == 0 {
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

func ScanFile(f afero.File) (err error) {
	for _, s := range fileScanners {
		if e := s.ScanFile(f); err == nil && e != nil {
			err = e
		}
	}
	return
}
