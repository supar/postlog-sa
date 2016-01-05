package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

var file_name = "sbss_sla.ini_"

func TestConfig_FileNotexists(t *testing.T) {
	var (
		err error
	)

	if _, err = NewConfig("/this/not/valid/File/check"); err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("Expected error 'IsNotExist', but got '%s'", err.Error())
		}
	} else {
		t.Fatal("Expected error on none valid file, but got nothing")
	}
}

func TestConfig_DirInsteadFile(t *testing.T) {
	var (
		dir      string
		err      error
		validErr *regexp.Regexp
	)

	if validErr, err = regexp.Compile("This is directory"); err != nil {
		t.Fatalf("Can't create regexp object, error: %s", err.Error())
	}

	if dir, err = ioutil.TempDir("", ""); err != nil {
		t.Fatalf("Can't get temporary dir, error: %s", err.Error())
	}

	if _, err = NewConfig(dir); err != nil {
		if v := validErr.MatchString(err.Error()); v == false {
			t.Fatalf("Expected err 'This is directory __, please, set right configuration file', but got %s", err.Error())
		}
	} else {
		t.Fatal("Expected error message, but got nothing")
	}
}

func TestConfig_DBRead(t *testing.T) {
	var (
		file     *os.File
		err      error
		ini_mock string
		cfg_json string
		cfg      *Config
	)

	ini_mock = `
# Some comment
[db]
name = dbname
host = local.dbhost
port = 3306
user = dbuser
pass = dbpassword
charset = utf8
`
	cfg_json = `{"User":"dbuser","Password":"dbpassword","Host":"local.dbhost","Port":3306,"Name":"dbname","Charset":"utf8","Location":""}`

	if file, err = ioutil.TempFile("", file_name); err != nil {
		t.Fatalf("Expected temporary file, but got error: %s", err.Error())
	}

	defer os.Remove(file.Name())

	if _, err = file.WriteString(ini_mock); err != nil {
		t.Fatalf("Can't write file content. Error: %s", err.Error())
	}

	file.Close()

	if cfg, err = NewConfig(file.Name()); err != nil {
		t.Fatalf("Expected to open file %s, but got error: %s", file.Name(), err.Error())
	}

	if v := cfg.GetJson("DB"); v != cfg_json {
		t.Logf("Configuration: %s", ini_mock)
		t.Fatalf("Excpeted db data '%s', but got '%s'", cfg_json, v)
	}
}

func TestConfig_LogRead(t *testing.T) {
	var (
		file     *os.File
		file_log *os.File
		err      error
		ini_mock string
		cfg_json string
		cfg      *Config
	)

	ini_mock = `
# Some comment
[log]
file = %s
level = 6
`

	cfg_json = `{"level":6,"filename":"%s"}`

	if file, err = ioutil.TempFile("", file_name); err != nil {
		t.Fatalf("Expected temporary file, but got error: %s", err.Error())
	}

	if file_log, err = ioutil.TempFile("", ""); err != nil {
		t.Fatalf("Expected temporary file, but got error: %s", err.Error())
	}

	defer os.Remove(file.Name())
	defer os.Remove(file_log.Name())

	ini_mock = fmt.Sprintf(ini_mock, file_log.Name())
	cfg_json = fmt.Sprintf(cfg_json, file_log.Name())

	if _, err = file.WriteString(ini_mock); err != nil {
		t.Fatalf("Can't write file content. Error: %s", err.Error())
	}

	file.Close()
	file_log.Close()

	if cfg, err = NewConfig(file.Name()); err != nil {
		t.Fatalf("Expected to open file %s, but got error: %s", file.Name(), err.Error())
	}

	if v := cfg.GetJson("Log"); v != cfg_json {
		t.Logf("Configuration: %s", ini_mock)
		t.Fatalf("Excpeted log data '%s', but got '%s'", cfg_json, v)
	}
}

