package report

import (
	"github.com/spf13/afero"

	"fmt"
	"io"
	"net/url"
	"strings"
)

type writer interface {
	write([]byte)
	finish()
}

type formatter interface {
	formatFileEntry(w io.Writer, f afero.File, description, message string, extra ...string)
	formatMessage(w io.Writer, format string, a ...interface{})
	finish(w io.Writer)
}

type target struct {
	writer io.WriteCloser
	formatter
}

func mkTarget(spec string) (target, error) {
	var t target
	for i, part := range strings.Split(spec, ",") {
		if i == 0 {
			u, err := url.Parse(part)
			if err != nil {
				u = &url.URL{Scheme: "file", Path: part}
			}
			if u.Scheme == "" {
				u.Scheme = "file"
			}
			switch {
			case u.Scheme == "file":
				t.writer = &fileWriter{path: u.Path}
			default:
				return target{}, fmt.Errorf("unrecognized scheme '%s'", u.Scheme)
			}
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 1 {
			kv = append(kv, "")
		}
		if kv[0] == "format" {
			switch kv[1] {
			case "plain":
				t.formatter = &formatterPlain{}
			case "tsjson":
				t.formatter = &formatterTSJSON{}
			default:
				return target{}, fmt.Errorf("unrecognized format %s", kv[1])
			}
		}
	}
	if t.formatter == nil {
		t.formatter = &formatterPlain{}
	}
	return t, nil
}
