package filter

import (
	"errors"
)

var (
	ErrorStrFormatNotSupported = errors.New("String format is not supported")
	ErrorStorageItemExists     = errors.New("Item with same id exists")
	ErrorUnknownSpamItem       = errors.New("Unknown spam entry")
	ErrorItemEmptyId           = errors.New("Item Id field is empty value")
	ErrorItemWrongId           = errors.New("Item Id is not same")
)

type Storage struct {
	threadDone func(ThreadFace)
	keepData   bool
	data       map[string]*MailThread
}

// Craete storage instance
func NewStorage() (s *Storage) {
	s = &Storage{
		threadDone: func(v ThreadFace) {},
		data:       make(map[string]*MailThread),
	}

	return
}

// Set keepData flag
func (this *Storage) SetKeepData(v bool) {
	this.keepData = v
}

// Set call back func on main thread information is fill full
func (this *Storage) SetThreadDoneCb(fn func(ThreadFace)) {
	this.threadDone = fn
}

// Get data storage length
func (this *Storage) Len() int {
	return len(this.data)
}

// Get thread object from data storage
func (this *Storage) Get(id string) (m *MailThread) {
	var ok bool = false

	if m, ok = this.data[id]; !ok {
		return nil
	}

	return m
}

// Add new items to the storage
func (this *Storage) Set(m *MailThread) {
	var (
		item,
		child *MailThread
		ok bool
	)

	if m == nil || m.GetId() == "" {
		return
	}

	if item, ok = this.data[m.Id]; ok {
		item.apply(m)
	} else {
		this.data[m.Id] = m
		item = m
	}

	if item.childId != "" {
		if child = this.Get(item.childId); child != nil {
			child.parentId = item.Id
		}
	}
}

// Test each thread to run callback function
func (this *Storage) ThreadDone(m *MailThread) {
	var (
		item,
		child,
		parent *MailThread
	)

	if m == nil || m.GetId() == "" {
		return
	}

	item = this.Get(m.GetId())

	// Get child threa
	if child = this.Get(item.childId); child == nil {
		child = this.Get(item.GetId())
	}

	// Get parent thread
	if parent = this.Get(item.parentId); parent == nil {
		parent = this.Get(item.GetId())
	}

	if parent.Removed == true {
		if parent == child {
			this.threadDone(parent)
			this.Destroy(parent.GetId())
		} else {
			if parent.Removed == child.Removed {
				parent.SpamScore = child.SpamScore

				this.threadDone(parent)
				this.Destroy(parent.GetId())
				this.Destroy(child.GetId())
			}
		}
	}
}

// Write spam statistics to the mail thread
func (this *Storage) SetSpamStat(sp *Spam) (err error) {
	if sp == nil {
		return ErrorUnknownSpamItem
	}

	for _, item := range this.data {
		if item.MsgId == sp.MsgId {
			if sp.QueuedAs != "" && sp.QueuedAs != item.Id {
				continue
			}

			item.SpamScore += sp.Score

			return
		}
	}

	return ErrorUnknownSpamItem
}

// Destroy mail thread
func (this *Storage) Destroy(id string) {
	if !this.keepData {
		delete(this.data, id)
	}
}
