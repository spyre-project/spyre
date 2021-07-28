package config

import (
	"bytes"
	"encoding/csv"
	"strings"
)

// StringSlice is a simpler version of the type backing
// pflag.StringSlice etc. whose Set method has "append" semantics.
type StringSlice []string

func (s *StringSlice) Set(val string) (err error) {
	r := csv.NewReader(strings.NewReader(val))
	r.Comma = ';'
	*s, err = r.Read()
	return
}

func (s *StringSlice) String() string {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	w.Comma = ';'
	w.Write(*s)
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
}

func (s StringSlice) Type() string { return "" }
