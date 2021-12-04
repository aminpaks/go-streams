package users

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/aminpaks/go-streams/pkg/xredis"
)

func userCreationConsumer() xredis.StreamConsumerFunc {
	return func(entry xredis.XStreamEntry, consumerId string) error {
		serializedValue := entry.Value
		isLastTry := entry.IsLastTry()
		lastError := entry.LastError

		// Parses the entry value and should be a User type
		user, err := ParseUser(serializedValue)
		if err != nil {
			// If the entry cannot be parse there is no need to retry processing this entry
			// We return nil and just report the invalid entry
			log.Printf("invalid entry got lost: %v -- entry -> %s", err, serializedValue)
			return nil
		}

		rnd := rand.Float32()
		// Simulation of random failure to demo the retries mechanics
		// If the condition is true we will fail the process
		if rnd >= 0.4 /* 60% chance of failure */ {
			// if this is not the last retry then we fail
			if !isLastTry {
				// if anything is wrong we can simply retry processing this entry by returning an error
				return fmt.Errorf("failed to process, random number '%f'", rnd)
			} else {
				// if this is the last try we must return nil and log what happened
				log.Printf("failed to process entry on %d tries, reference: %s, last error: %v - value: %s", entry.Retries, entry.Id, lastError, serializedValue)
				return nil
			}
		}

		// Do something useful with this entry
		log.Printf("user %s processed on %d retries successfully - consumer: %s", user.Name, entry.Retries, consumerId)

		return nil
	}
}
