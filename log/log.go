package log

import (
	"fmt"
	stdlog "log"
)

var GlobalLevel level = levelInfo

var initialized bool

type msg struct {
	lvl     level
	message string
}

var backlog []msg

func emit(lvl level, message string) {
	if !initialized {
		backlog = append(backlog, msg{lvl, message})
		return
	}
	if GlobalLevel > lvl {
		return
	}
	stdlog.Print(message)
}

func Init() {
	if !initialized {
		initialized = true
		for _, item := range backlog {
			emit(item.lvl, item.message)
		}
		backlog = nil
	}
}

func Tracef(f string, v ...interface{}) { emit(levelTrace, fmt.Sprintf(f, v...)) }

func Trace(v ...interface{}) { emit(levelTrace, fmt.Sprint(v...)) }

func Debugf(f string, v ...interface{}) { emit(levelDebug, fmt.Sprintf(f, v...)) }

func Debug(v ...interface{}) { emit(levelDebug, fmt.Sprint(v...)) }

func Infof(f string, v ...interface{}) { emit(levelInfo, fmt.Sprintf(f, v...)) }

func Info(v ...interface{}) { emit(levelInfo, fmt.Sprint(v...)) }

func Noticef(f string, v ...interface{}) { emit(levelNotice, fmt.Sprintf(f, v...)) }

func Notice(v ...interface{}) { emit(levelNotice, fmt.Sprint(v...)) }

func Warnf(f string, v ...interface{}) { emit(levelWarn, fmt.Sprintf(f, v...)) }

func Warn(v ...interface{}) { emit(levelWarn, fmt.Sprint(v...)) }

func Errorf(f string, v ...interface{}) { emit(levelError, fmt.Sprintf(f, v...)) }

func Error(v ...interface{}) { emit(levelError, fmt.Sprint(v...)) }
