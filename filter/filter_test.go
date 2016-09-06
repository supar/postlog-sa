package filter

import (
	"testing"
	"time"
)

func TestIsSmtpd(t *testing.T) {
	var m = []string{
		`Nov 22 01:45:57 mx postfix/smtpd[5910]: A2CAFB08A049: client=unknown[1.1.1.1]`,
		`Nov 22 02:08:42 mx postfix/smtpd[6223]: BDCA3B08A08A: client=mrelay1.hh.ru[001.001.001.001]`,
		`Nov 22 02:24:53 mx postfix/smtpd[6223]: B29AFB08A08A: client=unknown[01.01.001.01], sasl_method=PLAIN, sasl_username=abcd@somedomain.com`,
		`Nov 22 02:24:53 mx postfix/cleanup[6917]: B29AFB08A08A: message-id=<redmine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com>`,
	}

	var c int = 0

	for _, v := range m {
		if ok, _ := IsPostfix(v, []string{"smtpd"}); ok {
			c++
		}
	}

	if c != 3 {
		t.Errorf("Expected %d matches, but got %d", 3, c)
	}
}

func TestIsCleanup(t *testing.T) {
	var m = []string{
		`Nov 22 01:45:57 mx postfix/smtpd[5910]: A2CAFB08A049: client=unknown[1.1.1.1]`,
		`Nov 22 02:24:53 mx postfix/cleanup[6917]: B29AFB08A08A: message-id=<redmine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com>`,
		`Nov 22 09:25:03 mx postfix/pipe[18851]: 844E2B08A079: to=<i@domcom.com>, relay=dovecot, delay=0.14, delays=0.03/0.01/0/0.1, dsn=2.0.0, status=sent (delivered via dovecot service)`,
	}

	var c int = 0

	for _, v := range m {
		if IsCleanup(v) {
			c++
		}
	}

	if c != 1 {
		t.Errorf("Expected %d matches, but got %d", 1, c)
	}
}

func TestGetId(t *testing.T) {
	var m = []string{
		`Nov 22 01:45:57 mx postfix/smtpd[5910]: A2CAFB08A049: client=unknown[1.1.1.1]`,
		`Nov 22 02:08:42 mx postfix/smtpd[6223]: BDCA3B08A08A: client=mrelay1.hh.ru[001.001.001.001]`,
		`Nov 22 02:24:53 mx postfix/cleanup[6917]: B29AFB08A08A: message-id=<redmine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com>`,
		`Nov 22 03:47:04 mx postfix/qmgr[8015]: D549FB08A08B: from=<cvetkova09055@mail.ru>, size=1755, nrcpt=1 (queue active)`,
	}

	for i, v := range m {
		id, err := getId(v)

		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
			continue
		}

		switch i {
		case 0:
			if id != "A2CAFB08A049" {
				t.Errorf("Expected %s value, but got %s", "A2CAFB08A049", id)
			}

		case 1:
			if id != "BDCA3B08A08A" {
				t.Errorf("Expected %s value, but got %s", "BDCA3B08A08A", id)
			}

		case 2:
			if id != "B29AFB08A08A" {
				t.Errorf("Expected %s value, but got %s", "B29AFB08A08A", id)
			}

		case 3:
			if id != "D549FB08A08B" {
				t.Errorf("Expected %s value, but got %s", "D549FB08A08B", id)
			}
		}
	}
}

func TestGetIdEmptyError(t *testing.T) {
	var m = []string{
		`Nov 23 03:28:49 mx postfix/smtpd[18413]: connect from mail.rarmelatio.co.ua[85.25.159.20]`,
		`Nov 23 03:28:49 mx spamd[734]: spamd: connection from localhost [127.0.0.1] at port 56533`,
		`Nov 23 03:28:49 mx spamd[734]: spamd: setuid to spamd succeeded`,
		`Nov 23 03:28:49 mx spamd[734]: spamd: processing message <028f01d1257c$96138890$de47ede3@icvyvnk> for spamd:1003`,
		`Nov 23 03:28:56 mx dovecot: pop3(coretest@some.com): Disconnected: Logged out top=0/0, retr=0/0, del=0/0, size=0`,
	}

	for _, v := range m {
		_, err := getId(v)

		if err == nil {
			t.Errorf("Expected error %s, but got nothing", ErrorStrFormatNotSupported.Error())
			continue
		}

		if err != ErrorStrFormatNotSupported {
			t.Errorf("Expected error %s, but got", ErrorStrFormatNotSupported.Error(), err.Error())
		}
	}
}

