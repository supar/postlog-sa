package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"reflect"
)

var (
	NAME       = "spam-bug"
	CONFIGFILE = ""
	CONSOLELOG = LevelDebug

	VERSION   string
	BUILDDATE string

	Cfg *Config
)

type Config struct {
	Tail struct {
		File string `ini:"file"`
	} `ini:"tail"`

	DB struct {
		User     string `ini:"user"`
		Password string `ini:"pass"`
		Host     string `ini:"host"`
		Port     int    `ini:"port"`
		Name     string `ini:"name"`
		Charset  string `ini:"charset"`
		Location string `ini:"zone"`
	} `ini:"db"`

	Log struct {
		Level int    `ini:"level" json:"level"`
		File  string `ini:"file" json:"filename"`
	} `ini:"log"`

	Console struct {
		Level int `json:"level"`
	}
}

func init() {
	flag.StringVar(&CONFIGFILE, "C", "/etc/spam-bug/spam-bug.ini", "Configuration file path required")
	flag.IntVar(&CONSOLELOG, "v", 0, "Console verbose level output, default 0 - off, 7 - debug")
}

// Create new configuration
func NewConfig(file string) (c *Config, err error) {
	var (
		f os.FileInfo
		i *ini.File
	)

	c = &Config{}

	if f, err = os.Stat(file); os.IsNotExist(err) {
		return nil, err
	} else {
		if f.IsDir() {
			err = fmt.Errorf("This is directory `%s', please, set right configuration file\n", file)

			return nil, err
		}
	}

	if i, err = ini.Load(file); err != nil {
		return nil, err
	} else {
		if err = i.MapTo(c); err != nil {
			return nil, err
		}
	}

	if c.Log.File != "" {
		if lg, lg_err := os.Stat(c.Log.File); lg_err != nil && os.IsNotExist(lg_err) {
			// TODO: Need check if there possible to create log file
		} else {
			if lg.IsDir() {
				c.Log.File = ""
				c.Log.Level = 0
			}
		}
	}

	switch true {
	case CONSOLELOG <= 0:
		c.Console.Level = 0

	case CONSOLELOG >= 7:
		c.Console.Level = 7

	default:
		c.Console.Level = CONSOLELOG
	}

	return
}

// Convert part of the configuration struct or whole object to json string
func (this *Config) GetJson(s string) (c string) {
	var (
		v reflect.Value
		m []byte
	)

	if s == "" {
		m, _ = json.Marshal(this)
	} else {
		v = reflect.ValueOf(this)
		m, _ = json.Marshal(v.Elem().FieldByName(s).Interface())
	}

	return string(m)
}
