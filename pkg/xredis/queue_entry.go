package xredis

import (
	"encoding/json"
	"fmt"
)

type XQueueEntry struct {
	ReferenceUri string `json:"referenceUri"`
	Value        string `json:"value"`
}

func NewXQueueEntry(value string) XQueueEntry {
	return NewXQueueEntryByReference(value, "")
}
func NewXQueueEntryByReference(value string, referenceUri string) XQueueEntry {
	return XQueueEntry{
		ReferenceUri: referenceUri,
		Value:        value,
	}
}

func (xqe *XQueueEntry) String() string {
	b, _ := json.Marshal(xqe)
	return string(b)
}

func parseQueueEntry(i interface{}) (queueEntry XQueueEntry, err error) {
	var b []byte
	switch v := i.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return XQueueEntry{}, fmt.Errorf("type %T is not valid", i)
	}
	err = json.Unmarshal(b, &queueEntry)
	return queueEntry, err
}