func TestGetClient(t *testing.T) {
	var m = []string{
		`Nov 22 01:45:57 mx postfix/smtpd[5910]: A2CAFB08A049: client=unknown[1.1.1.1]`,
		`Nov 22 02:08:42 mx postfix/smtpd[6223]: BDCA3B08A08A: client=mrelay1.hh.ru[001.001.001.001]`,
		`Nov 22 02:24:53 mx postfix/smtpd[6223]: B29AFB08A08A: client=unknown[01.01.001.01], sasl_method=PLAIN, sasl_username=abcs@domain.com`,
		`Nov 22 02:24:53 mx postfix/cleanup[6917]: B29AFB08A08A: message-id=<redmine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com>`,
	}

	for i, v := range m {
		r := getClient(v)

		switch i {
		case 0:
			if r == nil {
				t.Errorf("Not nil value expected")
				continue
			}

			if r.Name != "unknown" {
				t.Errorf("Expected value %s, but got %s", "unknown", r.Name)
				continue
			}

			if r.IP != "1.1.1.1" {
				t.Errorf("Expected value %s, but got %s", "1.1.1.1", r.IP)
				continue
			}

			if x := r.At.Format(time.Stamp); x != "Nov 22 01:45:57" {
				t.Errorf("Expected time %s, but got %s", "Nov 22 01:45:57", x)
				continue
			}

		case 1:
			if r == nil {
				t.Errorf("Not nil value expected")
				continue
			}

			if r.Name != "mrelay1.hh.ru" {
				t.Errorf("Expected value %s, but got %s", "mrelay1.hh.ru", r.Name)
				continue
			}

			if r.IP != "001.001.001.001" {
				t.Errorf("Expected value %s, but got %s", "001.001.001.001", r.IP)
				continue
			}

		case 2:
			if r == nil {
				t.Errorf("Not nil value expected")
				continue
			}

			if r.Name != "unknown" {
				t.Errorf("Expected value %s, but got %s", "unknown", r.Name)
				continue
			}

			if r.IP != "01.01.001.01" {
				t.Errorf("Expected value %s, but got %s", "01.01.001.01", r.IP)
				continue
			}

		case 3:
			if r != nil {
				t.Errorf("Expected nil value, but got %v", r)
			}
		}
	}
}

func TestGetMessageId(t *testing.T) {
	var m = []string{
		`Nov 22 07:59:12 mx postfix/cleanup[15917]: B1B8DB08A08B: message-id=<E1a0Mks-0000Yl-1x@localhost>`,
		`Nov 22 09:25:03 mx postfix/cleanup[18849]: 844E2B08A079: message-id=<201511220625.tAM6P2J8019998@inside.domain.com>`,
		`Nov 22 09:00:31 mx postfix/cleanup[17854]: DECC7B08A08B: message-id=<>`,
		`Nov 22 02:24:53 mx postfix/cleanup[6917]: B29AFB08A08A: message-id=<mine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com>`,
		`Nov 22 03:47:04 mx postfix/cleanup[9076]: D549FB08A08B: message-id=4ea6228aa481371f55ed30bb51408481@mail.ru`,
	}

	for i, v := range m {
		id := getMessageId(v)

		switch i {
		case 0:
			if id != "E1a0Mks-0000Yl-1x@localhost" {
				t.Errorf("Expected %s value, but got %s", "E1a0Mks-0000Yl-1x@localhost", id)
			}

		case 1:
			if id != "201511220625.tAM6P2J8019998@inside.domain.com" {
				t.Errorf("Expected %s value, but got %s", "201511220625.tAM6P2J8019998@inside.domain.com", id)
			}

		case 3:
			if id != "mine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com" {
				t.Errorf("Expected %s value, but got %s", "mine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com", id)
			}

		case 4:
			if id != "4ea6228aa481371f55ed30bb51408481@mail.ru" {
				t.Errorf("Expected %s value, but got %s", "4ea6228aa481371f55ed30bb51408481@mail.ru", id)
			}

		default:
			if id != "" {
				t.Errorf("Expected empty value, but got %s", id)
			}
		}
	}
}