func TestConfig_LogReadIsDir(t *testing.T) {
	var (
		file     *os.File
		dir      string
		err      error
		ini_mock string
		cfg_json string
		cfg      *Config
	)

	ini_mock = `
# Some comment
[log]
file = %s
level = 6
`
	cfg_json = `{"level":0,"filename":""}`

	if file, err = ioutil.TempFile("", file_name); err != nil {
		t.Fatalf("Expected temporary file, but got error: %s", err.Error())
	}

	if dir, err = ioutil.TempDir("", ""); err != nil {
		t.Fatalf("Expected temporary dir, but got error: %s", err.Error())
	}

	defer os.Remove(file.Name())

	ini_mock = fmt.Sprintf(ini_mock, dir)

	if _, err = file.WriteString(ini_mock); err != nil {
		t.Fatalf("Can't write file content. Error: %s", err.Error())
	}

	file.Close()

	if cfg, err = NewConfig(file.Name()); err != nil {
		t.Fatalf("Expected to open file %s, but got error: %s", file.Name(), err.Error())
	}

	if v := cfg.GetJson("Log"); v != cfg_json {
		t.Logf("Configuration: %s", ini_mock)
		t.Fatalf("Excpeted log data '%s', but got '%s'", cfg_json, v)
	}
}

func TestConfig_LogToJSON(t *testing.T) {
	var (
		file     *os.File
		err      error
		ini_mock string
		cfg_json string
		cfg      *Config
	)

	ini_mock = `
# Some log comment
[tail]
file = 

# Some DB comment
[db]
name = 
host = 
port = 
user = 
pass = 
charset = 

# Some log comment
[log]
file = 
level = 
`
	cfg_json = fmt.Sprintf(
		`{"Tail":%s,"DB":%s,"Log":%s,"Console":%s}`,
		`{"File":""}`,
		`{"User":"","Password":"","Host":"","Port":0,"Name":"","Charset":"","Location":""}`,
		`{"level":0,"filename":""}`,
		`{"level":0}`,
	)

	if file, err = ioutil.TempFile("", file_name); err != nil {
		t.Fatalf("Expected temporary file, but got error: %s", err.Error())
	}

	defer os.Remove(file.Name())

	if _, err = file.WriteString(ini_mock); err != nil {
		t.Fatalf("Can't write file content. Error: %s", err.Error())
	}

	file.Close()

	if cfg, err = NewConfig(file.Name()); err != nil {
		t.Fatalf("Expected to open file %s, but got error: %s", file.Name(), err.Error())
	}

	if v := cfg.GetJson(""); v != cfg_json {
		t.Logf("Configuration: %s", ini_mock)
		t.Fatalf("Excpeted log data '%s', but got '%s'", cfg_json, v)
	}
}

func TestConfig_LogFilNotExists(t *testing.T) {
	var (
		file     *os.File
		dir      string
		err      error
		ini_mock string
		log_file string
		cfg_json string
		cfg      *Config
	)

	ini_mock = `
# Some comment
[log]
file = %s%s
`
	cfg_json = `{"level":0,"filename":"%s%s"}`

	log_file = "service_log_file.log"

	if file, err = ioutil.TempFile("", file_name); err != nil {
		t.Fatalf("Expected temporary file, but got error: %s", err.Error())
	}

	if dir, err = ioutil.TempDir("", ""); err != nil {
		t.Fatalf("Expected temporary dir, but got error: %s", err.Error())
	}

	defer os.Remove(file.Name())
	defer os.Remove(dir + log_file)

	ini_mock = fmt.Sprintf(ini_mock, dir, log_file)
	cfg_json = fmt.Sprintf(cfg_json, dir, log_file)

	if _, err = file.WriteString(ini_mock); err != nil {
		t.Fatalf("Can't write file content. Error: %s", err.Error())
	}

	file.Close()

	if cfg, err = NewConfig(file.Name()); err != nil {
		t.Fatalf("Expected to open file %s, but got error: %s", file.Name(), err.Error())
	}

	if v := cfg.GetJson("Log"); v != cfg_json {
		t.Logf("Configuration: %s", ini_mock)
		t.Fatalf("Excpeted log data '%s', but got '%s'", cfg_json, v)
	}
}
