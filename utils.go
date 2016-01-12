package main

import (
	"strconv"
	"strings"
)

// Check if string value not empty and return value
// otherwise return default value
func StrEmpty(v, def string) (s string) {
	if s = strings.Trim(v, " "); s != "" {
		return s
	}

	return def
}

// Convert integer value to the string
func Int2Str(v interface{}) (s string) {
	switch v.(type) {
	case string:
		s = v.(string)

	case int:
		s = strconv.Itoa(v.(int))
	}

	return
}

// Convert string value to the integer
func Str2Int(v interface{}) (i int) {
	switch v.(type) {
	case string:
		i, _ = strconv.Atoi(v.(string))

	case int:
		i = v.(int)
	}

	return
}

// Convert string value to the boolean
func Str2Bool(v interface{}) (t bool) {
	var i = 0

	switch v.(type) {
	case string:
		i, _ = strconv.Atoi(v.(string))

	case int:
		i = v.(int)

	case bool:
		if v.(bool) == true {
			i = 1
		} else {
			i = 0
		}
	}

	if i > 0 {
		t = true
	}

	return
}
