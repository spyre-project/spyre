package yara

import (
	yr "github.com/hillu/go-yara"
	"github.com/spf13/afero"

	"github.com/dcso/spyre/config"
	"github.com/dcso/spyre/log"
	"github.com/dcso/spyre/report"
	"github.com/dcso/spyre/scanner"
	"github.com/dcso/spyre/sortable"

	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() { scanner.RegisterFileScanner(&fileScanner{}) }

type fileScanner struct{ rules *yr.Rules }

func (s *fileScanner) Name() string { return "YARA-file" }

func (s *fileScanner) Init() error {
	var (
		paths sortable.Pathlist
		c     *yr.Compiler
		err   error
	)
	if c, err = yr.NewCompiler(); err != nil {
		return err
	}
	readRules := make(map[string]struct{})
	c.SetIncludeCallback(func(name, filename, namespace string) []byte {
		if filename != "" {
			log.Debugf("yara: init: File '%s' included from '%s'", name, filename)
		}
		if _, ok := readRules[name]; ok {
			log.Debugf("yara: init: %s has already been included; skipping.", name)
			return []byte{}
		}
		readRules[name] = struct{}{}
		f, err := config.Fs.Open(name)
		if err != nil {
			log.Errorf("yara: init: Open %s: %v", name, err)
			return nil
		}
		defer f.Close()
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			log.Errorf("yara: init: Read from %s: %v", name, err)
			return nil
		}
		return buf
	})
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
	if len(config.YaraFiles) > 0 {
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
	} else {
		log.Debug("reading yara rules from files from any file found")
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
		sort.Sort(paths)
	}
	if len(paths) == 0 {
		return errors.New("No YARA rule files found")
	}
	for _, path := range paths {
		// We use the include callback function to actually read files
		// because yr_compiler_add_string() does not accept a file
		// name.
		if err = c.AddString(fmt.Sprintf(`include "%s"`, path), ""); err != nil {
			return err
		}
	}
	if s.rules, err = c.GetRules(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (s *fileScanner) ScanFile(f afero.File) error {
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
		matches, err = s.rules.ScanFileDescriptor(fd, 0, 1*time.Minute)
	} else {
		var buf []byte
		if buf, err = ioutil.ReadAll(f); err != nil {
			report.AddFileInfo(f, "yara", "Error reading file",
				"error", err.Error())
			return err
		}
		matches, err = s.rules.ScanMem(buf, 0, 1*time.Minute)
	}
	for _, m := range matches {
		report.AddFileInfo(f, "yara", "YARA rule match",
			"rule", m.Rule)
	}
	return err
}
