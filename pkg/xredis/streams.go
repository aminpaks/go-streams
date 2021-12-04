package xredis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

type RedisClient = redis.Client
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

	consumerCounter := uint(1)
	for {
		if consumerCounter > options.Counts {
			break
		}
		go func() {
			consumerId := uuid.New().String()
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
					// NOGROUP is returned when the group doesn't exists
					if strings.Contains(err.Error(), "NOGROUP") {
						if b, err := client.SetNX(fmt.Sprintf("stream-[%s]-creation-lock", streamName), consumerId, time.Second*1).Result(); err != nil {
							log.Printf("consumer %s waiting for lock", consumerId)
							continue
						} else if !b {
							time.Sleep(time.Second * 1)
							continue
						}
						err = client.XGroupCreateMkStream(streamName, groupName, "0").Err()
						// BUSYGROUP is returned when the group already exists
						// this error can happend if there are multiple consumers
						if err != nil {
							if strings.Contains(err.Error(), "BUSYGROUP") {
								continue
							}
							log.Printf("ERROR: consumer '%s' failed to create stream '%s': %v", consumerId, streamName, err)
						}
					} else {
						log.Printf("ERROR: consumer '%s' failed to read stream '%s': %v", consumerId, streamName, err)
					}
					time.Sleep(time.Second * 5)
					continue
				}
				if len(entries) > 0 {
					for i := range entries[0].Messages {
						message := &entries[0].Messages[i]
						messageID := message.ID
						err = client.XAck(streamName, groupName, messageID).Err()
						if err != nil {
							log.Printf("failed to ack stream entry %s: %v", messageID, err)
						}

						entryData, err := parseStreamEntry(message.Values)
						if err != nil {
							log.Printf("failed to decode entry element: %v -> %s", err, message.Values)
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
								log.Printf("consumer '%s' failed to process stream '%s' entry -> %s", consumerId, streamName, message.Values)
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

func List(l []string) string {
	b, _ := json.MarshalIndent(l, "", "  ")
	return string(b)
}
func ParseList(v []byte) []string {
	list := []string{}
	_ = json.Unmarshal(v, &list)
	return list
}

func Lock(c *redis.Client, key string) (release func() error) {
	for {
		if b, err := c.SetNX(key, "LOCK", time.Hour*999).Result(); err != nil || !b {
			continue
		}
		break
	}
	return func() error {
		if _, err := c.Del(key).Result(); err != nil {
			return fmt.Errorf("failed to release lock '%s'", key)
		}
		return nil
	}
}
