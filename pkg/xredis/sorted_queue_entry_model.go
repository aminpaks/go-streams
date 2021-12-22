package xredis

import (
	"encoding/json"
	"time"
)

type XSortedQueueEntry struct {
	// Internal use only
	c *RedisClient `json:"-"`

	retry          bool   `json:"-"`
	maxRetries     int    `json:"-"`
	currentFailure error  `json:"-"`
	queue          string `json:"-"`

	Value          string        `json:"value,omitempty"`
	Priority       float64       `json:"priority"`
	ReferenceUri   string        `json:"referenceUri"`
	Expiration     time.Duration `json:"expiration"`
	Failures       []string      `json:"failures,omitempty"`
	CurrentRetries int           `json:"currentRetries"`
}
type internalPersistedXSortedQueueEntry struct {
	Retries      int           `json:"retries"`
	Value        string        `json:"value"`
	Priority     float64       `json:"priority"`
	ReferenceUri string        `json:"referenceUri"`
	Expiration   time.Duration `json:"expiration"`
	Failures     []string      `json:"failures,omitempty"`
}

func NewXSortedQueueEntry(value string, priority float64, referenceUri string, expiration time.Duration) XSortedQueueEntry {
	return XSortedQueueEntry{
		Value:        value,
		Priority:     priority,
		ReferenceUri: referenceUri,
		Expiration:   expiration,
		Failures:     []string{},
	}
}

func serializedXSortedQueueEntry(i XSortedQueueEntry, try int) string {
	b, _ := json.Marshal(internalPersistedXSortedQueueEntry{
		Retries:      try,
		Value:        i.Value,
		Priority:     i.Priority,
		ReferenceUri: i.ReferenceUri,
		Expiration:   i.Expiration,
		Failures:     i.Failures,
	})

	return string(b)
}

func (x *XSortedQueueEntry) String() string {
	b, _ := json.Marshal(x)
	return string(b)
}

func (x *XSortedQueueEntry) IsLastRetry() bool {
	return x.CurrentRetries >= x.maxRetries
}

func (x *XSortedQueueEntry) HasExhaustedRetries() bool {
	return x.CurrentRetries > x.maxRetries
}

func (x *XSortedQueueEntry) Retry(failure error) {
	x.retry = true
	if failure != nil {
		x.setFailure(failure)
	}
}

func (x *XSortedQueueEntry) Ack() error {
	if err := ackSortedQueueEntry(x.c, x.queue, *x); err != nil {
		x.setFailure(err)
		return err
	}
	return nil
}

func (x *XSortedQueueEntry) CleanUp() error {
	if err := cleanSortedQueueEntry(x.c, *x); err != nil {
		x.setFailure(err)
		return err
	}
	return nil
}

func (x *XSortedQueueEntry) setFailure(failure error) {
	x.currentFailure = failure
	if failure != nil {
		x.Failures = append(x.Failures, failure.Error())
	}
}
func (x *XSortedQueueEntry) setClient(c *RedisClient) {
	if x.c == nil {
		x.c = c
	}
}
