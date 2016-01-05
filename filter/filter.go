package filter

import (
	"regexp"
	"strings"
	"time"
)

var (
	amavisdRe,
	amavisEmlRe,
	amavisQueueRe,
	clientRe,
	fromRe,
	messageIdRe,
	postfixRe,
	queuedasRe,
	smtpstatusRe,
	spamdRe,
	timeRe *regexp.Regexp

	emailTpl string = `[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+`
)

type Client struct {
	Name, IP string
	At time.Time
}


func init() {
	// Pickup amavis log entry with statistics
	amavisdRe = regexp.MustCompile(`amavis\[(\d+)]\: \([0-9\-]+\) Passed (CLEAN|SPAM|SPAMMY).*\[\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\] \<([a-zA-Z0-9-_\.@]{1,})\> \-\> (.*)`)
	// Find emails list in the amavis statistics message
	amavisEmlRe = regexp.MustCompile(`(\<` + emailTpl + `\>\,){1,}`)
	// Find queued_as parameter in amavis message
	amavisQueueRe = regexp.MustCompile(`([Qq]ueue[_\-IDdas]+)\: ([a-zA-Z0-9]+)\,`)
	// Pick up client information from the postfix message
	clientRe = regexp.MustCompile(`client\=([a-zA-Z0-9-_\.]+)\[(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\]`)
	// Pick up email data from postfix message
	fromRe = regexp.MustCompile(`from\=\<(` + emailTpl + `)\>,`)
	// Common pattern to pick up message id from amavis or spamd message
	messageIdRe = regexp.MustCompile(`[Mm]essage\-[Ii][Dd](\=|\:)[\s\<]*([a-zA-Z0-9\-\_\.@\$]{1,})\>*`)
	// Postfix modules log messages
	postfixRe = regexp.MustCompile(` postfix\/(\w+)\[(\d+)\]\: ([a-zA-Z0-9]+)\: (.*)`)
	// Take message id from queued as string
	queuedasRe = regexp.MustCompile(`queued as ([a-zA-Z0-9]{1,})\)`)
	// Get smtp status
	smtpstatusRe = regexp.MustCompile(`status=(sent|deferred)`)
	// Spamd log message
	spamdRe = regexp.MustCompile(`spamd\[(\d+)\]: spamd: result: ([\.Y]{1}) ([\-\d]{1,}) - .*,mid\=\<*([a-zA-Z0-9-_\.@\$]{1,})\>*,`)
	// Log line time
	timeRe = regexp.MustCompile(`^(\w+\s+\d{1,2} \d{1,2}:\d{1,2}:\d{1,2})`)
}

// Check string if there is postfix system record
func IsPostfix(str string, service []string) (ok bool, res []string) {
	res = postfixRe.FindStringSubmatch(str)

	if len(res) == 0 {
		return false, res
	}

	// The first item is the whole regexp expression
	if len(service) == 0 {
		return true, res
	}

	if len(res) > 1 {
		for _, v := range service {
			if res[1] == v {
				return true, res
			}
		}
	}

	return false, res
}

// Check string if there is postfix cleanup process
func IsCleanup(str string) bool {
	ok, _ := IsPostfix(str, []string{"cleanup"})
	return ok
}

// Check string if there is mail thread item removed
func IsRemoved(str string) bool {
	ok, res := IsPostfix(str, []string{"qmgr"})

	if !ok || len(res) < 5 {
		return false
	}

	res[4] = strings.ToLower(strings.Trim(res[4], " "))
	return (res[4] == "removed")
}

// Get mail item thread id (this is not system process id)
func getId(str string) (v string, err error) {
	var (
		res []string
		ok  bool
	)

	ok, res = IsPostfix(str, []string{"smtpd", "smtp", "cleanup", "qmgr", "pipe"})
	// Is not postfix message
	if !ok || len(res) < 4 {
		return v, ErrorStrFormatNotSupported
	}

	res[3] = strings.ToUpper(res[3])

	switch res[3] {
	case "WARNING", "NOQUEUE":
		return v, ErrorStrFormatNotSupported

	default:
		v = strings.Trim(res[3], " ")
	}

	if v == "" {
		err = ErrorItemEmptyId
	}

	return
}

// Get time object from string
func getTime(str string) (t time.Time, err error) {
	var (
		res []string
	)

	res = timeRe.FindStringSubmatch(str)
	if len(res) < 2 {
		return t, ErrorStrFormatNotSupported
	}

	return time.Parse(time.Stamp, res[1])
}

// Get connected client identity
func getClient(str string) (v *Client) {
	var (
		res []string
		ok bool
	)

	ok, res = IsPostfix(str, []string{"smtpd"})
	// Is not postfix message
	if !ok || len(res) < 5 {
		return nil
	}

	res = clientRe.FindStringSubmatch(res[4])

	if len(res) < 3 {
		return nil
	}

	v = &Client{
		Name: res[1],
		IP:   res[2],
	}

	// Take time when mail was accepted for the delivery
	if t, err := getTime(str); err == nil {
		v.At = t
	}

	return v
}

// Get message identity from cleanup service
func getMessageId(str string) (v string) {
	var strs []string = messageIdRe.FindStringSubmatch(str)

	if len(strs) > 2 {
		v = strs[2]
	}

	return v
}

// Get "queued as" pointer
func getIdQueuedAs(str string) (v string) {
	var strs []string = queuedasRe.FindStringSubmatch(str)

	if len(strs) > 1 {
		v = strs[1]
	}

	return v
}

// Get from address
func getFrom(str string) (v string) {
	ok, res := IsPostfix(str, []string{"qmgr"})
	if !ok || len(res) < 5 {
		return v
	}

	if res := fromRe.FindStringSubmatch(res[4]); len(res) > 1 {
		v = res[1]
	}

	return v
}

// Get smtp status
func getSmtpStatus(str string) (v string) {
	ok, res := IsPostfix(str, []string{"smtp"})

	if !ok || len(res) < 5 {
		return v
	}

	if res = smtpstatusRe.FindStringSubmatch(res[4]); len(res) > 1 {
		v = res[1]
	}

	return v
}
