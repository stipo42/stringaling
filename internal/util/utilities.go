package util

import "fmt"

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
