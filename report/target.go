package report

import (
	"github.com/spf13/afero"

	"fmt"
	"io"
	"net/url"
	"strings"
)

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
			var u *url.URL
			var err error
			if len(part) >= 2 &&
				('a' <= part[0] && part[0] <= 'z') || ('A' <= part[0] && part[0] <= 'Z') &&
				part[1] == ':' {
				u = &url.URL{Scheme: "file", Path: part}
			} else if u, err = url.Parse(part); err != nil {
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
			case "tsjsonl", "tsjsonlines":
				t.formatter = &formatterTSJSONLines{}
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
