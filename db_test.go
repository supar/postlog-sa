package main

import (
	"database/sql"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"postlog-sa/filter"
	"reflect"
	"regexp"
	"testing"
)

// Helper to initialize sqlmock
func InitDBMock(t *testing.T) (db *sql.DB, mock sqlmock.Sqlmock) {
	var (
		err error
	)

	// open database stub
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	return
}

func TestNewDBUrl_UserRequired(t *testing.T) {
	var (
		cfg *Config
		err error
		rgx *regexp.Regexp
	)

	cfg = &Config{}

	if rgx, err = regexp.Compile("user name is required"); err != nil {
		t.Fatalf("Can't create regexp, error: %s", err.Error())
	}

	if _, err = NewDBUrl(cfg); err == nil {
		t.Fatal("Expected error message on empty user name")
	} else {
		if !rgx.MatchString(err.Error()) {
			t.Fatalf("Expected error message like user name is required, but got '%s'", err.Error())
		}
	}
}

func TestNewDBUrl_SBNameRequired(t *testing.T) {
	var (
		cfg *Config
		err error
		rgx *regexp.Regexp
	)

	cfg = &Config{}
	cfg.DB.User = "someuser"

	if rgx, err = regexp.Compile("DB name is required"); err != nil {
		t.Fatalf("Can't create regexp, error: %s", err.Error())
	}

	if _, err = NewDBUrl(cfg); err == nil {
		t.Fatal("Expected error message on empty user name")
	} else {
		if !rgx.MatchString(err.Error()) {
			t.Fatalf("Expected error message like DB name is required, but got '%s'", err.Error())
		}
	}
}

func TestNewDBUrl_Default(t *testing.T) {
	var (
		u      string
		u_mock string
		cfg    *Config
		err    error
	)

	cfg = &Config{}
	cfg.DB.User = "someuser"
	cfg.DB.Name = "somedb"
	cfg.DB.Password = "somepasswd"
	cfg.DB.Host = "localhost"

	u_mock = "someuser:somepasswd@tcp(localhost:3306)/somedb?charset=utf8&loc=Local&parseTime=true"

	if u, err = NewDBUrl(cfg); err != nil {
		t.Fatal("Unexpected error: %s", err.Error())
	}

	if u != u_mock {
		t.Fatalf("Expected url '%s', but got '%s'", u_mock, u)
	}
}

func TestNewDBUrl_TimeLocationMoscow(t *testing.T) {
	var (
		u      string
		u_mock string
		cfg    *Config
		err    error
	)

	cfg = &Config{}
	cfg.DB.User = "someuser"
	cfg.DB.Name = "somedb"
	cfg.DB.Password = "somepasswd"
	cfg.DB.Host = "localhost"
	cfg.DB.Location = "Europe/Moscow"

	u_mock = "someuser:somepasswd@tcp(localhost:3306)/somedb?charset=utf8&loc=Europe%2FMoscow&parseTime=true"

	if u, err = NewDBUrl(cfg); err != nil {
		t.Fatal("Unexpected error: %s", err.Error())
	}

	if u != u_mock {
		t.Fatalf("Expected url '%s', but got '%s'", u_mock, u)
	}
}

func TestParseStatmentQuery(t *testing.T) {
	var (
		m = []string{
			"INSERT INTO `table`(`field`) VALUES(?c)",
			"INSERT INTO `table`(`field`) VALUES(?f)",
			"INSERT INTO `table`(`field`) VALUES(?i)",
			"INSERT INTO `table`(`field`) VALUES(?m)",
			"INSERT INTO `table`(`field`) VALUES(?s)",
			"INSERT INTO `table`(`field`) VALUES(?t)",
		}

		r []rune
	)

	for i, q := range m {
		buf, runes, err := parseStmtQuery(q)

		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if buf == nil || runes == nil {
			t.Errorf("Unexpected nil values [%v] [%v]", buf, runes)
		}

		if v := buf.String(); v != "INSERT INTO `table`(`field`) VALUES(?)" {
			t.Errorf("Expected [%s], but got [%s]", "INSERT INTO `table`(`field`) VALUES(?t)", v)
		}

		switch i {
		case 0:
			r = []rune{99}
		case 1:
			r = []rune{102}
		case 2:
			r = []rune{105}
		case 3:
			r = []rune{109}
		case 4:
			r = []rune{115}
		case 5:
			r = []rune{116}

		default:
			r = make([]rune, 0)
		}

		if !reflect.DeepEqual(runes, r) {
			t.Errorf("Expected runes %v, but got %v", r, runes)
		}
	}
}

func TestParseStatmentQueryMixValues(t *testing.T) {
	var (
		m = []string{
			"INSERT INTO `table`(`a`, `b`, `c`, `d`, `e`) VALUES(?c, ?t, ?t, ?f, ?i)",
		}

		r []rune
	)

	for i, q := range m {
		buf, runes, err := parseStmtQuery(q)

		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if buf == nil || runes == nil {
			t.Errorf("Unexpected nil values [%v] [%v]", buf, runes)
		}

		switch i {
		case 0:
			r = []rune{99, 116, 116, 102, 105}

		default:
			r = make([]rune, 0)
		}

		if !reflect.DeepEqual(runes, r) {
			t.Errorf("Expected runes %v, but got %v", r, runes)
		}
	}
}

