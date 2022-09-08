package logger

import (
	"fmt"

	"github.com/getsentry/raven-go"
	glog "github.com/sirupsen/logrus"
)

type Logger struct {
	verbose, initialized bool
	tags                 map[string]string
}

var (
	defaultLogger = New(false, nil)
)

func New(verbose bool, tags map[string]string) *Logger {
	return &Logger{
		tags:    tags,
		verbose: verbose,
	}
}

func (l *Logger) Info(message string) {
	Info(message)
}

func (l *Logger) Infof(message string) {
	Infof(message)
}

func (l *Logger) Error(err error) {
	Error(err)
}

func (l *Logger) Fatal(err error) {
	Fatal(err)
}

func Info(message string) {
	glog.Info(message)
}

func Infof(format string, v ...interface{}) {
	Info(fmt.Sprintf(format, v...))
}

func Error(err error) {
	glog.Error(err)
	go raven.CaptureError(err, defaultLogger.tags)
}

func Fatal(err error) {
	raven.CaptureErrorAndWait(err, nil)
	glog.Fatal(err)
}
