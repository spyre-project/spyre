package log

import (
	"errors"
	"strings"
)

type Level uint8

var levelStrings = []string{"trace", "debug", "info", "notice", "warn", "error", "quiet"}

func (l Level) String() string {
	if s := levelStrings[int(l)]; s != "" {
		return s
	}
	return "unknown"
}

func (l *Level) Set(v string) error {
	for i, s := range levelStrings {
		if strings.ToLower(v) == s {
			*l = Level(i)
			return nil
		}
	}
	return errors.New("unknown level")
}

func (l Level) Type() string { return "loglevel" }

const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelNotice
	LevelWarn
	LevelError
	LevelQuiet
)
