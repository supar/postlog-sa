package main

import (
	"fmt"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"net"
	"net/url"
)

var (
	db *sql.DB
)

/**
 * Create connection url
 */
func NewDBUrl(cfg *Config) (source string, err error) {
	var (
		u url.Values
	)

	// Database user is required to connet to
	if source = StrEmpty(cfg.DB.User, ""); source == "" {
		err = fmt.Errorf("DB user name is required")
		return
	}
	// Add password value
	if v := StrEmpty(cfg.DB.Password, ""); v != "" {
		source += ":" + v
	}

	// Add connection way
	source += "@"
	if cfg.DB.Host != "" || cfg.DB.Port > 0 {
		if cfg.DB.Port == 0 {
			cfg.DB.Port = 3306
		}

		source += "tcp("
		source += net.JoinHostPort(
			StrEmpty(cfg.DB.Host, "127.0.0.1"),
			StrEmpty(Int2Str(cfg.DB.Port), "3306"),
		)
		source += ")"
	}

	// Add database name
	if v := StrEmpty(cfg.DB.Name, ""); v != "" {
		source += "/" + v
	} else {
		err = fmt.Errorf("DB name is required")
		return
	}

	// Add source url paramerters
	u = url.Values{}
	// Add charset parameter
	u.Add("charset", StrEmpty(cfg.DB.Charset, "utf8"))
	// Add time parameters
	u.Add("parseTime", "true")
	u.Add("loc", StrEmpty(cfg.DB.Location, "Local"))

	if v := u.Encode(); v != "" {
		source += "?" + v
	}

	return
}

/**
 * Create database connection on given url
 */
func OpenDB(url string) (db *sql.DB, err error) {
	if url == "" {
		err = fmt.Errorf("Empty connection string")

		return
	}

	db, err = sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}

	return
}