func TestIsPostfixAndGetMessageId(t *testing.T) {
	var m = []string{
		`Nov 22 07:59:12 mx postfix/cleanup[15917]: B1B8DB08A08B: message-id=<E1a0Mks-0000Yl-1x@localhost>`,
		`Nov 22 09:25:03 mx postfix/cleanup[18849]: 844E2B08A079: message-id=<201511220625.tAM6P2J8019998@inside.domain.com>`,
		`Nov 22 09:00:31 mx postfix/cleanup[17854]: DECC7B08A08B: message-id=<>`,
		`Nov 22 02:24:53 mx postfix/cleanup[6917]: B29AFB08A08A: message-id=<mine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com>`,
	}

	for i, v := range m {
		ok, res := IsPostfix(v, []string{"cleanup"})
		if !ok {
			t.Errorf("Expected true value from IsPostfix, nut got false")

			continue
		}
		id := getMessageId(res[4])

		switch i {
		case 0:
			if id != "E1a0Mks-0000Yl-1x@localhost" {
				t.Errorf("Expextd %s value, but got %s", "E1a0Mks-0000Yl-1x@localhost", id)
			}

		case 1:
			if id != "201511220625.tAM6P2J8019998@inside.domain.com" {
				t.Errorf("Expextd %s value, but got %s", "201511220625.tAM6P2J8019998@inside.domain.com", id)
			}

		case 3:
			if id != "mine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com" {
				t.Errorf("Expextd %s value, but got %s", "mine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com", id)
			}

		default:
			if id != "" {
				t.Errorf("Expextd empty value, but got %s", id)
			}
		}
	}
}

func TestNewValidMailThreadItem(t *testing.T) {
	var m = []string{
		`Nov 22 01:45:57 mx postfix/smtpd[5910]: A2CAFB08A049: client=unknown[1.1.1.1]`,
		`Nov 22 02:08:42 mx postfix/smtpd[6223]: BDCA3B08A08A: client=mrelay1.hh.ru[001.001.001.001]`,
	}

	for _, v := range m {
		if h, err := NewMailThread(v); err != nil {
			t.Errorf("Unexpected error %s", err.Error())
		} else {
			if h == nil {
				t.Errorf("Expected object, but got nil")
				continue
			}
		}
	}
}

func TestNewErrorMailThreadItem(t *testing.T) {
	var m = []string{
		`Nov 22 01:45:57 mx postfix/smtpd[5910]: client=unknown[1.1.1.1]`,
	}

	for _, v := range m {
		if _, err := NewMailThread(v); err == nil {
			t.Errorf("Expected error, but got nil", err.Error())
		}
	}
}

func TestIsRemoved(t *testing.T) {
	var m = []string{
		`Nov 23 09:37:24 mx postfix/qmgr[8015]: 8D7AAB08A049: removed`,
		`Nov 22 01:45:57 mx postfix/smtpd[5910]: A2CAFB08A049: client=unknown[1.1.1.1]`,
		`Nov 22 02:24:53 mx postfix/cleanup[6917]: B29AFB08A08A: message-id=<redmine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com>`,
	}

	var c int = 0

	for _, v := range m {
		if IsRemoved(v) {
			c++
		}
	}

	if c != 1 {
		t.Errorf("Expected %d matches, but got %d", 1, c)
	}
}

func TestGetFrom(t *testing.T) {
	var m = []string{
		`Nov 22 01:45:57 mx postfix/smtpd[5910]: A2CAFB08A049: client=unknown[1.1.1.1]`,
		`Nov 22 02:08:42 mx postfix/smtpd[6223]: BDCA3B08A08A: client=mrelay1.hh.ru[001.001.001.001]`,
		`Nov 22 02:24:53 mx postfix/cleanup[6917]: B29AFB08A08A: message-id=<redmine.issue-12522.20151121232449.dee37bbad722865f@somedomain.com>`,
		`Nov 22 03:47:04 mx postfix/qmgr[8015]: D549FB08A08B: from=<cvetkova09055@mail.ru>, size=1755, nrcpt=1 (queue active)`,
	}

	for i, v := range m {
		from := getFrom(v)

		switch i {
		case 0, 1, 2:
			if from != "" {
				t.Errorf("Expected empty value, but got %s", from)
			}

		case 3:
			if from != "cvetkova09055@mail.ru" {
				t.Errorf("Expected %s value, but got %s", "cvetkova09055@mail.ru", from)
			}
		}
	}
}

