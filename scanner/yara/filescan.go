package yara

import (
	yr "github.com/hillu/go-yara/v4"
	"github.com/spf13/afero"

	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func init() {
	scanner.RegisterFileScanner(&fileScanner{})
}

type fileScanner struct {
	RuleFiles      []string `yaml:"rule-files"`
	FailOnWarnings bool     `yaml:"fail-on-warnings"`
	rules          *yr.Rules
}

func (s *fileScanner) FriendlyName() string { return "YARA-file" }
func (s *fileScanner) ShortName() string    { return "yara" }

func (s *fileScanner) Init(c *config.ScannerConfig) error {
	var err error
	s.RuleFiles = []string{"filescan.yar"}
	s.FailOnWarnings = true
	if err = c.Config.Decode(s); err != nil {
		return err
	}
	s.rules, err = compile(filescan, s.RuleFiles, s.FailOnWarnings)
	return err
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
		report.AddStringf("yara: %s: Error accessing file information, error=%s",
			f.Name(), err.Error())
		return err
	}
	if int64(config.Global.MaxFileSize) > 0 && fi.Size() > int64(config.Global.MaxFileSize) {
		report.AddStringf("yara: %s: Skipping large file, size=%d, max_size=%d",
			f.Name(), fi.Size(),
			strconv.FormatInt(int64(config.Global.MaxFileSize), 10))
		return nil
	}
	if f, ok := f.(*os.File); ok {
		fd := f.Fd()
		err = s.rules.ScanFileDescriptor(fd, 0, 1*time.Minute, &matches)
	} else {
		var buf []byte
		if buf, err = ioutil.ReadAll(f); err != nil {
			report.AddStringf("yara: %s: Error reading file, error=%s",
				f.Name(), err.Error())
			return err
		}
		err = s.rules.ScanMem(buf, 0, 1*time.Minute, &matches)
	}
	for _, m := range matches {
		report.AddFileInfo(f, "yara", "YARA rule match", "rule", m.Rule)
	}
	return err
}
