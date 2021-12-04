package xredis

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const entrySerializedElementKey = "serializedEntryElement"

type XStreamEntry struct {
	maxRetries int
	Id         uuid.UUID `json:"id"`
	LastError  string    `json:"lastError"`
	Retries    int       `json:"retries"`
	Value      string    `json:"value"`
}

func (se *XStreamEntry) Build() map[string]interface{} {
	b, _ := json.Marshal(se)
	return map[string]interface{}{
		entrySerializedElementKey: b,
	}
}

func (se *XStreamEntry) WithIncreaseTries() *XStreamEntry {
	se.Retries += 1
	return se
}

func (se *XStreamEntry) withMaxRetries(v int) *XStreamEntry {
	se.maxRetries = v
	return se
}

func (se *XStreamEntry) IsLastTry() bool {
	return se.Retries >= se.maxRetries
}

func (se *XStreamEntry) WithError(err string) *XStreamEntry {
	se.LastError = err
	return se
}

func newStreamEntry(id uuid.UUID, element string, retries int, lastError string) *XStreamEntry {
	return &XStreamEntry{
		Id:        id,
		Retries:   retries,
		Value:     element,
		LastError: lastError,
	}
}

func parseStreamEntry(i map[string]interface{}) (*XStreamEntry, error) {
	var serializedValue []byte
	switch v := i[entrySerializedElementKey].(type) {
	case string:
		serializedValue = []byte(v)
	case []byte:
		serializedValue = v
	default:
		return nil, fmt.Errorf("failed to decode XStreamEntry: invalid input value %T", i[entrySerializedElementKey])
	}

	var entry XStreamEntry
	err := json.Unmarshal(serializedValue, &entry)
	return &entry, err
}