func TestGetSpamd(t *testing.T) {
	var m = []string{
		`Nov 22 08:57:02 mx spamd[1800]: spamd: result: . -20 - SHORTCIRCUIT,USER_IN_WHITELIST scantime=0.0,size=982,user=spamd,uid=1003,required_score=5.0,rhost=localhost,raddr=127.0.0.1,rport=41857,mid=<56515923.8060206@somed.foo>,autolearn=disabled,shortcircuit=ham`,
		`Nov 22 09:25:37 mx spamd[1279]: spamd: result: Y 6 - HTML_IMAGE_ONLY_24,HTML_MESSAGE,RCVD_IN_BRBL_LASTEXT,SPF_SOFTFAIL,URIBL_BLACK scantime=0.7,size=4334,user=spamd,uid=1003,required_score=5.0,rhost=localhost,raddr=127.0.0.1,rport=42431,mid=<OTkyNjkxMgAC2616215Y266BAMTQ0ODE3MTExMzE2MDM1@ww2.chilelinks.cl>,autolearn=disabled,shortcircuit=no`,
	}

	var cnt = 0

	for i, v := range m {
		a, err := getSpamd(v)

		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
		}

		if a != nil && a.Score > 0 {
			cnt++
		}

		switch i {
		case 0:
			if a.MsgId != "56515923.8060206@somed.foo" {
				t.Errorf("Expected message id=56515923.8060206@somed.foo, but go %s", a.MsgId)
			}

		case 1:
			if a.MsgId != "OTkyNjkxMgAC2616215Y266BAMTQ0ODE3MTExMzE2MDM1@ww2.chilelinks.cl" {
				t.Errorf("Expected message id=OTkyNjkxMgAC2616215Y266BAMTQ0ODE3MTExMzE2MDM1@ww2.chilelinks.cl, but go %s", a.MsgId)
			}
		}
	}

	if cnt != 1 {
		t.Errorf("Expected spam messages %d, but got %d", 1, cnt)
	}
}

func TestGetSpamdUnknownMsgId(t *testing.T) {
	var m = []string{
		`Nov 22 09:00:31 mx spamd[1279]: spamd: result: . -20 - SHORTCIRCUIT,USER_IN_WHITELIST scantime=0.0,size=955,user=spamd,uid=1003,required_score=5.0,rhost=localhost,raddr=127.0.0.1,rport=41916,mid=< >,autolearn=disabled,shortcircuit=ham`,
		`Nov 22 09:00:31 mx spamd[1279]: spamd: result: . -20 - SHORTCIRCUIT,USER_IN_WHITELIST scantime=0.0,size=955,user=spamd,uid=1003,required_score=5.0,rhost=localhost,raddr=127.0.0.1,rport=41916,mid=(unknown),autolearn=disabled,shortcircuit=ham`,
	}

	for _, v := range m {
		if _, err := getSpamd(v); err != ErrorStrFormatNotSupported {
			t.Errorf("Expected error %s", ErrorStrFormatNotSupported.Error())
		}
	}
}

