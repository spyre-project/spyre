package config

import (
	"bytes"
	"encoding/csv"
	"strings"
)

// stringSlice is a simpler version of the type backing
// pflag.StringSlice etc. whose Set method has "append" semantics.
type simpleStringSlice []string

func (s *simpleStringSlice) Set(val string) (err error) {
	r := csv.NewReader(strings.NewReader(val))
	r.Comma = ';'
	*s, err = r.Read()
	return
}

func (s *simpleStringSlice) String() string {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	w.Comma = ';'
	w.Write(*s)
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
}

func (s simpleStringSlice) Type() string { return "" }
