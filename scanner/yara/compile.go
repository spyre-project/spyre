package yara

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"

	yr "github.com/hillu/go-yara/v4"
	"github.com/spf13/afero"

	"io/ioutil"
	"path/filepath"

	"errors"
	"fmt"
	"strings"
)

// includeState tracks the current working directory of including and
// included files. It works around a limitation in YARA's
// YR_COMPILER_INCLUDE_CALLBACK_FUNC.
type includeState struct {
	fs       afero.Fs
	cwd      string
	included []string
}

func (is *includeState) IncludeCallback(name, filename, namespace string) []byte {
	if filename == "" {
		is.cwd = "/"
	}
	name = filepath.Join(is.cwd, name)
	is.cwd = filepath.Dir(name)
	if filename != "" {
		log.Debugf("yara: init: File '%s' included from '%s'", name, filename)
	}
	for _, file := range is.included {
		if name == file {
			log.Debugf("yara: init: %s has already been included; skipping.", name)
			return []byte{}
		}
	}
	is.included = append(is.included, name)
	f, err := is.fs.Open(name)
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
}

type extvardefs map[string]interface{}

const filescan = 0
const procscan = 1
const evtxscan = 2

var extvars = map[int]extvardefs{
	filescan: extvardefs{
		"filename":  "",
		"filepath":  "",
		"extension": "",
		"filetype":  "",
	},
	procscan: extvardefs{
		"pid":        -1,
		"executable": "",
	},
	evtxscan: extvardefs{},
}

func compile(purpose int, inputfiles []string) (*yr.Rules, error) {
	var c *yr.Compiler
	var err error
	var paths []string
	if c, err = yr.NewCompiler(); err != nil {
		return nil, err
	}
	is := &includeState{fs: config.Fs}
	c.SetIncludeCallback(is.IncludeCallback)

	for k, v := range extvars[purpose] {
		if err = c.DefineVariable(k, v); err != nil {
			return nil, err
		}
	}

	log.Debugf("reading yara rules from specified files: %s", strings.Join(inputfiles, ", "))
	for _, path := range inputfiles {
		if fi, err := config.Fs.Stat(path); err != nil {
			log.Errorf("yara: init: %v", err)
			return nil, err
		} else if fi.IsDir() {
			log.Errorf("yara: init: %s is a directory", path)
		}
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		return nil, errors.New("No YARA rule files found")
	}
	for _, path := range paths {
		// We use the include callback function to actually read files
		// because yr_compiler_add_string() does not accept a file
		// name.
		log.Debugf("yara: init: Adding %s", path)
		if err = c.AddString(fmt.Sprintf(`include "%s"`, path), ""); err != nil {
			return nil, err
		}
	}
	purposeStr := [...]string{"file", "process", "evtx"}[purpose]
	rs, err := c.GetRules()
	if err != nil {
		for _, e := range c.Errors {
			log.Errorf("YARA compiler error in %s ruleset: %s:%d %s",
				purposeStr, e.Filename, e.Line, e.Text)
		}
		return nil, fmt.Errorf("%d YARA compiler errors(s) found, rejecting %s ruleset",
			len(c.Errors), purposeStr)
	}
	if len(c.Warnings) > 0 {
		for _, w := range c.Warnings {
			log.Warnf("YARA compiler warning in %s ruleset: %s:%d %s",
				purposeStr, w.Filename, w.Line, w.Text)
		}
		if config.YaraFailOnWarnings {
			return nil, fmt.Errorf("%d YARA compiler warning(s) found, rejecting %s ruleset",
				len(c.Warnings), purposeStr)
		}
	}
	if len(rs.GetRules()) == 0 {
		return nil, errors.New("No YARA rules defined")
	}
	return rs, nil
}
