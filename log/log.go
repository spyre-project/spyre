package log

import (
	stdlog "log"
)

var GlobalLevel Level = LevelInfo

func Tracef(f string, v ...interface{}) {
	if GlobalLevel > LevelTrace {
		return
	}
	stdlog.Printf(f, v...)
}

func Trace(v ...interface{}) {
	if GlobalLevel > LevelTrace {
		return
	}
	stdlog.Print(v...)
}

func Debugf(f string, v ...interface{}) {
	if GlobalLevel > LevelDebug {
		return
	}
	stdlog.Printf(f, v...)
}

func Debug(v ...interface{}) {
	if GlobalLevel > LevelDebug {
		return
	}
	stdlog.Print(v...)
}

func Infof(f string, v ...interface{}) {
	if GlobalLevel > LevelInfo {
		return
	}
	stdlog.Printf(f, v...)
}

func Info(v ...interface{}) {
	if GlobalLevel > LevelInfo {
		return
	}
	stdlog.Print(v...)
}

func Noticef(f string, v ...interface{}) {
	if GlobalLevel > LevelNotice {
		return
	}
	stdlog.Printf(f, v...)
}

func Notice(v ...interface{}) {
	if GlobalLevel > LevelNotice {
		return
	}
	stdlog.Print(v...)
}

func Warnf(f string, v ...interface{}) {
	if GlobalLevel > LevelWarn {
		return
	}
	stdlog.Printf(f, v...)
}

func Warn(v ...interface{}) {
	if GlobalLevel > LevelWarn {
		return
	}
	stdlog.Print(v...)
}

func Errorf(f string, v ...interface{}) {
	if GlobalLevel > LevelError {
		return
	}
	stdlog.Printf(f, v...)
}

func Error(v ...interface{}) {
	if GlobalLevel > LevelError {
		return
	}
	stdlog.Print(v...)
}
