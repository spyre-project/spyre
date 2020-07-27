package yara

import (
	yr "github.com/hillu/go-yara/v4"
	"github.com/spf13/afero"

	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() { scanner.RegisterFileScanner(&fileScanner{}) }

type fileScanner struct{ rules *yr.Rules }

func (s *fileScanner) Name() string { return "YARA-file" }

func (s *fileScanner) Init() error {
	var (
		paths []string
		c     *yr.Compiler
		err   error
	)
	if c, err = yr.NewCompiler(); err != nil {
		return err
	}
	is := &includeState{fs: config.Fs}
	c.SetIncludeCallback(is.IncludeCallback)
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
	log.Debugf("reading yara rules from specified files: %s", strings.Join(config.YaraFiles, ", "))
	for _, path := range config.YaraFiles {
		if fi, err := config.Fs.Stat(path); err != nil {
			log.Errorf("yara: init: %v", err)
			return err
		} else if fi.IsDir() {
			log.Errorf("yara: init: %s is a directory", path)
		}
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		return errors.New("No YARA rule files found")
	}
	for _, path := range paths {
		// We use the include callback function to actually read files
		// because yr_compiler_add_string() does not accept a file
		// name.
		log.Debugf("yara: init: Adding %s", path)
		if err = c.AddString(fmt.Sprintf(`include "%s"`, path), ""); err != nil {
			return err
		}
	}
	if s.rules, err = c.GetRules(); err != nil {
		return err
	}
	if len(s.rules.GetRules()) == 0 {
		return errors.New("No YARA rules defined")
	}
	return nil
}

func (s *fileScanner) ScanFile(f afero.File) error {
	var (
		matches yr.MatchRules
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
		if err = s.rules.DefineVariable(v.name, v.value); err != nil {
			return err
		}
	}
	fi, err := f.Stat()
	if err != nil {
		report.AddFileInfo(f, "yara", "Error accessing file information",
			"error", err.Error())
		return err
	}
	if int64(config.MaxFileSize) > 0 && fi.Size() > int64(config.MaxFileSize) {
		report.AddFileInfo(f, "yara", "Skipping large file",
			"max_size", strconv.Itoa(int(config.MaxFileSize)))
	}
	if f, ok := f.(*os.File); ok {
		fd := f.Fd()
		err = s.rules.ScanFileDescriptor(fd, 0, 1*time.Minute, &matches)
	} else {
		var buf []byte
		if buf, err = ioutil.ReadAll(f); err != nil {
			report.AddFileInfo(f, "yara", "Error reading file",
				"error", err.Error())
			return err
		}
		err = s.rules.ScanMem(buf, 0, 1*time.Minute, &matches)
	}
	for _, m := range matches {
		report.AddFileInfo(f, "yara", "YARA rule match",
			"rule", m.Rule)
	}
	return err
}
