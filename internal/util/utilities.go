package util

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

var DEBUG = false

func Debug(s string, args ...interface{}) {
	if DEBUG {
		Log("DEBUG", s, args...)
	}
}
func Info(s string, args ...interface{}) {
	Log("INFO ", s, args...)
}
func Error(s string, args ...interface{}) {
	Log("ERROR", s, args...)
}

func Log(level string, msg string, args ...interface{}) {
	o := msg
	if len(args) > 0 {
		o = fmt.Sprintf(msg, args...)
	}
	fmt.Println(fmt.Sprintf("*%s* %s", level, o))
}

func HumanReadable(ns int64) string {
	var d int64
	var h int64
	var m int64
	var s int64
	ms := ns / 1000000
	if ms >= 0 {
		ns = ns % 1000000
	}
	if ms >= 1000 {
		s = ms / 1000
		ms = ms % 1000
		if s >= 60 {
			m = s / 60.0
			s = s % 60
			if m >= 60 {
				h = m / 60
				m = m % 60
				if h >= 24 {
					d = d / 24
					h = h % 24
				}
			}
		}
	}
	var sb strings.Builder
	sb.WriteString(unitFormatter(d, "day"))
	sb.WriteString(unitFormatter(h, "hour"))
	sb.WriteString(unitFormatter(m, "minute"))
	sb.WriteString(unitFormatter(s, "second"))
	sb.WriteString(unitFormatter(ms, "millisecond"))
	sb.WriteString(unitFormatter(ns, "nanosecond"))
	out := sb.String()
	if out == "" {
		out = "no time"
	}
	return out
}

func unitFormatter(v int64, unit string) string {
	var s string
	if v > 0 {
		s = fmt.Sprintf(" %d %s", v, unit)
	}
	if v > 1 {
		s += "s"
	}
	return s
}

// GetCleanFile gets a file by fileName, deleting it first if it already exists.
func GetCleanFile(fileName string) (file *os.File, err error) {
	_, err = os.Stat(fileName)
	if err == nil {
		Debug("%s exists, deleting it", fileName)
		rerr := os.Remove(fileName)
		if rerr != nil {
			Debug("cannot remove %s: %s", fileName, rerr)
		} else {
			err = errors.New("ok")
		}
	}
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			Debug("error creating file %s: %s", fileName, err)
		} else {
			if runtime.GOOS != "windows" {
				err = file.Chmod(0777)
			}
		}
	}
	return
}