func TestGetAmavisSpamReport(t *testing.T) {
	var m = []string{
		`Dec  2 16:53:57 mx amavis[30290]: (30290-03) Passed SPAMMY {RelayedTaggedInbound}, [1.1.111.11]:49199 [1.1.111.11] <dashapopovich@yahoo.com> -> <ko@foo.net>,<ko@foo.net>,<sa@foo.net>,<sm@foo.net>,<ret@foo.net>,<lich@foo.net>,<lad@foo.net>,<arov@foo.net>, Queue-ID: 33F124562005, Message-ID: <4FB5F6D3A87C5EFD21246DD33739E940@ip-7-77-51-20.bb.netby.net>, mail_id: Tq0EZ1qm6_xb, Hits: 12.525, size: 34901, queued_as: E27B04562007, 667 ms`,
		`Dec  1 09:59:32 mx amavis[27587]: (27587-19) Passed SPAM, LOCAL [10.10.1.2] [3.6.4.4] <efmattl@interstelloz.co.ua> -> <lova@foo.net>, quarantine: 4/spam-4ubUIQY5JxVu.gz, Message-ID: <b98901d12bc6$ae2ae220$ce0f9374@efmattl>, mail_id: 4ubUIQY5JxVu, Hits: 11.004, size: 144953, queued_as: 354BEA680DB, 749 ms`,
		`Dec 30 12:47:53 mx amavis[12781]: (12781-09) Passed SPAMMY {RelayedTaggedInbound}, [78.186.119.5]:47116 [78.186.119.5] <correspondence3052@list.ru> -> <kaliko@mailserver.net>,<kiselev@mailserver.net>,<userkaa@mailserver.net>,<laptev@mailserver.net>,<vin@mailserver.net>,<shaul@mailserver.net>,<sasa@mailserver.net>,<smirnov@mailserver.net>,<yakov@mailserver.net>,<chay@mailserver.net>,<vlad@mailserver.net>,<userzaa@mailserver.net>, Queue-ID: A3C5A4562012, Message-ID: <32QMPM-XI7RQ9-2N@list.ru>, mail_id: DyyjQNpkYV4x, Hits: 12.406, size: 2835, queued_as: 9AC08456201F, 342 ms`,
		`Dec 30 12:54:47 mx amavis[12781]: (12781-13) Passed SPAMMY {RelayedTaggedInbound}, [117.68.193.4]:3223 [117.68.193.4] <fytygxa@uvupa.com> -> <kaliko@mailserver.net>,<userkaa@mailserver.net>,<sasa@mailserver.net>,<smirnov@mailserver.net>,<yakov@mailserver.net>,<chay@mailserver.net>,<vlad@mailserver.net>,<userzaa@mailserver.net>, Queue-ID: AE1454562012, mail_id: evYBMeO836NQ, Hits: 6.76, size: 1425, queued_as: EC3AE456201F, 13114 ms`,
		`Sep  5 11:21:07 mx amavis[18886]: (18886-11) Blocked BANNED (.asc,credit_card_receipt_D0CDD328.js) {DiscardedInbound,Quarantined}, [188.174.71.22]:59058 [188.174.71.22] <Blevins.9952@idotconnect.net> -> <gooduser@mailserver.net>, quarantine: banned@mailserver.net, Queue-ID: 26866B08A06B, Message-ID: <99564dd4f5dd96bf9dc7541eea5b75ba@,ailserver.net, mail_id: 7Vg0b3luDu2U, Hits: -, size: 16447, 181 ms`,
	}

	for i, v := range m {
		f, err := getAmavisd(v)

		if err != nil {
			t.Errorf("Expected error %s", ErrorStrFormatNotSupported.Error())
		}

		if f == nil {
			t.Errorf("Expected Spam object, but got nil")
		}

		switch i {
		case 0:
			if f.Score != 8 {
				t.Errorf("Expected score %d, but got %d", 8, f.Score)
			}

			if f.QueueId != "33F124562005" {
				t.Errorf("Expected Queue-ID %s, but got %s", "33F124562005", f.QueueId)
			}

			if f.QueuedAs != "E27B04562007" {
				t.Errorf("Expected queued_as %s, but got %s", "E27B04562007", f.QueuedAs)
			}

		case 1:
			if f.Score != 1 {
				t.Errorf("Expected score %d, but got %d", 1, f.Score)
			}

			if f.QueuedAs != "354BEA680DB" {
				t.Errorf("Expected queued_as %s, but got %s", "354BEA680DB", f.QueuedAs)
			}

		case 2:
			if f.Score != 12 {
				t.Errorf("Expected score %d, but got %d", 12, f.Score)
			}

		case 3:
			if f.Score != 8 {
				t.Errorf("Expected score %d, but got %d", 8, f.Score)
			}

		case 4:
			if f.QueueId != "26866B08A06B" {
				t.Errorf("Expected Queue-ID %s, but got %s", "26866B08A06B", f.QueueId)
			}
		}
	}
}

