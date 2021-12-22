package engine

import (
	"log"

	"github.com/aminpaks/go-streams/pkg/xredis"
)

func EngineConsumer() xredis.StreamConsumerFunc {
	return func(entry xredis.XStreamEntry, consumerId string) error {
		log.Printf("entry: %v", entry.Value)
		return nil
	}
}
