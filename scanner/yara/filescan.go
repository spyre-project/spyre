package yara

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"

	yr "github.com/hillu/go-yara/v4"
	"github.com/spf13/afero"

	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func init() { scanner.RegisterFileScanner(&fileScanner{}) }

type fileScanner struct{ rules *yr.Rules }

func (s *fileScanner) Name() string { return "YARA-file" }

func (s *fileScanner) Init() error {
	var err error
	s.rules, err = compile(filescan, config.YaraFileRules)
	return err
}

func (s *fileScanner) ScanFile(f afero.File) error {
	var (
		matches yr.MatchRules
		err     error
		md5sum  string
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
	/*
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
	*/
	if f, ok := f.(*os.File); ok {
		fd := f.Fd()
		err = s.rules.ScanFileDescriptor(fd, 0, 1*time.Minute, &matches)
		hash := md5.New()
		_, _ = io.Copy(hash, file)
		md5sum = fmt.Sprintf("%x", md5.Sum(nil))
	} else {
		var buf []byte
		if buf, err = ioutil.ReadAll(f); err != nil {
			report.AddFileInfo(f, "yara", "Error reading file",
				"error", err.Error())
			return err
		}
		err = s.rules.ScanMem(buf, 0, 1*time.Minute, &matches)
		md5sum = fmt.Sprintf("%x", md5.Sum(buf))
	}
	for _, m := range matches {
		var matchx []string
		for _, ms := range m.Strings {
			if stringInSlice(ms.Name+"-->"+string(ms.Data), matchx) {
				matchx = append(matchx, ms.Name+"-->"+string(ms.Data))
			}
		}
		matched := strings.Join(matchx[:], " | ")
		report.AddFileInfo(f, "yara", "YARA rule match",
			"rule", m.Rule, "hash", string(md5sum), "string_match", string(matched))
	}
	return err
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.EqualFold(b, a) {
			return false
		}
	}
	return true
}
