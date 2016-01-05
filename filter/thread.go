package filter

import (
	"time"
)

type MailThread struct {
	Id       string
	MsgId    string
	From     string
	childId  string
	parentId string

	SpamScore  uint
	smtpStatus uint8

	Client  *Client
	Removed bool
}

type ThreadFace interface {
	GetId() string
	GetChildId() string
	GetMessageId() string
	GetFrom() string
	GetFromIp() string
	GetTime() time.Time
	GetSpamScore() uint
}

func NewMailThread(str string) (m *MailThread, err error) {
	var (
		id string
	)

	id, err = getId(str)

	if err != nil {
		return nil, err
	}

	m = &MailThread{
		Id:        id,
		From:      getFrom(str),
		SpamScore: 0,
		Removed:   IsRemoved(str),
	}

	m.Client = getClient(str)

	if ok, mid := IsPostfix(str, []string{"cleanup"}); ok && len(mid) > 4 {
		m.MsgId = getMessageId(mid[4])
	}

	if ok, mid := IsPostfix(str, []string{"smtp"}); ok && len(mid) > 4 {
		m.childId = getIdQueuedAs(mid[4])

		switch getSmtpStatus(mid[4]) {
			case "sent":
				m.smtpStatus = 1

			case "deferred":
				m.smtpStatus = 2
		}
	}

	return m, nil
}

// Get thread ID
func (this *MailThread) GetId() string {
	return this.Id
}

// Get child thread ID
func (this *MailThread) GetChildId() string {
	return this.childId
}

// Get message id
func (this *MailThread) GetMessageId() string {
	return this.MsgId
}

// Get from value
func (this *MailThread) GetFrom() string {
	return this.From
}

// Get client ip
func (this *MailThread) GetFromIp() (v string) {
	if this.Client != nil {
		v = this.Client.IP
	}
	return v
}

// Return time value when mail was accepted by server for the delivery
func (this *MailThread) GetTime() (t time.Time) {
	if this.Client != nil {
		t = this.Client.At
	}
	return t
}

func (this *MailThread) GetSpamScore() uint {
	return this.SpamScore
}

// Add new values to the object
func (this *MailThread) apply(m *MailThread) error {
	if this.Id != m.Id {
		return ErrorItemWrongId
	}

	if this.From == "" && m.From != "" {
		this.From = m.From
	}

	if this.MsgId == "" && m.MsgId != "" {
		this.MsgId = m.MsgId
	}

	if this.Client == nil && m.Client != nil {
		*this.Client = *m.Client
	}

	if this.childId == "" && m.childId != "" {
		this.childId = m.childId
	}

	if this.smtpStatus == 0 && m.smtpStatus > 0 {
		this.smtpStatus = m.smtpStatus
	}

	if !this.Removed && m.Removed {
		this.Removed = true
	}

	return nil
}
