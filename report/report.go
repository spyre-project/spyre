package report

import (
	"github.com/spf13/afero"
)

var Targets TargetList

func init() {
	Targets.Set("spyre.log")
	Targets.reset = true
}

// AddStringf adds a single message with fmt.Printf-style parameters.
func AddStringf(f string, v ...interface{}) {
	for _, t := range Targets.targets {
		t.formatMessage(t.writer, f, v...)
	}
}

func AddFileInfo(file afero.File, description, message string, extra ...string) {
	for _, t := range Targets.targets {
		t.formatFileEntry(t.writer, file, description, message, extra...)
	}
}

// Close shuts down all reporting targets
func Close() {
	for _, t := range Targets.targets {
		t.finish(t.writer)
		t.writer.Close()
	}
}