func TestStatmentCall(t *testing.T) {
	var (
		m = []string{
			`Dec  4 10:33:23 mx postfix/smtpd[14247]: connect from unknown[1.7.1.1]`,
			`Dec  4 10:33:24 mx postfix/smtpd[14247]: 5247C4562029: client=unknown[1.7.1.1]`,
			`Dec  4 10:33:24 mx postfix/cleanup[14676]: 5247C4562029: message-id=<659691365ACB4CBBA5AD2A0C2E2639E2@ip-1-7-1-1.bb.net.net>`,
			`Dec  4 10:33:24 mx postfix/qmgr[22753]: 5247C4562029: from=<simonova@yahoo.com>, size=18194, nrcpt=8 (queue active)`,
			`Dec  4 10:33:24 mx postfix/smtpd[14247]: disconnect from unknown[176.77.17.151]`,
			`Dec  4 10:33:24 mx postfix/smtpd[14680]: connect from localhost[127.0.0.1]`,
			`Dec  4 10:33:24 mx postfix/smtpd[14680]: F3C59456202A: client=localhost[127.0.0.1]`,
			`Dec  4 10:33:25 mx postfix/cleanup[14676]: F3C59456202A: message-id=<659691365ACB4CBBA5AD2A0C2E2639E2@ip-1-7-1-1.bb.net.net>`,
			`Dec  4 10:33:25 mx postfix/smtpd[14680]: disconnect from localhost[127.0.0.1]`,
			`Dec  4 10:33:25 mx postfix/qmgr[22753]: F3C59456202A: from=<simonovaa@yahoo.com>, size=19017, nrcpt=8 (queue active)`,
			`Dec  4 10:33:25 mx amavis[13901]: (13901-15) Passed SPAMMY {RelayedTaggedInbound}, [1.7.1.1]:4578 [1.7.1.1] <simonova@yahoo.com> -> <lik@some.net>,<ochki@some.net>,<tik@some.net>, Queue-ID: 5247C4562029, Message-ID: <659691365ACB4CBBA5AD2A0C2E2639E2@ip-1-7-1-1.bb.net.net>, mail_id: 3JiB02xP9nql, Hits: 18.715, size: 18198, queued_as: F3C59456202A, 659 ms`,
			`Dec  4 10:33:25 mx postfix/smtp[14678]: 5247C4562029: to=<lik@some.net>, relay=127.0.0.1[127.0.0.1]:10024, delay=1.3, delays=0.67/0/0/0.66, dsn=2.0.0, status=sent (250 2.0.0 from MTA(smtp:[127.0.0.1]:10025): 250 2.0.0 Ok: queued as F3C59456202A)`,
			`Dec  4 10:33:25 mx postfix/smtp[14678]: 5247C4562029: to=<ochki@some.net>, relay=127.0.0.1[127.0.0.1]:10024, delay=1.3, delays=0.67/0/0/0.66, dsn=2.0.0, status=sent (250 2.0.0 from MTA(smtp:[127.0.0.1]:10025): 250 2.0.0 Ok: queued as F3C59456202A)`,
			`Dec  4 10:33:25 mx postfix/smtp[14678]: 5247C4562029: to=<tik@some.net>, relay=127.0.0.1[127.0.0.1]:10024, delay=1.3, delays=0.67/0/0/0.66, dsn=2.0.0, status=sent (250 2.0.0 from MTA(smtp:[127.0.0.1]:10025): 250 2.0.0 Ok: queued as F3C59456202A)`,
			`Dec  4 10:33:25 mx postfix/qmgr[22753]: 5247C4562029: removed`,
			`Dec  4 10:33:25 mx dovecot: lda(lik@some.net): msgid=<659691365ACB4CBBA5AD2A0C2E2639E2@ip-1-7-1-1.bb.net.net>: saved mail to INBOX`,
			`Dec  4 10:33:25 mx postfix/pipe[14682]: F3C59456202A: to=<lik@some.net>, relay=dovecot, delay=0.09, delays=0.01/0/0/0.08, dsn=2.0.0, status=sent (delivered via dovecot service)`,
			`Dec  4 10:33:25 mx dovecot: lda(ochki@some.net): msgid=<659691365ACB4CBBA5AD2A0C2E2639E2@ip-1-7-1-1.bb.net.net>: saved mail to INBOX`,
			`Dec  4 10:33:25 mx postfix/pipe[14689]: F3C59456202A: to=<ochki@some.net>, relay=dovecot, delay=0.04, delays=0.01/0.01/0/0.02, dsn=2.0.0, status=sent (delivered via dovecot service)`,
			`Dec  4 10:33:25 mx postfix/qmgr[22753]: F3C59456202A: removed`,
		}

		s    *filter.Storage
		stmt *StmtMap
		err  error
	)

	db, mock := InitDBMock(t)
	mock.ExpectPrepare("INSERT").
		ExpectExec().
		WithArgs("simonova@yahoo.com", "0000-12-04 10:33:24", "1.7.1.1").
		WillReturnResult(sqlmock.NewResult(1, 0))

	stmt, err = NewStmt(db, "INSERT INTO `table`(`a`) VALUES(?f, ?t, ?c)")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	var fn = func(item filter.ThreadFace, args ...interface{}) (err error) {
		if item.GetSpamScore() == 0 {
			return
		}
		if err = stmt.Call(item); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		return
	}

	s = filter.NewStorage()

	s.SetThreadDoneCb(fn)

	for _, l := range m {
		if err = parseLine(s, l); err != nil {
			t.Errorf("Unexpected error: %s at `%s`", err.Error(), l)
		}
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf(err.Error())
	}
}
