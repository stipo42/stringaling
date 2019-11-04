package util

import (
	"fmt"
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
