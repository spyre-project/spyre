package report

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"

	"github.com/mitchellh/go-ps"
	"github.com/spf13/afero"

	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var targets []target

type fileInfo struct {
	file                 afero.File
	collectSize          int64
	description, message string
	extra                []string
}

var (
	fileInfoCh     chan fileInfo
	procFileInfoWg sync.WaitGroup
)

func procFileInfo(c chan fileInfo) {
	for fi := range c {
		fs := afero.NewOsFs()
		if f, err := fs.Open(fi.file.Name()); err != nil {
			log.Errorf("open: %s: %v", fi.file.Name(), err)
		} else {
			if finfo, err := f.Stat(); err != nil {
				log.Errorf("stat: %s: %v", fi.file.Name(), err)
			} else {
				fi.extra = append(fi.extra, "file_size", strconv.Itoa(int(finfo.Size())))
				fi.extra = append(fi.extra, "last_written", finfo.ModTime().UTC().Format(time.RFC3339))
			}
			if sum, err := hashFile(f); err != nil {
				log.Errorf("hash: %s: %v", f.Name(), err)
			} else {
				sum := hex.EncodeToString(sum)
				fi.extra = append(fi.extra, "sha256", sum)
				if fi.collectSize != 0 && collector != nil {
					if err := collector.addFile(f, sum, fi.collectSize); err != nil {
						log.Errorf("Cannot write evidence to file: %v", err)
						collector.finalize()
						collector = nil
					}
				}
			}
		}
		for _, t := range targets {
			t.formatFileEntry(t.writer, fi.file, fi.description, fi.message, fi.extra...)
		}
	}
	procFileInfoWg.Done()
}

func Init() error {
	fileInfoCh = make(chan fileInfo, 32)
	go procFileInfo(fileInfoCh)
	procFileInfoWg.Add(1)

	var outfiles []string
	for _, spec := range config.Global.ReportTargets {
		tgt, err := mkTarget(spec)
		if err != nil {
			return err
		}
		targets = append(targets, tgt)
		if fw, ok := tgt.writer.(*fileWriter); ok {
			outfiles = append(outfiles, fw.path)
		}
	}
	log.Noticef("Writing report to %s", strings.Join(outfiles, ", "))

	if ec := config.Global.EvidenceCollection; !ec.Disabled {
		collector = &evidenceCollector{file: ec.File, password: ec.Password, maxsize: ec.MaxSize}
		collector.file = filepath.FromSlash(expand(collector.file))
		if collector.password == "" {
			collector.password = "infected"
		}
		if collector.maxsize == 0 {
			collector.maxsize = 1024 * 1024 * 1024 // 1 GB
		}
		log.Noticef("Collecting evidence (if any) to %s; password='%s'; max-size=%s",
			collector.file, collector.password, &collector.maxsize)
	}

	return nil
}

// AddStringf adds a single message with fmt.Printf-style parameters.
func AddStringf(f string, v ...interface{}) {
	for _, t := range targets {
		t.formatMessage(t.writer, f, v...)
	}
}

func AddFileInfo(file afero.File, collectSize int64, description, message string, extra ...string) {
	Stats.File.Matches++
	fileInfoCh <- fileInfo{file, collectSize, description, message, extra}
}

func AddProcInfo(proc ps.Process, description, message string, extra ...string) {
	Stats.Process.Matches++
	for _, t := range targets {
		t.formatProcEntry(t.writer, proc, description, message, extra...)
	}
}

func AddNetstatInfo(description, message string, extra ...string) {
	for _, t := range targets {
		t.formatNetstatEntry(t.writer, description, message, extra...)
	}
}

// Close shuts down all reporting targets
func Close() {
	close(fileInfoCh)
	procFileInfoWg.Wait()
	ts := time.Now().Format("2006-01-02 15:04:05.000 -0700 MST")
	log.Infof("Scan finished at %s", ts)
	AddStringf("Scan finished at %s", ts)
	if collector != nil {
		collector.finalize()
	}
	for _, t := range targets {
		t.finish(t.writer)
		t.writer.Close()
	}
}
