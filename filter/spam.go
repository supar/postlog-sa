package filter

import (
	"regexp"
	"strings"
)

type Spam struct {
	MsgId,
	QueueId,
	QueuedAs string
	Score uint
}

// Exemain log line and found if there is spam information
func NewSpam(str string) (sp *Spam, err error) {
	if sp, err = getSpamd(str); err != nil {
		if err != ErrorStrFormatNotSupported {
			return
		}
		// Reset error
		err = nil
	}

	// Do not check amavis if spamd was found
	if sp == nil {
		if sp, err = getAmavisd(str); err != nil {
			if err != ErrorStrFormatNotSupported {
				return
			}
			// Reset error
			err = nil
		}
	}

	return
}

// Check if string is spamd message
func getSpamd(str string) (s *Spam, err error) {
	var (
		res []string
	)

	if res = spamdRe.FindStringSubmatch(str); len(res) < 5 {
		return nil, ErrorStrFormatNotSupported
	}

	s = &Spam{
		Score: 0,
	}

	if strings.Contains(strings.ToLower(res[2]), "y") {
		s.Score++
	}

	s.MsgId = res[4]

	return
}

// Check if string is amavis message
func getAmavisd(str string) (s *Spam, err error) {
	var (
		res []string
	)

	if res = amavisdRe.FindStringSubmatch(str); len(res) < 5 {
		return nil, ErrorStrFormatNotSupported
	}

	// Convert to upper
	res[2] = strings.ToUpper(res[2])

	s = &Spam{
		Score: 0,
	}

	s.MsgId = getMessageId(res[4])

	// Pick up queues values
	if q := amavisQueueRe.FindAllStringSubmatch(res[4], -1); len(q) > 0 {
		for _, val := range q {
			if len(val) < 3 {
				continue
			}

			val[1] = strings.ToUpper(strings.Trim(val[1], " "))

			switch val[1] {
			case "QUEUE-ID":
				s.QueueId = val[2]
			case "QUEUED_AS":
				s.QueuedAs = val[2]
			}
		}
	}

	if !strings.HasPrefix(res[2], "SPAM") &&
		!strings.HasPrefix(res[2], "BANN") {
		return
	}

	if res = amavisEmlRe.FindStringSubmatch(res[4]); len(res) > 0 {
		emls := strings.Split(res[0], ",")

		for _, i := range emls {
			i = strings.Trim(i, " <>")

			if ok, err := regexp.MatchString(emailTpl, i); err != nil {
				return nil, err
			} else {
				if ok {
					s.Score++
				}
			}
		}
	}

	return
}
