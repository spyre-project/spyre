package report

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"

	"github.com/hillu/go-archive-zip-crypto"
	"github.com/spf13/afero"

	"encoding/json"
	"io"
	"os"
)

var collector *evidenceCollector

type evidenceCollector struct {
	file, password string
	maxsize        config.FileSize
	size           uint64
	writer         io.WriteCloser
	zipWriter      *zip.Writer
	manifest       map[string]string
	sums           map[string]struct{}
	done           bool
}

func (ec *evidenceCollector) initialize() error {
	if f, err := os.Create(ec.file); err != nil {
		return err
	} else {
		ec.writer = f
		ec.zipWriter = zip.NewWriter(ec.writer)
	}

	ec.manifest = make(map[string]string)
	ec.sums = make(map[string]struct{})
	return nil
}

func (ec *evidenceCollector) addFile(f afero.File, sum string) error {
	if ec.done {
		return nil
	}
	if ec.zipWriter == nil {
		if err := ec.initialize(); err != nil {
			log.Errorf("evidence: initialize: %s: %v", ec.file, err)
			return err
		}
	}

	if _, ok := ec.sums[sum]; !ok {
		fi, err := f.Stat()
		if err != nil {
			log.Errorf("evidence: Can't get size of %s", f.Name())
			return nil
		}
		if ec.size+uint64(fi.Size()) > uint64(ec.maxsize) {
			log.Noticef("evidence: Skipping %s (%d bytes) due to size constraints",
				f.Name(), fi.Size())
			ec.manifest[f.Name()] = "(skipped)"
			ec.sums[sum] = struct{}{}
			return nil
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			log.Debugf("evidence: Can't access %s: %v", f.Name(), err)
			return nil
		}
		if w, err := ec.zipWriter.Encrypt(
			"files/"+sum, ec.password, zip.AES256Encryption,
		); err != nil {
			return err
		} else if _, err := io.Copy(w, f); err != nil {
			return err
		}
		ec.size += uint64(fi.Size())
		ec.sums[sum] = struct{}{}
	}

	ec.manifest[f.Name()] = "files/" + sum

	return nil
}

func (ec *evidenceCollector) finalize() error {
	if ec.done {
		return nil
	}
	if ec.writer == nil {
		return nil
	}
	if len(ec.manifest) > 0 {
		if w, err := ec.zipWriter.Encrypt(
			"manifest/"+config.Global.Hostname, ec.password, zip.AES256Encryption,
		); err != nil {
			log.Errorf("evidence: Can't create manifest: %v", err)
		} else {
			e := json.NewEncoder(w)
			e.SetIndent("", "    ")
			if err := e.Encode(ec.manifest); err != nil {
				log.Errorf("evidence: Can't write manifest: %v", err)
			}
		}
	}
	ec.zipWriter.Close()
	ec.writer.Close()
	if ec.size == 0 {
		log.Debugf("evidence: Removing %s because no content was collected", ec.file)
		os.Remove(ec.file)
	}
	ec.done = true
	return nil
}
