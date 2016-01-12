package main

import (
	"postlog-sa/filter"
	"testing"
)

func init() {
	log.SetLevel(0)
}

func TestApplyMailThreadItem(t *testing.T) {
	var (
		m = []string{
			`Nov 22 03:47:02 mx postfix/smtpd[9032]: connect from catv-89-135-152-48.catv.broadband.hu[89.135.152.48]`,
			`Nov 22 03:47:02 mx postfix/smtpd[9032]: D549FB08A08B: client=catv-89-135-152-48.catv.broadband.hu[89.135.152.48]`,
			`Nov 22 03:47:04 mx postfix/cleanup[9076]: D549FB08A08B: message-id=4ea6228aa481371f55ed30bb51408481@mail.ru`,
			`Nov 22 03:47:04 mx postfix/qmgr[8015]: D549FB08A08B: from=<cvetkova09055@mail.ru>, size=1755, nrcpt=1 (queue active)`,
			`Nov 22 03:47:04 mx spamd[1279]: spamd: connection from localhost [127.0.0.1] at port 38559`,
			`Nov 22 03:47:04 mx spamd[1279]: spamd: setuid to spamd succeeded`,
			`Nov 22 03:47:04 mx spamd[1279]: spamd: processing message <4ea6228aa481371f55ed30bb51408481@mail.ru> for spamd:1003`,
			`Nov 22 03:47:04 mx postfix/smtpd[9032]: disconnect from catv-89-135-152-48.catv.broadband.hu[89.135.152.48]`,
			`Nov 22 03:47:13 mx spamd[1279]: spamd: identified spam (8.5/5.0) for spamd:1003 in 9.8 seconds, 1801 bytes.`,
			`Nov 22 03:47:13 mx spamd[1279]: spamd: result: Y 8 - FREEMAIL_ENVFROM_END_DIGIT,FREEMAIL_FROM,INVALID_MSGID,MISSING_DATE,RCVD_IN_BRBL_LASTEXT,RCVD_IN_PSBL,RDNS_DYNAMIC,SPF_SOFTFAIL,TO_NO_BRKTS_DYNIP,T_TO_NO_BRKTS_FREEMAIL scantime=9.8,size=1801,user=spamd,uid=1003,required_score=5.0,rhost=localhost,raddr=127.0.0.1,rport=38559,mid=<4ea6228aa481371f55ed30bb51408481@mail.ru>,autolearn=disabled,shortcircuit=no`,
			`Nov 22 03:47:13 mx dovecot: lda(inovv@domain.com): msgid=4ea6228aa481371f55ed30bb51408481@mail.ru: saved mail to INBOX`,
			`Nov 22 03:47:13 mx postfix/pipe[9078]: D549FB08A08B: to=<sinovv@domain.com>, relay=dovecot, delay=11, delays=1.3/0.01/0/9.8, dsn=2.0.0, status=sent (delivered via dovecot service)`,
			`Nov 22 03:47:13 mx postfix/qmgr[8015]: D549FB08A08B: removed`,
		}

		s    *filter.Storage
		err  error
		iter uint
	)

	var fn = func(item filter.ThreadFace, args ...interface{}) error {
		if v := item.GetMessageId(); v != "4ea6228aa481371f55ed30bb51408481@mail.ru" {
			t.Errorf("Expected message id=%s, but got %s", "4ea6228aa481371f55ed30bb51408481@mail.ru", v)
		}

		if v := item.GetSpamScore(); v < 1 {
			t.Errorf("Expected spam score more than 0, but got %d", v)
		}

		iter++

		return nil
	}

	s = filter.NewStorage()
	s.SetThreadDoneCb(fn)

	for _, l := range m {
		if err = parseLine(s, l); err != nil {
			t.Errorf("Unexpected error: %s at `%s`", err.Error(), l)
		}
	}

	if iter != 1 {
		t.Errorf("Expected callback once, but got %d", iter)
	}

	if v := s.Len(); v != 0 {
		t.Errorf("Expecte storage length 0, but got %d", v)
	}
}

func TestSpamStatWithDefferedMail(t *testing.T) {
	var (
		m = []string{
			`Dec 27 02:38:54 mx postfix/qmgr[22753]: C93CEB08A046: from=<ina@somedomain.com>, size=1844, nrcpt=1 (queue active)`,
			`Dec 27 02:38:54 mx postfix/smtp[24234]: connect to arobor.ru[178.210.73.19]:25: Connection timed out`,
			`Dec 27 02:39:24 mx postfix/smtp[15816]: C93CEB08A046: to=<order@arobor.ru>, relay=none, delay=307136, delays=307106/0.02/30/0, dsn=4.4.1, status=deferred (connect to arobor.ru[178.210.73.19]:25: Connection timed out)`,
			`Dec 27 03:48:54 mx postfix/qmgr[22753]: C93CEB08A046: from=<ina@somedomain.com>, size=1844, nrcpt=1 (queue active)`,
			`Dec 27 03:48:54 mx postfix/smtp[24234]: connect to arobor.ru[178.210.73.19]:25: Connection timed out`,
			`Dec 27 03:49:24 mx postfix/smtp[15816]: C93CEB08A046: to=<order@arobor.ru>, relay=none, delay=307136, delays=307106/0.02/30/0, dsn=4.4.1, status=deferred (connect to arobor.ru[178.210.73.19]:25: Connection timed out)`,
			`Dec 28 13:39:28 mx postfix/qmgr[22753]: C93CEB08A046: from=<ina@somedomain.com>, status=expired, returned to sender`,
			`Dec 28 13:39:28 mx postfix/cleanup[19210]: B1B894562003: message-id=<20151228103928.B1B894562003@mx.somedomain.com>`,
			`Dec 28 13:39:28 mx postfix/bounce[19155]: C93CEB08A046: sender non-delivery notification: B1B894562003`,
			`Dec 28 13:39:28 mx postfix/qmgr[22753]: B1B894562003: from=<>, size=3773, nrcpt=1 (queue active)`,
			`Dec 28 13:39:28 mx postfix/qmgr[22753]: C93CEB08A046: removed`,
		}

		s    *filter.Storage
		err  error
		iter uint
	)

	var fn = func(item filter.ThreadFace, args ...interface{}) error {
		if sp := item.GetSpamScore(); sp > 0 {
			t.Errorf("Expected spam score 0, but got %d", sp)
		}

		iter++

		return nil
	}

	s = filter.NewStorage()

	s.SetThreadDoneCb(fn)

	for _, l := range m {
		if err = parseLine(s, l); err != nil {
			t.Errorf("Unexpected error: %s at `%s`", err.Error(), l)
		}
	}

	if iter != 1 {
		t.Errorf("Expected callback triger once, but got %d", iter)
	}
}

func TestStorageSetMutlipleSmtpd(t *testing.T) {
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
		err  error
		iter int
		a    uint = 3
	)

	var fn = func(item filter.ThreadFace, args ...interface{}) error {
		if v := item.GetSpamScore(); v != a {
			t.Errorf("Expected spam score %d, but got %d", a, v)
		}

		iter++

		return nil
	}

	s = filter.NewStorage()

	s.SetThreadDoneCb(fn)

	for _, l := range m {
		if err = parseLine(s, l); err != nil {
			t.Errorf("Unexpected error: %s at `%s`", err.Error(), l)
		}
	}

	if iter != 1 {
		t.Fatalf("Expected one callback execution, but got %d", iter)
	}
}
