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

	// Send greeting
	greeting(log)

	tl, err = tail.TailFile(Cfg.Tail.File, tail.Config{
		Follow: true,
		Logger: log,
	})

	if err != nil {
		log.Critical("%s", err.Error())
	}

	// Create storage
	st = filter.NewStorage()
	// Create callback
	st.SetThreadDoneCb(threadComplete)

	for line := range tl.Lines {
		log.Debug("Parsing:{%s}", line.Text)

		if err = parseLine(st, line.Text); err != nil {
			log.Error("%s", err.Error())
		}
	}
}


// Agregate log entries to object with full information to analyze mail
func parseLine(store *filter.Storage, line string) (err error) {
	var (
		mi  *filter.MailThread
		sp  *filter.Spam
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
		store.ThreadDone(mi)
	}

	return nil
}

// Greeting
func greeting(l *Log) {
	l.Info("Service %s started (Version: %s, build date: %s)", NAME, VERSION, BUILDDATE)
}

func threadComplete(item filter.ThreadFace) {
	if item.GetSpamScore() > 0 {
		log.Info("ID: %s, at: %s, from: %s, IP: %s, score: %d", item.GetId(), item.GetTime().Format(time.Stamp), item.GetFrom(), item.GetFromIp(), item.GetSpamScore())
	}
}
