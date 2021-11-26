package xredis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

type StreamConsumerFunc func(entry XStreamEntry, consumerId string) error

var ErrStreamConsumer = errors.New("stream consumer")
var ErrStreamAppend = errors.New("failed to append")

func RegisterConsumer(ctx context.Context, streamName string, groupName string, consumerFn StreamConsumerFunc, options *StreamConsumerOptions) error {
	// Get the Redis client from dependency context
	client, err := GetClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStreamConsumer, err)
	}
	if options == nil {
		options = NewStreamConsumerOptions(1, 5)
	}
	options.Normalize() // Removes invalid options

	err = client.XGroupCreateMkStream(streamName, groupName, "0").Err()
	// BUSYGROUP is returned when the group already exists
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	consumerCounter := uint(1)
	for consumerCounter <= options.Counts {
		consumerId := uuid.New().String()
		go func() {
			log.Printf("Consumer %s running on stream '%s'", consumerId, streamName)
			for {
				entries, err := client.XReadGroup(&redis.XReadGroupArgs{
					Group:    groupName,
					Consumer: consumerId,
					Streams:  []string{streamName, ">"},
					Count:    2,
					Block:    0,
					NoAck:    false,
				}).Result()
				if err != nil {
					log.Fatalf("consumer '%s' failed to read stream '%s': %v", consumerId, streamName, err)
				}
				if len(entries) > 0 {
					for i := range entries[0].Messages {
						message := &entries[0].Messages[i]
						messageID := message.ID
						err = client.XAck(streamName, groupName, messageID).Err()
						if err != nil {
							log.Fatalf("failed to ack stream entry %s: %v", messageID, err)
						}

						entryData, err := parseStreamEntry(message.Values)
						if err != nil {
							log.Fatalf("failed to decode entry element: %v -> %s", err, message.Values)
						}
						entryErr := consumerFn(
							*entryData.
								WithIncreaseTries().
								withMaxRetries(options.Retries),
							consumerId,
						)

						if entryErr != nil {
							if entryData.Retries < options.Retries {
								// log.Printf("retrying %d stream entry -> %s", entryData.Retries, entryData.Value)
								internalStreamAppend(ctx, streamName, entryData.
									WithError(entryErr.Error()).
									Build(),
								)
							} else {
								log.Fatalf("consumer '%s' failed to process stream '%s' entry -> %s", consumerId, streamName, message.Values)
							}
						}
					}
				}
			}
		}()
		consumerCounter += 1
	}

	return nil
}

func internalStreamAppend(depsCtx context.Context, streamName string, values map[string]interface{}) error {
	client, err := GetClient(depsCtx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStreamAppend, err)
	}

	err = client.XAdd(&redis.XAddArgs{
		Stream:       streamName,
		MaxLen:       0,
		MaxLenApprox: 0,
		ID:           "",
		Values:       values,
	}).Err()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStreamAppend, err)
	}

	return nil
}

func StreamAppend(depsCtx context.Context, streamName string, value string) (entryRef uuid.UUID, err error) {
	entryRef = uuid.New()
	err = internalStreamAppend(depsCtx, streamName, newStreamEntry(entryRef, value, 0, "").Build())
	return entryRef, err
}