func TestGetAmavisNotSpamReport(t *testing.T) {
	var m = []string{
		`Dec  2 16:53:57 mx amavis[30290]: (30290-03) Passed CLEAN {RelayedTaggedInbound}, [1.1.111.11]:49199 [1.1.111.11] <dashapopovich@yahoo.com> -> <ko@foo.net>,<ko@foo.net>,<sa@foo.net>,<sm@foo.net>,<ret@foo.net>,<lich@foo.net>,<lad@foo.net>,<arov@foo.net>, Queue-ID: 33F124562005, Message-ID: <4FB5F6D3A87C5EFD21246DD33739E940@ip-7-77-51-20.bb.netby.net>, mail_id: Tq0EZ1qm6_xb, Hits: 12.525, size: 34901, queued_as: E27B04562007, 667 ms`,
		`Dec  1 09:59:32 mx amavis[27587]: (27587-19) Passed CLEAN, LOCAL [10.10.1.2] [3.6.4.4] <efmattl@interstelloz.co.ua> -> <lova@foo.net>, quarantine: 4/spam-4ubUIQY5JxVu.gz, Message-ID: <b98901d12bc6$ae2ae220$ce0f9374@efmattl>, mail_id: 4ubUIQY5JxVu, Hits: 11.004, size: 144953, queued_as: 354BEA680DB, 749 ms`,
	}

	for _, v := range m {
		f, err := getAmavisd(v)

		if err != nil {
			t.Errorf("Expected error %s", ErrorStrFormatNotSupported.Error())
		}

		if f == nil {
			t.Errorf("Expected Spam object, but got nil")
		}

		if f.Score != 0 {
			t.Errorf("Expected score %d, but got %d", 0, f.Score)
		}
	}
}

func TestGetQueuedAsIdValue(t *testing.T) {
	var (
		m = []string{
			`Dec  4 10:33:25 mx postfix/smtp[14678]: 5247C4562029: to=<lik@some.net>, relay=127.0.0.1[127.0.0.1]:10024, delay=1.3, delays=0.67/0/0/0.66, dsn=2.0.0, status=sent (250 2.0.0 from MTA(smtp:[127.0.0.1]:10025): 250 2.0.0 Ok: queued as F3C59456202A)`,
		}
	)

	for _, l := range m {
		ok, p := IsPostfix(l, []string{"smtp"})

		if !ok {
			t.Errorf("Line was not recognized as smtp log entry, %s", l)
		}

		if v := getIdQueuedAs(p[4]); v != "F3C59456202A" {
			t.Errorf("Expected queued id F3C59456202A, but got %s", v)
		}
	}
}

func TestNewMailThreadChildIdValue(t *testing.T) {
	var (
		m = []string{
			`Dec  4 10:33:25 mx postfix/smtp[14678]: 5247C4562029: to=<lik@some.net>, relay=127.0.0.1[127.0.0.1]:10024, delay=1.3, delays=0.67/0/0/0.66, dsn=2.0.0, status=sent (250 2.0.0 from MTA(smtp:[127.0.0.1]:10025): 250 2.0.0 Ok: queued as F3C59456202A)`,
		}
	)

	for _, l := range m {
		h, err := NewMailThread(l)

		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())

			continue
		}

		if h == nil {
			t.Errorf("Expected object but got nil")

			continue
		}

		if h.childId != "F3C59456202A" {
			t.Errorf("Expected child queued id F3C59456202A, but got %s", h.childId)
		}
	}
}

func TestGetSmtpStatusValue(t *testing.T) {
	var (
		m = []string{
			`Dec 27 02:39:24 mx postfix/smtp[15816]: C93CEB08A046: to=<order@arobor.ru>, relay=none, delay=307136, delays=307106/0.02/30/0, dsn=4.4.1, status=deferred (connect to arobor.ru[7.0.7.1]:25: Connection timed out)`,
			`Dec 27 03:25:40 mx postfix/smtp[18396]: AC807B08A04A: to=<ovad@domain.com>, relay=127.0.0.1[127.0.0.1]:10024, delay=13, delays=1.2/0.01/0.01/11, dsn=2.0.0, status=sent (250 2.0.0 from MTA(smtp:[127.0.0.1]:10025): 250 2.0.0 Ok: queued as 346DC4562001)`,
		}
	)

	for i, l := range m {
		switch i {
		case 0:
			if v := getSmtpStatus(l); v != "deferred" {
				t.Errorf("Expected deferred status, but got %s", v)
			}
		case 1:
			if v := getSmtpStatus(l); v != "sent" {
				t.Errorf("Expected sent status, but got %s", v)
			}
		}
	}
}

func TestGetLogEntryTime(t *testing.T) {
	var (
		m = []string{
			`Nov 22 04:06:51 mx postfix/smtpd[9477]: connect from unknown[127.0.0.1]`,
		}
	)

	for i, l := range m {
		v, err := getTime(l)

		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
			continue
		}

		switch i {
		case 0:
			if ti := v.Format(time.Stamp); ti != "Nov 22 04:06:51" {
				t.Errorf("Expected parsed time value `Nov 22 04:06:51`, but got %s", ti)
			}
		}
	}
}
