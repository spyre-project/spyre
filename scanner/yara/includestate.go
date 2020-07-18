package yara

import (
	"github.com/spyre-project/spyre/log"

	"github.com/spf13/afero"

	"io/ioutil"
	"path/filepath"
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
