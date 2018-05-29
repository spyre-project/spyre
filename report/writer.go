package report

import (
	"github.com/dcso/spyre/log"

	"io"
	"os"
)

type fileWriter struct {
	path     string
	disabled bool
	w        io.WriteCloser
}

func (fw *fileWriter) Write(buf []byte) (n int, e error) {
	n, e = len(buf), nil
	if fw.disabled {
		return
	}
	var err error
	if fw.w == nil {
		const mode = os.O_APPEND | os.O_WRONLY | os.O_CREATE | os.O_APPEND
		if fw.path == "-" {
			fw.w = os.Stdout
		} else if fw.w, err = os.OpenFile(fw.path, mode, 0666); err != nil {
			log.Errorf("Could not open report file %s: %s", fw.path, err)
			fw.disabled = true
			return
		}
	}
	if n, err = fw.w.Write(buf); err != nil {
		log.Errorf("Could not write to report file %s: %s", fw.path, err)
		fw.disabled = true
		fw.w.Close()
		fw.w = nil
	}
	return
}

func (fw *fileWriter) Close() error {
	if fw.w != nil {
		fw.w.Close()
		fw.w = nil
	}
	return nil
}
