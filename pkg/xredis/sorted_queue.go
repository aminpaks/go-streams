package xredis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type XSortedQueueEntryConsumerFunc func(entries []XSortedQueueEntry, consumerId string) []XSortedQueueEntry
type XSortedQueueFailureHandlerFunc func(failures []XFailure, consumerId string)

func NewSortedQueueConsumer(
	client *RedisClient,
	ctx context.Context,
	queue string,
	entryConsumer XSortedQueueEntryConsumerFunc,
	failureHandler XSortedQueueFailureHandlerFunc,
) chan struct{} {
	return NewSortedQueueConsumerWithOptions(client, ctx, queue, entryConsumer, failureHandler, nil)
}
func NewSortedQueueConsumerWithOptions(
	client *RedisClient,
	ctx context.Context,
	queue string,
	entryConsumer XSortedQueueEntryConsumerFunc,
	failureHandler XSortedQueueFailureHandlerFunc,
	options *XSortedQueueOptions,
) chan struct{} {
	if options == nil {
		options = NewXSortedQueueOptions()
	} else {
		options.Initialize()
	}

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(options.Consumers)
	for i := 0; i < options.Consumers; i++ {
		consumerId := uuid.New().String()
		go func() {
			log.Printf("running sorted queue consumer %s", consumerId)
			internalProcessMissingSortedEntries(client, ctx, consumerId, queue, failureHandler)
			for {
				internalConsumeSortedQueue(client, ctx, consumerId, queue, entryConsumer, failureHandler, *options)

				select {
				case <-ctx.Done():
					log.Printf("consumer %s is done", consumerId)
					wg.Done()
					return
				default:
					continue
				}
			}
		}()
	}
	// This go routine will make sure we close the channel once all the consumers
	// safely completed their work on shuting down
	go func() {
		wg.Wait()
		close(done)
	}()

	return done
}

func internalConsumeSortedQueue(
	client *RedisClient,
	shutdown context.Context,
	consumerId string,
	queue string,
	entryConsumer XSortedQueueEntryConsumerFunc,
	failureHandler XSortedQueueFailureHandlerFunc,
	options XSortedQueueOptions,
) {
	rawEntries, err := client.ZPopMin(shutdown, queue, options.Consuming).Result()
	if err != nil {
		if !strings.HasPrefix(err.Error(), "context canceled") {
			failureHandler([]XFailure{{Err: fmt.Errorf("failed to read queue: %v", err)}}, consumerId)
			// If we encounter an internal error we should retry with cautious
			// we don't wanna exchaust all the server's resources executing
			// the same failure in a loop
			time.Sleep(time.Minute)
		}
		return
	}

	if len(rawEntries) > 0 {
		queueFailures := []XFailure{}
		entries := []XSortedQueueEntry{}
		for _, e := range rawEntries {
			// Reads entry value by its reference URI via Redis.Get()
			entryValue, err := getXSortedQueueEntryByUri(client, e.Member)
			if err != nil {
				// Appends entry reference URI for failure report
				queueFailures = append(queueFailures, XFailure{Err: err, Payload: XGenericMap{"value": e.Member}})
				continue
			}
			// Parses sorted queue entry
			entry, err := parseXSortedQueueEntry(entryValue)
			if err != nil {
				// Appends entry value for failure report
				queueFailures = append(queueFailures, XFailure{Err: err, Payload: XGenericMap{"referenceUri": e.Member, "value": entryValue}})
				continue
			}
			// Updates the entry max retries field
			entry.maxRetries = options.MaxRetries
			// Set queue name
			entry.queue = queue
			// Set redis client for internal usage
			entry.setClient(client)
			if entry.HasExhaustedRetries() {
				// Exhausted retries should be passed to failure handler
				queueFailures = append(queueFailures, XFailure{Err: errors.New("retries exhausted"), Payload: XGenericMap{"entry": *entry}})
			} else {
				// Appends for processing
				entries = append(entries, *entry)
			}
		}
		if len(entries) > 0 {
			if err := markSortedEntryForProcessing(client, queue, entries...); err != nil {
				for _, e := range entries {
					queueFailures = append(queueFailures, XFailure{Err: err, Payload: XGenericMap{"entry": e}})
				}
			} else {
				queueFailures = append(queueFailures, handleXSortedQueueEntries(client, consumerId, queue, entryConsumer, entries, options)...)
			}
		}
		if len(queueFailures) > 0 {
			// Reports the failures to failure handler
			failureHandler(queueFailures, consumerId)
		}
	}

	// Either we should use BZPopMin which requires Redis 7.0
	// or this is required to slow down the CPU usage of the consumers
	time.Sleep(time.Millisecond * 100)
}

