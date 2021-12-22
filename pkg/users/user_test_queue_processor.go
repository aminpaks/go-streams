package users

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/aminpaks/go-streams/pkg/throttler"
	"github.com/aminpaks/go-streams/pkg/xredis"
)

func testQueueEntryConsumer(worker *throttler.Throttler) xredis.XSortedQueueEntryConsumerFunc {
	rand.Seed(rand.Int63())

	return func(entries []xredis.XSortedQueueEntry, consumerId string) []xredis.XSortedQueueEntry {
		log.Printf("batch length: %d -- %s", len(entries), consumerId)

		for i := range entries {
			e := &entries[i]
			log.Printf("** %s :: processing entry %s -- retries: %d", consumerId, e.ReferenceUri, e.CurrentRetries)
			// Doing some work
			dur, err := worker.Work(e.ReferenceUri)
			// Checks if work was completed or not
			if err != nil {
				// Reports back the entry to queue to keep the item for later processing
				e.Retry(fmt.Errorf("some work: %v", err))
				continue
			}

			// Let's play some random scenarios
			rnd := rand.Float32()

			// 10% chance that we're gonna retry this entry
			if rnd <= 0.1 {
				e.Retry(fmt.Errorf("failed due to %f is less than 0.1", rnd))
				log.Printf("entry failed randomly, gonna retry %s -- rnd %f", e.ReferenceUri, rnd)
				continue
			} else {
				log.Printf("done some work that took: %s", dur)

				// We must always acknowledge the queue entry once the work is complete
				// otherwise it will be passed to failure handler as a violation.
				if err := e.Ack(); err != nil {
					// We don't retry internal errors, this entry will be passed to failure handler
					log.Printf("failed to acknowledge entry: %v", err)
					continue
				}
				log.Printf("++ %s :: entry %s processed successfully -- retries %d, priority: %f", consumerId, e.ReferenceUri, e.CurrentRetries, e.Priority)
			}
		}

		return entries
	}
}

func testQueueFailureHandler() xredis.XSortedQueueFailureHandlerFunc {
	return func(failures []xredis.XFailure, consumerId string) {
		// Failures can be caused by all sorta errors, there is a guarantee that we will
		// always receive at least one item in the slice
		for _, f := range failures {
			// Just notify for the sake of the demo
			log.Printf("-- %s :: handling failure -- %v --> (%v)", consumerId, f.Err, f.Payload.String())

			// Payload of failure can include the entry too, in that case we will handle
			// the entry manually or wanna correct its state and queue it for retry.
			// IMPORTANT NOTE: We must always clean the failed entries as they will NOT
			//                 be cleaned up automatically by the queue.
			switch e := f.Payload["entry"].(type) {
			case xredis.XSortedQueueEntry:
				// Attempt to clean up the resource
				if err := e.CleanUp(); err != nil {
					log.Printf("Failed to clean up: %v", err)
				}
			}
		}
	}
}
