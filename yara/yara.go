package yara

import (
	yr "github.com/hillu/go-yara"
	"github.com/spf13/afero"

	"github.com/dcso/spyre/config"
	"github.com/dcso/spyre/log"
	"github.com/dcso/spyre/report"
	"github.com/dcso/spyre/sortable"

	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	rules *yr.Rules
)

func Init() error {
	var (
		paths sortable.Pathlist
		c     *yr.Compiler
		err   error
	)
	if c, err = yr.NewCompiler(); err != nil {
		return err
	}
	for _, v := range []struct {
		name  string
		value interface{}
	}{
		{"filename", ""},
		{"filepath", ""},
		{"extension", ""},
		{"filetype", ""},
	} {
		if err = c.DefineVariable(v.name, v.value); err != nil {
			return err
		}
	}
	afero.Walk(config.Fs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error(err)
			return nil
		}
		if info.IsDir() {
			if info.Mode()&os.ModeSymlink != 0 {
				return filepath.SkipDir
			}
			return nil
		}
		p := strings.ToLower(path)
		if strings.HasSuffix(p, ".yr") ||
			strings.HasSuffix(p, ".yar") ||
			strings.HasSuffix(p, ".yara") {
			log.Debugf("yara: init: Adding %s", path)
			paths = append(paths, path)
		}
		return nil
	})
	if len(paths) == 0 {
		err := errors.New("No YARA rule files found")
		log.Errorf("yara: init: %v", err)
		return err
	}
	sort.Sort(paths)
	for _, path := range paths {
		var buf []byte
		if buf, err = afero.ReadFile(config.Fs, path); err != nil {
			log.Errorf("yara: init: Could not read %s: %s", path, err)
			return err
		}
		if err = c.AddString(string(buf), ""); err != nil {
			log.Errorf("yara: init: Could not parse %s: %s", path, err)
			return err
		}
	}
	if rules, err = c.GetRules(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func ScanFile(f afero.File) error {
	var (
		matches []yr.MatchRule
		err     error
	)
	for _, v := range []struct {
		name  string
		value interface{}
	}{
		{"filename", filepath.ToSlash(filepath.Base(f.Name()))},
		{"filepath", filepath.ToSlash(f.Name())},
		{"extension", filepath.Ext(f.Name())},
	} {
		if err = rules.DefineVariable(v.name, v.value); err != nil {
			return err
		}
	}
	fi, err := f.Stat()
	if err != nil {
		report.AddFileInfo(f, "yara", "Error accessing file information",
			"error", err.Error())
		return err
	}
	if config.MaxFileSize > 0 && fi.Size() > config.MaxFileSize {
		report.AddFileInfo(f, "yara", "Skipping large file",
			"max_size", strconv.Itoa(int(config.MaxFileSize)))
	}
	if f, ok := f.(*os.File); ok {
		fd := f.Fd()
		matches, err = rules.ScanFileDescriptor(fd, 0, 1*time.Minute)
	} else {
		var buf []byte
		if buf, err = ioutil.ReadAll(f); err != nil {
			report.AddFileInfo(f, "yara", "Error reading file",
				"error", err.Error())
			return err
		}
		matches, err = rules.ScanMem(buf, 0, 1*time.Minute)
	}
	for _, m := range matches {
		report.AddFileInfo(f, "yara", "YARA rule match",
			"rule", m.Rule)
	}
	return err
}