func handleXSortedQueueEntries(
	client *RedisClient,
	consumerId string,
	queue string,
	consumer XSortedQueueEntryConsumerFunc,
	entries []XSortedQueueEntry,
	options XSortedQueueOptions,
) (failures []XFailure) {
	defer func() {
		if r := recover(); r != nil {
			for _, entry := range entries {
				if ok, err := client.SIsMember(context.Background(), sortedQueueProcessingReferenceKey(queue), entry.ReferenceUri).Result(); err == nil && ok {
					entry.setFailure(fmt.Errorf("PANIC: %v", r))
					if err := retrySortedQueueEntry(client, entry); err != nil {
						failures = append(failures, XFailure{Err: err, Payload: XGenericMap{"entry": entry}})
					}
				}
			}
		}
	}()

	for _, e := range consumer(entries, consumerId) {
		// Checks if consumer has marked the entry with failure
		if e.currentFailure != nil {
			// We retry the failed entries by adding them back to the queue with in lower priority
			// and the Background context we provide here is not cancellable
			if err := retrySortedQueueEntry(client, e); err != nil {
				failures = append(failures, XFailure{Err: err, Payload: XGenericMap{"entry": e}})
				continue
			}
			continue
		}
		// Clean up the resource after successfully processed
		cleanSortedQueueEntry(client, e)
	}

	return failures
}

func retrySortedQueueEntry(client *RedisClient, entry XSortedQueueEntry) error {
	// Checks if retries have been exhausted or not
	if entry.IsLastRetry() {
		// Reports the entry to failure handler
		return errors.New("retries exhausted")
	}
	// Next line will attempt to lower the priority and enqueue the entry to be processed again
	// Note: Higher numbers have lower priority, we increase the entry's priority by 10%
	entry.Priority = entry.Priority * 1.1
	// We retry the failed entries by adding them back to the queue with in lower priority
	// and the Background context we provide here is not cancellable
	if err := enqueueSortedEntryWithPayload(client, context.Background(), entry.queue, entry, entry.CurrentRetries+1); err != nil {
		return fmt.Errorf("failed to retry: %v", err)
	}

	return nil
}

func internalProcessMissingSortedEntries(client *RedisClient, ctx context.Context, consumerId string, queue string, failureHandler XSortedQueueFailureHandlerFunc) {
	unlock := lock(client, queue)
	defer unlock()

	v, err := client.SMembers(ctx, sortedQueueProcessingReferenceKey(queue)).Result()
	if err != nil {
		log.Printf("failed: %v", err)
	}
	successRefs := []string{}
	failures := []XFailure{}
	for _, ref := range v {
		priorityStr, err := client.HGet(context.Background(), sortedQueueProcessingPriorityKey(queue), ref).Result()
		if err != nil {
			failures = append(failures, XFailure{Err: fmt.Errorf("failed to revive processing sorted queue: %v", err), Payload: XGenericMap{"referenceUri": ref}})
			continue
		}
		priority, err := strconv.ParseFloat(priorityStr, 64)
		if err != nil {
			failures = append(failures, XFailure{Err: fmt.Errorf("failed to decode sorted queue priority: %v", err), Payload: XGenericMap{"referenceUri": ref, "priorityStr": priorityStr}})
			continue
		}
		if err := enqueueSortedEntry(client, ctx, queue, ref, priority); err != nil {
			failures = append(failures, XFailure{Err: fmt.Errorf("failed to enqueue sorted queue entry: %v", err), Payload: XGenericMap{"referenceUri": ref, "priority": priority}})
			continue
		}
		successRefs = append(successRefs, ref)
		log.Printf("ref %s priority %v", ref, priorityStr)
	}

	// All the removing functions should rely on the cancel context
	// to avoid removing elements from the lists if the operation should be cancelled
	for _, ref := range successRefs {
		if err := client.HDel(ctx, sortedQueueProcessingPriorityKey(queue), ref).Err(); err != nil {
			failures = append(failures, XFailure{Err: fmt.Errorf("failed to clean up priority hash: %v", err), Payload: XGenericMap{"referenceUri": ref}})
			continue
		}
		if err := client.SRem(ctx, sortedQueueProcessingReferenceKey(queue), ref).Err(); err != nil {
			failures = append(failures, XFailure{Err: fmt.Errorf("failed to remove reference from processing set: %v", err), Payload: XGenericMap{"referenceUri": ref}})
		}
	}

	if len(failures) > 0 {
		failureHandler(failures, consumerId)
	}
}
