package main

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"os"
)

var log *Log

// RFC5424 log message levels.
// 0       Emergency: system is unusable
// 1       Alert: action must be taken immediately
// 2       Critical: critical conditions
// 3       Error: error conditions
// 4       Warning: warning conditions
// 5       Notice: normal but significant condition
// 6       Informational: informational messages
// 7       Debug: debug-level messages
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

type LogIface interface {
	Critical(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})

	Error(format interface{}, v ...interface{})
	Warning(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Notice(format string, v ...interface{})
	Info(format string, v ...interface{})
}

// Wrap external library for the short way to log events.
// This will help to implement other logic
type Log struct {
	// In future its's better to discard chinese code
	*logs.BeeLogger

	// Log level
	Level int
}

// Create log object
func NewLogger(recbuf int64) (logger *Log) {
	logger = &Log{
		BeeLogger: logs.NewLogger(recbuf),
		Level:     LevelDebug,
	}

	// Set default log level
	logger.SetLevel(LevelDebug)

	return
}

// Override to set level value: main class and embeded BeeLog
func (this *Log) SetLevel(lv int) {
	if lv < LevelEmergency || lv > LevelDebug {
		lv = LevelError
	}

	this.Level = lv
	this.BeeLogger.SetLevel(this.Level)
}

func (this *Log) Error(format interface{}, v ...interface{}) {
	switch format.(type) {
	case string:
		this.BeeLogger.Error(format.(string), v...)

	case error:
		this.BeeLogger.Error(format.(error).Error())

	default:
		this.BeeLogger.Error("Unknown Error")
	}
}

func (this *Log) Critical(format string, v ...interface{}) {
	this.BeeLogger.Critical(format, v...)

	this.Die(true)
}

func (this *Log) Die(iserror bool) {
	var code int = -1

	switch iserror {
	case true:
		this.Error("Exit")

	case false:
		this.Info("Exit")
		code = 0
	}

	this.Close()
	os.Exit(code)
}

func (this *Log) Fatal(v ...interface{}) {
	this.Critical(fmt.Sprint(v))
}

func (this *Log) Fatalf(format string, v ...interface{}) {
	this.Critical(format, v)
}

func (this *Log) Fatalln(v ...interface{}) {
	this.Fatal(v)
}

func (this *Log) Panic(v ...interface{}) {
	this.Fatal(v)
}

func (this *Log) Panicf(format string, v ...interface{}) {
	this.Fatal(format, v)
}

func (this *Log) Panicln(v ...interface{}) {
	this.Fatal(v)
}

func (this *Log) Print(v ...interface{}) {
	this.Info(fmt.Sprint(v))
}

func (this *Log) Printf(format string, v ...interface{}) {
	this.Info(format, v)
}

func (this *Log) Println(v ...interface{}) {
	this.Print(v)
}
