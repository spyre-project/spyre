package log

import (
	"errors"
	"strings"
)

type level uint8

var levelStrings = []string{"trace", "debug", "info", "notice", "warn", "error", "quiet"}

func (l level) String() string {
	if s := levelStrings[int(l)]; s != "" {
		return s
	}
	return "unknown"
}

func (l *level) Set(v string) error {
	for i, s := range levelStrings {
		if strings.ToLower(v) == s {
			*l = level(i)
			Init()
			return nil
		}
	}
	return errors.New("unknown level")
}

func (l level) Type() string { return "loglevel" }

const (
	levelTrace level = iota
	levelDebug
	levelInfo
	levelNotice
	levelWarn
	levelError
	levelQuiet
)
