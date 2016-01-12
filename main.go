package main

import (
	"flag"
	"github.com/hpcloud/tail"
	"postlog-sa/filter"
	"time"
)

func init() {
	log = NewLogger(10000)
	// Set console log as default
	log.SetLogger("console", `{"level":7}`)
}

func main() {
	var (
		err error
		tl  *tail.Tail
		st  *filter.Storage
		sm  *StmtMap
	)
	defer log.Close()

	// Read flags
	flag.Parse()

	// Read configuration
	if Cfg, err = NewConfig(CONFIGFILE); err != nil {
		log.Critical(err.Error())
	}

	// Reconfigure console log
	log.DelLogger("console")
	switch CONSOLELOG {
	case 0:

	default:
		// Rewrite console log adapter if log level is not 7
		log.SetLogger("console", Cfg.GetJson("Console"))
	}

	// Write log to file if there is settings
	if Cfg.Log.Level > 0 {
		log.SetLogger("file", Cfg.GetJson("Log"))

	}

	// Create databse connection
	if src, src_err := NewDBUrl(Cfg); src_err != nil {
		log.Error(src_err.Error())
	} else {
		db, err = OpenDB(src)
		if err == nil {
			Cfg.DB.Ok = true

			// Close DB connection on main function finish
			defer db.Close()

			if sm, err = NewStmt(db, Cfg.SQL.Query); err == nil {
				Cfg.SQL.Ok = true

				defer sm.stmt.Close()
			}
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	// Send greeting
	greeting(log)

	tl, err = tail.TailFile(Cfg.Tail.File, tail.Config{
		Follow: true,
		Logger: log,
	})

	if err != nil {
		log.Critical(err.Error())
	}

	// Create storage
	st = filter.NewStorage()
	// Create callback
	st.SetThreadDoneCb(threadComplete)

	for line := range tl.Lines {
		log.Debug("Parsing:{%s}", line.Text)

		if err = parseLine(st, line.Text, sm, Cfg); err != nil {
			log.Error(err.Error())
		}
	}
}

// Agregate log entries to object with full information to analyze mail
func parseLine(store *filter.Storage, line string, args ...interface{}) (err error) {
	var (
		mi *filter.MailThread
		sp *filter.Spam
	)

	mi, err = filter.NewMailThread(line)
	if err == nil {
		// Write to storage
		store.Set(mi)
	} else {
		if err != filter.ErrorStrFormatNotSupported {
			return err
		}

		err = nil
	}

	if mi == nil {
		if sp, err = filter.NewSpam(line); err != nil {
			return err
		}

		if sp != nil {
			if err = store.SetSpamStat(sp); err == filter.ErrorUnknownSpamItem {
				log.Warn("Can not idendify mail thread for %v", sp)
			}
		}
	}

	if mi != nil {
		store.ThreadDone(mi, args...)
	}

	return nil
}

// Greeting
func greeting(l *Log) {
	l.Info("Service %s started (Version: %s, build date: %s)", NAME, VERSION, BUILDDATE)
}

func threadComplete(item filter.ThreadFace, args ...interface{}) (err error) {
	var (
		sm *StmtMap
		cf *Config
	)

	if item.GetSpamScore() > 0 {
		log.Info(
			"ID: %s, at: %s, from: %s, IP: %s, score: %d",
			item.GetId(),
			item.GetTime().Format(time.Stamp),
			item.GetFrom(),
			item.GetFromIp(),
			item.GetSpamScore(),
		)

		// Dereference arguments
		for _, a := range args {
			switch a.(type) {
			case *StmtMap:
				sm = a.(*StmtMap)
			case *Config:
				cf = a.(*Config)
			}
		}
		if cf != nil && sm != nil && cf.CanSql() {
			err = sm.Call(item)

			if err != nil {
				log.Error(err.Error())
			}
		}
	}

	return
}
