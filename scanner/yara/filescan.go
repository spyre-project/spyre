package yara

import (
	yr "github.com/hillu/go-yara/v4"
	"github.com/spf13/afero"

	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/report"
	"github.com/spyre-project/spyre/scanner"

  "crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	fi, err := f.Stat()
	var datem = ""
	var content_file = ""
	if err == nil {
		date_tmp := fi.ModTime()
		datem = date_tmp.String()
	}
	if f, ok := f.(*os.File); ok {
		fd := f.Fd()
		err = s.rules.ScanFileDescriptor(fd, 0, 1*time.Minute, &matches)
		if matches != nil {
			var buf []byte
			if buf, err = ioutil.ReadAll(f); err == nil {
				md5sum = fmt.Sprintf("%x", md5.Sum(buf))
				content_file = base64.StdEncoding.EncodeToString(buf)
			}
		}
	} else {
		var buf []byte
		if buf, err = ioutil.ReadAll(f); err != nil {
			report.AddFileInfo(f, "yara", "Error reading file",
				"error", err.Error())
			return err
		}
		err = s.rules.ScanMem(buf, 0, 1*time.Minute, &matches)
		if matches != nil {
			md5sum = fmt.Sprintf("%x", md5.Sum(buf))
			content_file = base64.StdEncoding.EncodeToString(buf)
		}
	}
	for _, m := range matches {
		var matchx []string
		for _, ms := range m.Strings {
			if stringInSlice(ms.Name+"-->"+string(ms.Data), matchx) {
				matchx = append(matchx, ms.Name+"-->"+string(ms.Data))
			}
		}
		matched := strings.Join(matchx[:], " | ")
    message := m.Rule + " (yara) matched on file: " + f.Name() + " (" + string(md5sum) + ")"
		if strings.Contains(m.Rule,"_keepfile") {
		  report.AddFileInfo(f, "yara_on_file", message,
			  "rule", m.Rule, "Filehash", string(md5sum), "real_date", datem, "Filepath", f.Name(), "string_match", string(matched), "extracted_file", content_file)
	  } else {
			report.AddFileInfo(f, "yara_on_file", message,
				"rule", m.Rule, "Filehash", string(md5sum), "real_date", datem, "Filepath", f.Name(), "string_match", string(matched))
		}
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
