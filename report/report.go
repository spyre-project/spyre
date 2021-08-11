package report

import (
	"github.com/spyre-project/spyre/config"
	"github.com/spyre-project/spyre/log"

	"github.com/mitchellh/go-ps"
	"github.com/spf13/afero"

	"strings"
)

var targets []target

func Init() error {
	var outfiles []string
	for _, spec := range config.Global.ReportTargets {
		tgt, err := mkTarget(spec)
		if err != nil {
			return err
		}
		targets = append(targets, tgt)
		outfiles = append(outfiles, tgt.path)
	}
	log.Noticef("Writing report to %s", strings.Join(outfiles, ", "))
	return nil
}

// AddStringf adds a single message with fmt.Printf-style parameters.
func AddStringf(f string, v ...interface{}) {
	for _, t := range targets {
		t.formatMessage(t.writer, f, v...)
	}
}

func AddFileInfo(file afero.File, description, message string, extra ...string) {
	for _, t := range targets {
		t.formatFileEntry(t.writer, file, description, message, extra...)
	}
}

func AddProcInfo(proc ps.Process, description, message string, extra ...string) {
	for _, t := range targets {
		t.formatProcEntry(t.writer, proc, description, message, extra...)
	}
}

// Close shuts down all reporting targets
func Close() {
	for _, t := range targets {
		t.finish(t.writer)
		t.writer.Close()
	}
}
